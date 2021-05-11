package clients

import (
	"context"
	"fmt"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
	"net"
)

//go:generate mockgen -source=mongo_client.go -destination=mock/mongo_client_mock.go

type (
	mongoClientImplDeps struct {
		fx.In

		Logger    log.Logger
		Config    cfg.Config
		Lifecycle fx.Lifecycle
	}

	MongoClient interface {
		AddRateDocument(context.Context, *model.ExchangeRateDocument) error
		GetLatestRateDocument(context.Context) (*model.ExchangeRateDocument, error)
	}

	mongoClientImpl struct {
		deps       mongoClientImplDeps
		client     *mongo.Client
		collection *mongo.Collection
	}
)

const (
	appNameKey    = "mortar.name"
	hostKey       = "exchangerate.database.host"
	portKey       = "exchangerate.database.port"
	userKey       = "exchangerate.database.user"
	passwordKey   = "exchangerate.database.password"
	databaseKey   = "exchangerate.database.name"
	collectionKey = "exchangerate.database.collection"
)

func CreateMongoClient(deps mongoClientImplDeps) (result MongoClient, err error) {
	appName := deps.Config.Get(appNameKey).String()
	dbName := deps.Config.Get(databaseKey).String()
	host := deps.Config.Get(hostKey).String()
	port := deps.Config.Get(portKey).String()
	userName := deps.Config.Get(userKey).String()
	password := deps.Config.Get(passwordKey).String()
	collectionName := deps.Config.Get(collectionKey).String()

	uri := fmt.Sprintf("mongodb://%s/%s", net.JoinHostPort(host, port), dbName)
	if len(userName) > 0 && len(password) > 0 {
		uri = fmt.Sprintf("mongodb://%s:%s@%s/%s", userName, password, net.JoinHostPort(host, port), dbName)
	}
	clientOptions := options.Client().ApplyURI(uri).SetAppName(appName)
	var mongoClient *mongo.Client
	var collection *mongo.Collection
	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) (startError error) {
			if mongoClient, startError = mongo.Connect(ctx, clientOptions); startError != nil {
				deps.Logger.WithError(startError).Error(ctx, "failed to create mongo client")
				return
			}
			if startError = mongoClient.Ping(ctx, nil); startError != nil {
				deps.Logger.WithError(startError).Error(ctx, "failed to ping mongo db")
			}
			collection = mongoClient.Database(dbName).Collection(collectionName)
			indexModel := mongo.IndexModel{
				Keys:    bson.D{{"created_at", 1}},
				Options: options.Index().SetUnique(true),
			}

			_, startError = collection.Indexes().CreateOne(ctx, indexModel)
			return
		},
		OnStop: func(ctx context.Context) (stopError error) {
			if mongoClient != nil {
				if stopError = mongoClient.Disconnect(ctx); stopError != nil {
					deps.Logger.WithError(stopError).Error(ctx, "failed to disconnect from mongo db")
				}
			}
			return
		},
	})
	result = &mongoClientImpl{
		deps:       deps,
		client:     mongoClient,
		collection: collection,
	}
	return
}

func (impl *mongoClientImpl) AddRateDocument(ctx context.Context, document *model.ExchangeRateDocument) (err error) {
	_, err = impl.collection.InsertOne(ctx, document)
	return
}

func (impl *mongoClientImpl) GetLatestRateDocument(ctx context.Context) (result *model.ExchangeRateDocument, err error) {
	findOneOptions := options.FindOne()
	findOneOptions.SetSort(bson.M{"created_at": -1})
	var doc model.ExchangeRateDocument
	if err = impl.collection.FindOne(ctx, bson.M{}, findOneOptions).Decode(&doc); err != nil {
		impl.deps.Logger.WithError(err).Error(ctx, "failed decoding result")
		return
	}
	result = &doc
	return
}
