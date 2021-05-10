package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type AfterHook func(error)

type BeforeHook func(context.Context, QueryInfo) AfterHook

type QueryInfo struct {
	OperationName string
}

type customCommand struct {
	cmd    interface{}
	ignore bool
}

type ErrorHandler func(error)

type mongoConfig struct {
	appName               string
	uri                   string
	databaseName          string
	collectionsAndIndices map[string][]mongo.IndexModel // famous last words... Don't bother about thread safety
	username              string
	password              string
	customCommands        []*customCommand
	hooks                 []BeforeHook
	errorHandler          ErrorHandler
}

type mongoOptionsBuilder struct {
	option func(*mongoConfig)
	next   *mongoOptionsBuilder
}

func Builder() *mongoOptionsBuilder {
	return new(mongoOptionsBuilder)
}

func (ob *mongoOptionsBuilder) BuildAndConnect() (MongoWrapper, error) {
	pointer := ob
	conf := &mongoConfig{}
	for pointer.next != nil {
		pointer.option(conf)
		pointer = pointer.next
	}
	if conf.errorHandler == nil {
		conf.errorHandler = func(_ error) {}
	}
	return create(conf)
}

func (ob *mongoOptionsBuilder) AppName(name string) *mongoOptionsBuilder {
	return &mongoOptionsBuilder{
		option: func(cfg *mongoConfig) {
			cfg.appName = name
		},
		next: ob,
	}
}

func (ob *mongoOptionsBuilder) URI(uri string) *mongoOptionsBuilder {
	return &mongoOptionsBuilder{
		option: func(cfg *mongoConfig) {
			cfg.uri = uri
		},
		next: ob,
	}
}

func (ob *mongoOptionsBuilder) DatabaseName(name string) *mongoOptionsBuilder {
	return &mongoOptionsBuilder{
		option: func(cfg *mongoConfig) {
			cfg.databaseName = name
		},
		next: ob,
	}
}

func (ob *mongoOptionsBuilder) UsernamePassword(username, password string) *mongoOptionsBuilder {
	return &mongoOptionsBuilder{
		option: func(cfg *mongoConfig) {
			cfg.username = username
			cfg.password = password
		},
		next: ob,
	}
}

func (ob *mongoOptionsBuilder) AddCustomCommandOption(command interface{}, ignoreError bool) *mongoOptionsBuilder {
	return &mongoOptionsBuilder{
		option: func(cfg *mongoConfig) {
			cfg.customCommands = append(cfg.customCommands, &customCommand{command, ignoreError})
		},
		next: ob,
	}
}

func (ob *mongoOptionsBuilder) AddIndex2Collection(collectionName string, indexModel mongo.IndexModel) *mongoOptionsBuilder {
	return &mongoOptionsBuilder{
		option: func(cfg *mongoConfig) {
			if len(cfg.collectionsAndIndices) == 0 {
				cfg.collectionsAndIndices = make(map[string][]mongo.IndexModel)
			}
			v := cfg.collectionsAndIndices[collectionName]
			cfg.collectionsAndIndices[collectionName] = append(v, indexModel)
		},
		next: ob,
	}
}

func (ob *mongoOptionsBuilder) AddHook(hook BeforeHook) *mongoOptionsBuilder {
	return &mongoOptionsBuilder{
		option: func(cfg *mongoConfig) {
			cfg.hooks = append(cfg.hooks, hook)
		},
		next: ob,
	}
}

func (ob *mongoOptionsBuilder) ErrorHandler(handler ErrorHandler) *mongoOptionsBuilder {
	return &mongoOptionsBuilder{
		option: func(cfg *mongoConfig) {
			cfg.errorHandler = handler
		},
		next: ob,
	}
}
