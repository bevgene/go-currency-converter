package tests

import (
	"context"
	"encoding/json"
	"github.com/bevgene/go-currency-rate/app/clients"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/bevgene/go-currency-rate/app/mortar"
	"github.com/go-masonry/mortar/interfaces/log"
	"github.com/go-masonry/mortar/providers"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"testing"
	"time"
)

type (
	mongoClientTestSuiteDeps struct {
		fx.In

		MongoClient *clients.LazyMongoClient
		Ctx         context.Context
		Logger      log.Logger
		Rates       *model.ExchangeRatesModel
	}

	mongoClientTestSuite struct {
		suite.Suite

		TestApp *fxtest.App
		deps    mongoClientTestSuiteDeps
	}
)

const (
	jsonInput = "{\"success\":true,\"date\":\"2021-05-13\",\"base\":\"EUR\",\"timestamp\":1620891063,\"rates\":{\"AED\":4.444896,\"AFN\":95.138176,\"ALL\":123.06269,\"AMD\":634.0723,\"ANG\":2.179995,\"AOA\":791.7302,\"ARS\":113.74339,\"AUD\":1.566062,\"AWG\":2.178767,\"AZN\":2.055924,\"BAM\":1.958881,\"BBD\":2.452139,\"BDT\":102.9815,\"BGN\":1.963605,\"BHD\":0.456266,\"BIF\":2398.939,\"BMD\":1.21009,\"BND\":1.613526,\"BOB\":8.385961,\"BRL\":6.418804,\"BSD\":1.214502,\"BTC\":0.0000237046,\"BTN\":89.2038,\"BWP\":13.009963,\"BYN\":3.077519,\"BYR\":23717.758,\"BZD\":2.448033,\"CAD\":1.468414,\"CDF\":2420.1794,\"CHF\":1.097388,\"CLF\":0.031023,\"CLP\":856.0107,\"CNY\":7.806655,\"COP\":4538.4536,\"CRC\":746.77313,\"CUC\":1.21009,\"CUP\":32.06738,\"CVE\":110.437004,\"CZK\":25.578636,\"DJF\":216.20853,\"DKK\":7.436037,\"DOP\":69.06757,\"DZD\":161.32309,\"EGP\":18.975662,\"ERN\":18.153753,\"ETB\":51.854965,\"EUR\":1,\"FJD\":2.466102,\"FKP\":0.86107,\"GBP\":0.86067,\"GEL\":4.156653,\"GGP\":0.86107,\"GHS\":7.001505,\"GIP\":0.86107,\"GMD\":62.01748,\"GNF\":11974.751,\"GTQ\":9.367212,\"GYD\":253.59497,\"HKD\":9.399088,\"HNL\":29.17444,\"HRK\":7.527121,\"HTG\":106.99653,\"HUF\":357.29712,\"IDR\":17315.355,\"ILS\":3.979453,\"IMP\":0.86107,\"INR\":89.02425,\"IQD\":1771.9052,\"IRR\":50950.83,\"ISK\":150.49892,\"JEP\":0.86107,\"JMD\":183.629,\"JOD\":0.857999,\"JPY\":132.6585,\"KES\":129.53986,\"KGS\":102.488914,\"KHR\":4939.334,\"KMF\":492.6878,\"KPW\":1089.0802,\"KRW\":1367.2932,\"KWD\":0.364235,\"KYD\":1.012076,\"KZT\":518.23145,\"LAK\":11446.522,\"LBP\":1831.3676,\"LKR\":238.64667,\"LRD\":208.0751,\"LSL\":16.941523,\"LTL\":3.57308,\"LVL\":0.731971,\"LYD\":5.421389,\"MAD\":10.757255,\"MDL\":21.496458,\"MGA\":4567.023,\"MKD\":61.69609,\"MMK\":1891.5631,\"MNT\":3449.8606,\"MOP\":9.71465,\"MRO\":432.00183,\"MUR\":49.274227,\"MVR\":18.642546,\"MWK\":968.54553,\"MXN\":24.373594,\"MYR\":4.992227,\"MZN\":70.923225,\"NAD\":16.941055,\"NGN\":496.41522,\"NIO\":42.416416,\"NOK\":10.094738,\"NPR\":142.7273,\"NZD\":1.688226,\"OMR\":0.465921,\"PAB\":1.214492,\"PEN\":4.502713,\"PGK\":4.318927,\"PHP\":57.946358,\"PKR\":184.64452,\"PLN\":4.549877,\"PYG\":8121.316,\"QAR\":4.405945,\"RON\":4.927604,\"RSD\":117.76342,\"RUB\":89.97066,\"RWF\":1215.8638,\"SAR\":4.538699,\"SBD\":9.661434,\"SCR\":18.66583,\"SDG\":493.71628,\"SEK\":10.182614,\"SGD\":1.613408,\"SHP\":0.86107,\"SLL\":12385.268,\"SOS\":707.90295,\"SRD\":17.127632,\"STD\":25092.16,\"SVC\":10.626551,\"SYP\":1521.557,\"SZL\":17.00102,\"THB\":37.91238,\"TJS\":13.850959,\"TMT\":4.235314,\"TND\":3.308994,\"TOP\":2.731291,\"TRY\":10.231911,\"TTD\":8.258531,\"TWD\":33.837784,\"TZS\":2806.1982,\"UAH\":33.583206,\"UGX\":4300.1978,\"USD\":1.21009,\"UYU\":53.4192,\"UZS\":12803.8125,\"VEF\":258753760000,\"VND\":27892.568,\"VUV\":131.18246,\"WST\":3.044525,\"XAF\":656.9802,\"XAG\":0.044738,\"XAU\":0.000665,\"XCD\":3.270327,\"XDR\":0.842212,\"XOF\":656.9857,\"XPF\":119.97884,\"YER\":302.5229,\"ZAR\":17.038431,\"ZMK\":10892.264,\"ZMW\":27.180258,\"ZWL\":389.64893}}"
)

func TestMongoClient(t *testing.T) {
	suite.Run(t, new(mongoClientTestSuite))
}

func (impl *mongoClientTestSuite) SetupTest() {
	testApp := fxtest.New(
		impl.T(),
		fx.Supply(impl.T()),
		mortar.ViperFxOption("../config/config.yml", "../config/config_test.yml"),
		mortar.LoggerFxOption(),
		mortar.HttpServerFxOptions(),
		mortar.HttpClientFxOptions(),
		mortar.InternalHttpHandlersFxOptions(),
		mortar.ServiceAPIsAndOtherDependenciesFxOption(),
		fx.Provide(
			clients.CreateMongoClient,
			createRatesModel,
			func() context.Context { return context.Background() },
		),
		providers.BuildMortarWebServiceFxOption(),
		fx.Populate(&impl.deps),
	)
	impl.TestApp = testApp
	impl.TestApp.RequireStart()
}

func createRatesModel() (result *model.ExchangeRatesModel, err error) {
	var ratesModel model.ExchangeRatesModel
	if err = json.Unmarshal([]byte(jsonInput), &ratesModel); err != nil {
		return
	}
	result = &ratesModel
	return
}

func (impl *mongoClientTestSuite) TearDownTest() {
	if impl.TestApp != nil {
		impl.TestApp.RequireStop()
	}
}

func (impl *mongoClientTestSuite) TestAdd() {
	t := impl.T()

	params := gopter.DefaultTestParameters()
	props := gopter.NewProperties(params)
	props.Property("happy convert test", impl.happyInsert(t))
	props.TestingRun(t)
}

func (impl *mongoClientTestSuite) happyInsert(t *testing.T) gopter.Prop {
	return prop.ForAll(
		func(createdAt time.Time) bool {
			docPtr := model.ConvertExchangeRatesModel(*impl.deps.Rates)
			docPtr.CreatedAt = createdAt
			err := impl.deps.MongoClient.Client.AddRateDocument(context.Background(), docPtr)
			return assert.NoError(t, err, "failed to insert document")
		},
		gen.TimeRange(time.Now().UTC().Add(-24*time.Hour), 24*time.Hour),
	)
}
