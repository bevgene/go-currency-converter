package clients

import (
	"context"
	"fmt"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/log"
	"github.com/go-masonry/mortar/interfaces/monitor"
	"github.com/go-masonry/mortar/interfaces/trace"
	"github.com/uber-go/tally"
	promreporter "github.com/uber-go/tally/prometheus"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"time"
)

//go:generate mockgen -source=temporal_client.go -destination=mock/temporal_client_mock.go

type (
	temporalClientDeps struct {
		fx.In

		Logger    log.Logger
		Config    cfg.Config
		Tracer    trace.OpenTracer `optional:"true"`
		Metrics   monitor.Metrics  `optional:"true"`
		Lifecycle fx.Lifecycle
	}

	LazyClient struct {
		client.Client
	}

	temporalLogger struct {
		logger log.Logger
	}
)

const (
	hostPortKye     = "exchangerate.temporal.hostPort"
	namespaceKey    = "exchangerate.temporal.namespace"
	workflowNameKey = "exchangerate.temporal.workflowName"
)

func CreateTemporalClient(deps temporalClientDeps) *LazyClient {
	hostPort := deps.Config.Get(hostPortKye).String()
	namespace := deps.Config.Get(namespaceKey).String()
	//workflowName := deps.Config.Get(workflowNameKey).String()

	options := client.Options{
		HostPort:     hostPort,
		Namespace:    namespace,
		Logger:       &temporalLogger{deps.Logger},
		MetricsScope: deps.tally(),
	}

	if deps.Tracer != nil {
		options.Tracer = deps.Tracer.Tracer()
	}

	var clientPtr = new(LazyClient)
	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			deps.Logger.Info(ctx, "Starting Temporal Client...")
			var newClient client.Client
			if newClient, err = client.NewClient(options); err != nil {
				return
			}
			clientPtr.Client = newClient
			return
		},

		OnStop: func(ctx context.Context) error {
			if clientPtr.Client != nil {
				clientPtr.Client.Close()
			}
			return nil
		},
	})

	return clientPtr
}

func (impl *temporalLogger) Debug(msg string, keyvals ...interface{}) {
	impl.mapKeyValues(log.DebugLevel, msg, keyvals...)
}

func (impl *temporalLogger) Info(msg string, keyvals ...interface{}) {
	impl.mapKeyValues(log.InfoLevel, msg, keyvals...)
}

func (impl *temporalLogger) Warn(msg string, keyvals ...interface{}) {
	impl.mapKeyValues(log.WarnLevel, msg, keyvals...)
}

func (impl *temporalLogger) Error(msg string, keyvals ...interface{}) {
	impl.mapKeyValues(log.ErrorLevel, msg, keyvals...)
}

func (impl *temporalLogger) mapKeyValues(level log.Level, msg string, keyvals ...interface{}) {
	withField := impl.logger.WithField("temporal_log", true)
	if (len(keyvals) % 2) == 0 {
		for i := 0; i < len(keyvals); i += 2 {
			switch key := keyvals[i].(type) {
			case string:
				withField = withField.WithField(fmt.Sprintf("temporal_key_%s", key), keyvals[i+1])
			}
		}
	}
	withField.Custom(nil, level, 2, msg, keyvals...)
}

func (impl *temporalClientDeps) tally() tally.Scope {
	if impl.Metrics == nil {
		return nil
	}
	// Prometheus doesn't like metrics names with "." or "-" in them.
	options := tally.ScopeOptions{
		Prefix:         "",
		Tags:           nil,
		CachedReporter: createCachedReporter(impl.Metrics),
		Separator:      promreporter.DefaultSeparator,
	}
	scope, closer := tally.NewRootScope(options, time.Second)
	impl.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return closer.Close()
		},
	})
	return scope
}
