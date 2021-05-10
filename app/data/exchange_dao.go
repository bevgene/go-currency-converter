package data

import (
	"context"
	"fmt"
	"github.com/bevgene/go-currency-rate/app/data/mongo"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.uber.org/fx"
)

type (
	CurrencyRateDao interface {
		UpdateRates(context.Context, model.ExchangeRateDocument) error
		GetRates(ctx context.Context) (model.ExchangeRateDocument, error)
	}

	currencyRateDaoImplDeps struct {
		fx.In

		Logger    log.Logger
		Config    cfg.Config
		Lifecycle fx.Lifecycle
	}

	currencyRateDaoImpl struct {
		deps       currencyRateDaoImplDeps
		collection mongo.CollectionWrapper
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

func CreateCurrencyRateDao(deps currencyRateDaoImplDeps) (result CurrencyRateDao, err error) {
	appName := deps.Config.Get(appNameKey).String()
	dbName := deps.Config.Get(databaseKey).String()
	host := deps.Config.Get(hostKey).String()
	port := deps.Config.Get(portKey).String()
	userName := deps.Config.Get(userKey).String()
	password := deps.Config.Get(passwordKey).String()
	collectionName := deps.Config.Get(collectionKey).String()
	uri := fmt.Sprintf("mongodb://%s:%s/%s", host, port, dbName)
	if len(userName) > 0 && len(password) > 0 {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", userName, password, host, port, dbName)
	}
	var wrapper mongo.MongoWrapper
	if wrapper, err = mongo.Builder().
		AppName(appName).
		URI(uri).
		DatabaseName(dbName).
		AddIndex2Collection(
			collectionName,
			mongo.NewIndexBuilder().
				Keys().
				Asc("created_at").
				Options().
				Unique(true).
				Done().
				Build()).
		BuildAndConnect(); err != nil {
		return
	}
	var collection mongo.CollectionWrapper
	if collection, err = wrapper.Collection(collectionName); err != nil {
		return
	}
	result = &currencyRateDaoImpl{
		deps:       deps,
		collection: collection,
	}
	return
}

func (impl *currencyRateDaoImpl) UpdateRates(ctx context.Context, ratesModel model.ExchangeRateDocument) error {
	panic("implement me")
}

func (impl *currencyRateDaoImpl) GetRates(ctx context.Context) (model.ExchangeRateDocument, error) {
	panic("implement me")
}
