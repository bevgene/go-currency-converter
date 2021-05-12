package mortar

import (
	"github.com/bevgene/go-currency-rate/app/clients"
	"go.uber.org/fx"
)

func MongoFxOptions() fx.Option {
	return fx.Options(
		fx.Provide(clients.CreateMongoClient),
	)
}
