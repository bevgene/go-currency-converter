package mongo

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type Document struct {
	ID    primitive.ObjectID `bson:"_id"`
	Name  string             `bson:"name"`
	Age   int                `bson:"age"`
	Date  time.Time          `bson:"date"`
	Child *Sub               `bson:"child"`
}

type Sub struct {
	Numbers []int `bson:"numbers"`
	Bool    bool  `bson:"bool"`
}

type suite struct {
	wrapper  CollectionWrapper
	inserted int
}

func TestDBHappy(t *testing.T) {
	wrapper, err := Builder().AppName("mongo-wrapper-test").URI("mongodb://localhost:27017").DatabaseName("official-driver").AddCustomCommandOption(bson.M{
		"create": "cappedCol",
		"capped": true,
		"size":   64 * 1024 * 2,
	}, true).
		AddIndex2Collection("cappedCol", NewIndexBuilder().Build()). // No index
		AddIndex2Collection("anotherCol", NewIndexBuilder().
			Keys().Asc("one").Desc("two").
			Options().Unique(true).Done().
			Build(),
		).AddIndex2Collection("anotherCol", NewIndexBuilder().
		Keys().Text("desc").
		Options().Done().
		Build(),
	).AddIndex2Collection("anotherCol", NewIndexBuilder().
		Keys().Sphere("loc").
		Options().SphereVersion(3).Done().
		Build(),
	).BuildAndConnect()

	defer func() {
		err := wrapper.InternalClient().Disconnect(context.Background())
		assert.NoError(t, err)
	}()
	require.NoError(t, err, "Error initializing db")
	cappedCollection, err := wrapper.Collection("cappedCol")
	require.NoError(t, err, "There should be a [cappedCol] collection")
	_, err = cappedCollection.Count(nil, bsonx.Doc{})
	require.NoError(t, err, "Capped collection 'count' failed")
	var col CollectionWrapper
	col, err = wrapper.Collection("anotherCol")
	require.NoError(t, err, "There should be an [anotherCol] collection")
	var list *mongo.Cursor
	list, err = col.InternalCollection().Indexes().List(context.Background())
	require.NoError(t, err, "Error fetching indexes of [anotherCol]")
	defer func() {
		err := list.Close(context.Background())
		assert.NoError(t, err)
	}()
	for list.Next(context.Background()) {
		bytes := list.Current
		require.NotEmpty(t, bytes, "Error decoding index")
	}
	database := wrapper.InternalDatabase()
	require.Equal(t, "official-driver", database.Name())
	err = wrapper.Ping()
	require.NoError(t, err, "Ping error")
}

func TestDBNotSoHappy(t *testing.T) {
	t.Run("Bad command", badCommand)
	t.Run("Bad DB", badDB)
	t.Run("Bad Collection", badCollection)
	t.Run("Bad Auth", badAuth)
	t.Run("Bad Flow", badFlow)
	t.Run("Bad Hook", badHook)
	t.Run("Bad DB operation", errorDuringDBOperation)
}

func TestMongoWrapper(t *testing.T) {
	background := context.Background()
	client, e := mongo.Connect(background, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, e, "Error connecting to DB")
	defer func() {
		err := client.Disconnect(background)
		assert.NoError(t, err)
	}()
	database := client.Database("official-driver")
	metricsCollection := database.Collection("metrics")
	_, e = metricsCollection.DeleteMany(nil, bsonx.Doc{})
	require.NoError(t, e, "Cleaning metrics collection failed")
	wrapper, err := Builder().
		AppName("mongo-wrapper-test").
		DatabaseName("official-driver").
		AddIndex2Collection("test", NewIndexBuilder().Build()).
		URI("mongodb://localhost:27017").
		AddHook(timeHook(metricsCollection)).
		BuildAndConnect()
	require.NoError(t, err, "Error connecting to DB")
	defer func() {
		err := wrapper.Disconnect(background)
		assert.NoError(t, err)
	}()
	var collWrapper CollectionWrapper
	collWrapper, err = wrapper.Collection("test")
	_, err = collWrapper.RemoveMany(background, bsonx.Doc{})
	require.NoError(t, err, "Cleaning the collection failed")
	s := &suite{
		wrapper:  collWrapper,
		inserted: 0,
	}
	t.Run("Insert", s.insert)
	t.Run("Insert many", s.insertMany)
	t.Run("Find One", s.findOne)
	t.Run("Find Many", s.find)
	t.Run("Find and Modify", s.findAndModify)
	t.Run("Count", s.count)
	t.Run("Update all", s.updateAll)
	t.Run("Update one", s.updateOne)
	t.Run("Remove", s.remove)
	t.Run("Aggregate", s.aggregate)
	t.Run("Internal collection", s.internalCollection)
	t.Run("Errors", s.errors)
}

func documentGen() gopter.Gen {
	return gen.StructPtr(reflect.TypeOf(&Document{}), map[string]gopter.Gen{
		"ID":    objectIdGen(),
		"Name":  gen.AlphaString(),
		"Age":   gen.IntRange(10, 40),
		"Date":  gen.TimeRange(time.Now(), 20*time.Hour),
		"Child": subGen(),
	})
}

func objectIdGen() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		return gopter.NewGenResult(primitive.NewObjectID(), gopter.NoShrinker)
	}
}

func subGen() gopter.Gen {
	return gen.StructPtr(reflect.TypeOf(&Sub{}), map[string]gopter.Gen{
		"Numbers": gen.SliceOfN(10, gen.IntRange(0, 10)),
		"Bool":    gen.Bool(),
	})
}

func timeHook(coll *mongo.Collection) BeforeHook {
	return func(ctx context.Context, info QueryInfo) AfterHook {
		start := time.Now()
		// Extract whatever needed from context
		return func(err error) {
			var document bsonx.Doc
			end := time.Now()
			document = document.Append("operation", bsonx.String(info.OperationName)).
				Append("startTimeSec", bsonx.Int64(start.Unix()*1000)).
				Append("startTimeNano", bsonx.Int32(int32(start.Nanosecond()))).
				Append("endTimeSec", bsonx.Int64(end.Unix()*1000)).
				Append("endTimeNano", bsonx.Int32(int32(end.Nanosecond())))
			if err != nil {
				document = document.Append("failed", bsonx.String(err.Error()))
			}
			coll.InsertOne(ctx, document)
		}
	}
}

func panicHookStr(ctx context.Context, info QueryInfo) AfterHook {
	return func(err error) {
		panic("Something bad happened")
	}
}

func panicHookError(ctx context.Context, info QueryInfo) AfterHook {
	return func(err error) {
		panic(fmt.Errorf("Something bad happened as an error"))
	}
}

func panicHookUnknown(ctx context.Context, info QueryInfo) AfterHook {
	return func(err error) {
		panic(5)
	}
}
