package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type dbWrapper struct {
	client      *mongo.Client
	db          *mongo.Database
	collections map[string]*mongo.Collection
	config      *mongoConfig
}

func create(conf *mongoConfig) (wrapper MongoWrapper, err error) {
	var client *mongo.Client
	var clientOptions = options.Client()
	if len(conf.databaseName) == 0 || len(conf.collectionsAndIndices) == 0 {
		err = fmt.Errorf("no db/collection specified, This incident will be reported") // https://xkcd.com/838
		return
	}
	if len(conf.uri) == 0 { // fallback to localhost
		conf.uri = "mongodb://localhost:27017" // default
	}
	clientOptions = clientOptions.ApplyURI(conf.uri) // Set URI
	if len(conf.appName) > 0 {
		clientOptions = clientOptions.SetAppName(conf.appName)
	}
	if len(conf.username) > 0 && len(conf.password) > 0 {
		auth := options.Credential{
			AuthMechanism: "SCRAM-SHA-1", // Let's hope this one will not change
			Password:      conf.password,
			Username:      conf.username,
		}
		clientOptions = clientOptions.SetAuth(auth)
	}
	// if you found this during debug, then please add an option to set custom timeout
	// Unless you found this during Index creation on Initialization, in that case make your index a background one
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if client, err = mongo.Connect(ctx, clientOptions); err == nil {
		database := client.Database(conf.databaseName, options.Database()) // you can set read/write preferences but unless we have a real use case. it's unsupported
		for _, customCmd := range conf.customCommands {
			result := database.RunCommand(ctx, customCmd.cmd)
			if result.Err() != nil && !customCmd.ignore {
				err = fmt.Errorf("custom command failed %s", result.Err().Error())
				return
			}
		}
		wrapperImpl := &dbWrapper{
			client:      client,
			db:          database,
			collections: make(map[string]*mongo.Collection),
			config:      conf,
		}
		for name, indices := range conf.collectionsAndIndices {
			collection := database.Collection(name) // you can set read/write preferences but unless we have a real use case. it's unsupported
			for _, index := range indices {
				if doc, ok := index.Keys.(bsonx.Doc); ok && len(doc) > 0 {
					if _, err = collection.Indexes().CreateOne(ctx, index); err != nil {
						return
					}
				}
			}
			wrapperImpl.collections[name] = collection
		}
		wrapper = wrapperImpl
	}
	return
}

func (dw *dbWrapper) Collection(name string) (CollectionWrapper, error) {
	if coll, ok := dw.collections[name]; ok {
		return &collectionWrapper{
			c:            coll,
			hooks:        dw.config.hooks,
			errorHandler: dw.config.errorHandler,
		}, nil
	}
	return nil, fmt.Errorf("unknown collection: %s", name)
}

func (dw *dbWrapper) InternalDatabase() *mongo.Database {
	return dw.db
}

func (dw *dbWrapper) InternalClient() *mongo.Client {
	return dw.client
}

func (dw *dbWrapper) Ping() error {
	return dw.client.Ping(nil, nil)
}

func (dw *dbWrapper) Disconnect(ctx context.Context) error {
	return dw.client.Disconnect(ctx)
}
