package mongo

import (
	"context"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/multierr"
)

type collectionWrapper struct {
	c            *mongo.Collection
	hooks        []BeforeHook
	errorHandler ErrorHandler
}

func (impl *collectionWrapper) InsertOne(ctx context.Context, doc interface{}) (id interface{}, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "InsertOne"})
	defer func() {
		after(err)
	}()
	var result *mongo.InsertOneResult
	if result, err = impl.c.InsertOne(ctx, doc); err == nil {
		id = result.InsertedID
	}
	return
}

func (impl *collectionWrapper) InsertUnordered(ctx context.Context, docs []interface{}) (ids []interface{}, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "InsertUnordered"})
	defer func() {
		after(err)
	}()
	return impl.insertMany(ctx, false, docs)
}

func (impl *collectionWrapper) InsertOrdered(ctx context.Context, docs []interface{}) (ids []interface{}, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "InsertOrdered"})
	defer func() {
		after(err)
	}()
	return impl.insertMany(ctx, true, docs)
}

func (impl *collectionWrapper) insertMany(ctx context.Context, ordered bool, docs []interface{}) (ids []interface{}, err error) {
	var result *mongo.InsertManyResult

	if result, err = impl.c.InsertMany(ctx, docs, options.InsertMany().SetOrdered(ordered)); err == nil {
		ids = result.InsertedIDs
	}
	return
}

func (impl *collectionWrapper) FindOne(ctx context.Context, query interface{}, result interface{}) (err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "FindOne"})
	defer func() {
		after(err)
	}()
	err = impl.c.FindOne(ctx, query).Decode(result)
	return
}

func (impl *collectionWrapper) Find(ctx context.Context, query, sort, hint interface{}, limit, offset int, result interface{}) (err error) {
	// No hook here since we are calling FindCursor
	var cursor *mongo.Cursor
	if cursor, err = impl.FindCursor(ctx, query, sort, hint, limit, offset); err == nil {
		err = impl.consumeCursor(ctx, cursor, result)
	}
	return
}

func (impl *collectionWrapper) FindCursor(ctx context.Context, query, sort, hint interface{}, limit, offset int) (cursor *mongo.Cursor, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "Find"})
	defer func() {
		after(err)
	}()
	findOptions := options.Find().SetSort(sort).SetLimit(int64(limit)).SetSkip(int64(offset))
	if hint != nil {
		findOptions = findOptions.SetHint(hint)
	}
	return impl.c.Find(ctx, query, findOptions)
}

func (impl *collectionWrapper) FindAndModify(ctx context.Context, query, update, sort interface{}, returnUpdated, upsert bool, result interface{}) (err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "FindAndModify"})
	defer func() {
		after(err)
	}()
	findAndModifyOptions := options.FindOneAndUpdate().SetSort(sort).SetUpsert(upsert)
	if returnUpdated {
		findAndModifyOptions = findAndModifyOptions.SetReturnDocument(options.After)
	}
	return impl.c.FindOneAndUpdate(ctx, query, update, findAndModifyOptions).Decode(result)
}

func (impl *collectionWrapper) Count(ctx context.Context, query interface{}) (result int64, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "Count"})
	defer func() {
		after(err)
	}()
	return impl.c.CountDocuments(ctx, query)
}

func (impl *collectionWrapper) UpdateAll(ctx context.Context, query, update interface{}) (matched, modified int64, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "UpdateAll"})
	defer func() {
		after(err)
	}()
	var result *mongo.UpdateResult
	if result, err = impl.c.UpdateMany(ctx, query, update); err == nil {
		matched = result.MatchedCount
		modified = result.ModifiedCount
	}
	return
}

func (impl *collectionWrapper) UpdateOne(ctx context.Context, query, update interface{}, upsert bool) (result interface{}, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "UpdateOne"})
	defer func() {
		after(err)
	}()
	return impl.c.UpdateOne(ctx, query, update, options.Update().SetUpsert(upsert))
}

func (impl *collectionWrapper) RemoveOne(ctx context.Context, query interface{}) (removedCount int64, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "RemoveOne"})
	defer func() {
		after(err)
	}()
	var result *mongo.DeleteResult
	if result, err = impl.c.DeleteOne(ctx, query); err == nil {
		removedCount = result.DeletedCount
	}
	return
}

func (impl *collectionWrapper) RemoveMany(ctx context.Context, query interface{}) (removedCount int64, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "RemoveMany"})
	defer func() {
		after(err)
	}()
	var result *mongo.DeleteResult
	if result, err = impl.c.DeleteMany(ctx, query); err == nil {
		removedCount = result.DeletedCount
	}
	return
}

func (impl *collectionWrapper) Aggregate(ctx context.Context, stages []interface{}, result interface{}, opts ...*options.AggregateOptions) (err error) {
	// No hook here since we are calling AggregateCursor
	var cursor *mongo.Cursor
	if cursor, err = impl.AggregateCursor(ctx, stages, opts...); err == nil {
		err = impl.consumeCursor(ctx, cursor, result)
	}
	return
}

func (impl *collectionWrapper) AggregateCursor(ctx context.Context, stages []interface{}, opts ...*options.AggregateOptions) (cursor *mongo.Cursor, err error) {
	after := impl.before(ctx, QueryInfo{OperationName: "Aggregate"})
	defer func() {
		after(err)
	}()
	var pipeline []interface{}
	for _, stage := range stages {
		pipeline = append(pipeline, stage)
	}
	cursor, err = impl.c.Aggregate(ctx, pipeline, opts...)
	return
}

func (impl *collectionWrapper) InternalCollection() *mongo.Collection {
	return impl.c
}

func (impl *collectionWrapper) consumeCursor(ctx context.Context, cursor *mongo.Cursor, result interface{}) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	defer func() {
		err = multierr.Append(err, cursor.Close(ctx)) // we don't want to loose an error here
	}()
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result argument must be a slice address")
	}
	sliceElement := resultValue.Elem()
	sliceValue := reflect.MakeSlice(sliceElement.Type(), 0, 16) // Create slice, numbers taken out of the blue if you find it wrong please fix it
	for cursor.Next(ctx) {
		value := reflect.New(sliceValue.Type().Elem()) // Create a new placeholder for our value
		if err = cursor.Decode(value.Interface()); err != nil {
			return // Fail fast if you have an error
		}
		sliceValue = reflect.Append(sliceValue, value.Elem()) // Add decoded object to a slice
	}
	err = cursor.Err()
	sliceElement.Set(sliceValue)
	return
}

func (impl *collectionWrapper) before(ctx context.Context, info QueryInfo) func(err error) {
	defer impl.panicHandler()
	if ctx == nil {
		ctx = context.Background()
	}
	var afterEffects []AfterHook
	for _, before := range impl.hooks {
		if before != nil {
			afterEffects = append(afterEffects, before(ctx, info))
		}
	}
	return func(err error) {
		for _, after := range afterEffects {
			go func(e error, hook AfterHook) {
				defer impl.panicHandler()
				hook(e)
			}(err, after)
		}
	}
}

func (impl *collectionWrapper) handleError(err error) {
	if impl.errorHandler != nil {
		impl.errorHandler(err)
	}
}

func (impl *collectionWrapper) panicHandler() {
	if r := recover(); r != nil {
		switch t := r.(type) {
		case error:
			impl.handleError(t)
		case string:
			impl.handleError(fmt.Errorf(t))
		default:
			impl.handleError(fmt.Errorf("hook paniced"))
		}
	}
}
