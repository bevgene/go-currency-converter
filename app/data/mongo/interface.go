package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoWrapper interface {
	Collection(name string) (CollectionWrapper, error)
	InternalDatabase() *mongo.Database
	InternalClient() *mongo.Client
	Ping() error
	Disconnect(ctx context.Context) error
}

type CollectionWrapper interface {
	InsertOne(ctx context.Context, doc interface{}) (id interface{}, err error)
	InsertOrdered(ctx context.Context, docs []interface{}) (ids []interface{}, err error)
	InsertUnordered(ctx context.Context, docs []interface{}) (ids []interface{}, err error)
	FindOne(ctx context.Context, query interface{}, result interface{}) (err error)
	/*
		WARNING: Obviously, Find must not be used with result sets that may be
		 potentially large, since it may consume all memory until the system
		 crashes. Consider building the query with a Limit clause to ensure the
		 result size is bounded.
	*/
	Find(ctx context.Context, query, sort, hint interface{}, limit, offset int, result interface{}) (err error)
	//	Please make sure you close the cursor when you done
	//	For instance:
	//
	//	    if cursor, err := wrapper.FindCursor(ctx, query, sort, limit, offset); err == nil {
	//		      defer cursor.Close(ctx)
	//		      ...
	//		   }
	//
	FindCursor(ctx context.Context, query, sort, hint interface{}, limit, offset int) (cursor *mongo.Cursor, err error)
	FindAndModify(ctx context.Context, query, update, sort interface{}, returnUpdated, upsert bool, result interface{}) (err error)
	// UpdateAll doesn't expose upsert option since it will return only the first ID of the upserted document
	// However if you do find this useful please change this method accordingly
	//     https://docs.mongodb.com/manual/reference/method/db.collection.updateMany/#db.collection.updateMany
	UpdateAll(ctx context.Context, query, update interface{}) (matched, modified int64, err error)
	Count(ctx context.Context, query interface{}) (result int64, err error)
	// If upsert was performed then this function will return a upsertedId holding the ID of this new document
	UpdateOne(ctx context.Context, query, update interface{}, upsert bool) (upsertedId interface{}, err error)
	RemoveOne(ctx context.Context, query interface{}) (removedCount int64, err error)
	RemoveMany(ctx context.Context, query interface{}) (removedCount int64, err error)
	/*
		WARNING: Obviously, Aggregate must not be used with result sets that may be
		 potentially large, since it may consume all memory until the system
		 crashes. Consider building the pipeline with a Limit clause to ensure the
		 result size is bounded.
		NOTE: It's possible to use Disk as a storage, but that is left to the caller to decide

		Example:
				var docs []Document
				matchStage := M{
					"$match": M{
						"age": M{
							"$gte": 0,
						},
					},
				}
				sortStage := M{
					"$sort": M{
						"age": 1,
					},
				}
				err := wrapper.Aggregate(ctx, []interface{}{matchStage,sortStage}, &docs)
	*/
	Aggregate(ctx context.Context, stages []interface{}, result interface{}, opts ...*options.AggregateOptions) (err error)
	//	Please make sure you close the cursor when you done
	//	For instance:
	//
	//		matchStage := M{
	//			"$match": M{
	//				"age": M{
	//					"$gte": 0,
	//				},
	//			},
	//		}
	//		sortStage := M{
	//			"$sort": M{
	//				"age": 1,
	//			},
	//		}
	//
	//		if cursor, err := wrapper.AggregateCursor(ctx, []interface{}{matchStage, sortStage}); err == nil {
	//		      defer cursor.Close(ctx)
	//		      ...
	//		}
	//
	AggregateCursor(ctx context.Context, stages []interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error)
	InternalCollection() *mongo.Collection
}
