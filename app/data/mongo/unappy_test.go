package mongo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func badCommand(t *testing.T) {
	_, err := Builder().AddCustomCommandOption(bson.M{
		"unknown command": "whatever",
	}, false).DatabaseName("unreal").AddIndex2Collection("cappedCol", NewIndexBuilder().Build()).BuildAndConnect()
	require.Error(t, err, "Unknown command should have failed")
}

func badDB(t *testing.T) {
	_, err := Builder().BuildAndConnect()
	require.Error(t, err, "No db name")
}

func badCollection(t *testing.T) {
	_, err := Builder().DatabaseName("unreal").BuildAndConnect()
	require.Error(t, err, "No collection supplied")
}

func badAuth(t *testing.T) {
	_, err := Builder().DatabaseName("official-driver").
		UsernamePassword("whatever", "password").
		AddIndex2Collection("anotherCol", NewIndexBuilder().Keys().Asc("one").Desc("two").
			Options().Unique(true).Done().
			Build()).
		BuildAndConnect()
	require.Error(t, err, "DB does not have Auth")
}

func badFlow(t *testing.T) {
	wrapper, err := Builder().DatabaseName("official-driver").URI("mongodb://localhost:27017").
		AddIndex2Collection("anotherCol", NewIndexBuilder().Keys().Asc("one").Desc("two").
			Options().Unique(true).Done().
			Build()).
		BuildAndConnect()
	require.NoError(t, err, "This one should connect")
	_, err = wrapper.Collection("unknownCollection")
	require.Error(t, err, "no such collection was found")
}

func errorDuringDBOperation(t *testing.T) {
	var gotErrors int
	errorAwareHook := func(ctx context.Context, info QueryInfo) AfterHook {
		return func(err error) {
			if err != nil {
				gotErrors++
			}
		}
	}
	wrapper, err := Builder().
		DatabaseName("official-driver").
		AddIndex2Collection("test", NewIndexBuilder().Build()).
		AddHook(errorAwareHook).
		BuildAndConnect()
	require.NoError(t, err, "Error connecting to DB")
	collection, err := wrapper.Collection("test")
	require.NoError(t, err, "Collection unknown...")
	var result int
	err = collection.FindOne(context.Background(), "That's not a valid query", &result) // wrong query
	require.Error(t, err, "Find should have failed")
	err = collection.FindOne(context.Background(), nil, &result) // wrong result type
	require.Error(t, err, "Find should have failed")
	time.Sleep(time.Second)
	require.Equal(t, 2, gotErrors, "No error treated in hook")
}

func badHook(t *testing.T) {
	var hookCounter int
	var mx sync.Mutex
	sanityHook := func(ctx context.Context, info QueryInfo) AfterHook {
		return func(err error) {
			mx.Lock()
			defer mx.Unlock()
			hookCounter += 3
		}
	}
	handlerCheck := func(e error) {
		require.Error(t, e, "There should be error")
		mx.Lock()
		defer mx.Unlock()
		hookCounter++
	}
	wrapper, err := Builder().
		DatabaseName("official-driver").
		URI("mongodb://localhost:27017").
		AddIndex2Collection("badcollection", NewIndexBuilder().Build()).
		AddHook(sanityHook).
		AddHook(panicHookError).
		AddHook(panicHookStr).
		AddHook(panicHookUnknown).
		ErrorHandler(handlerCheck).
		BuildAndConnect()
	require.NoError(t, err, "This one should connect")
	collection, err := wrapper.Collection("badcollection")
	require.NoError(t, err, "This collection should be there")
	_, err = collection.InsertOne(context.Background(), bson.M{
		"JustChecking": true,
	})
	require.NoError(t, err, "Insertion failed")
	time.Sleep(time.Second)
	require.Equal(t, 6, hookCounter, "Hooks didn't finish in time")
}
