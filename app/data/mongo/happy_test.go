package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

func (m *suite) insert(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Insert one", prop.ForAll(
		func(doc *Document) bool {
			id, err := m.wrapper.InsertOne(context.Background(), doc)
			m.inserted++
			return assert.NoError(t, err, "Insert failed") &&
				assert.NotZero(t, id, "Id returned is a zero")
		}, documentGen(),
	))
	properties.TestingRun(t)
}

func (m *suite) insertMany(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Insert many - ordered", prop.ForAll(
		func(docs []*Document) bool {
			slice := make([]interface{}, len(docs))
			for i, doc := range docs {
				slice[i] = doc
			}
			ids, err := m.wrapper.InsertOrdered(context.Background(), slice)
			m.inserted = m.inserted + len(ids)
			return assert.NoError(t, err, "Ordered Insert failed") &&
				assert.Len(t, ids, len(docs), "Amount of returned ids doesn't match intended amount")
		}, gen.SliceOfN(10, documentGen()),
	))

	properties.Property("Insert many - unordered", prop.ForAll(
		func(docs []*Document) bool {
			slice := make([]interface{}, len(docs))
			for i, doc := range docs {
				slice[i] = doc
			}
			ids, err := m.wrapper.InsertUnordered(context.Background(), slice)
			m.inserted = m.inserted + len(ids)
			return assert.NoError(t, err, "Unordered Insert failed") &&
				assert.Len(t, ids, len(docs), "Amount of returned ids doesn't match intended amount")
		}, gen.SliceOfN(10, documentGen()),
	))

	properties.TestingRun(t)
}

func (m *suite) find(t *testing.T) {
	var docs []Document
	query := bson.M{
		"child.numbers": bson.M{
			"$in": []int{5},
		},
	}
	lim := 100
	err := m.wrapper.Find(context.Background(), query, nil, nil, lim, 0, &docs)
	require.NoError(t, err, "Find of many failed")
	require.Condition(t, func() bool {
		for _, doc := range docs {
			require.Contains(t, doc.Child.Numbers, 5, "No number 5")
		}
		return true
	}, "not all numbers have 5 in them")
}

func (m *suite) findOne(t *testing.T) {
	var doc Document
	query := bson.M{
		"age": bson.M{
			"$gt": 10,
		},
	}
	err := m.wrapper.FindOne(context.Background(), query, &doc)
	require.NoError(t, err, "Find failed")
	require.True(t, doc.Age > 10, "Age is not as expected")
}

func (m *suite) findAndModify(t *testing.T) {
	query := bson.M{
		"age": bson.M{
			"$lt": 1000,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"age": 1000,
		},
	}
	properties := gopter.NewProperties(nil)
	properties.Property("Find and modify many times", prop.ForAll(
		func(returnUpdated bool) bool {
			var doc Document
			// Previous version
			err := m.wrapper.FindAndModify(nil, query, update, nil, returnUpdated, false, &doc)
			if !returnUpdated {
				var updateDoc Document
				err = m.wrapper.FindOne(nil, bson.M{"_id": doc.ID}, &updateDoc)
				return assert.NoError(t, err, "Updated document wasn't found") &&
					assert.Condition(t, func() bool {
						return assert.Equal(t, doc.ID, updateDoc.ID, "ID should be the same") &&
							assert.Equal(t, doc.Name, updateDoc.Name, "Name is not the same") &&
							assert.True(t, doc.Age < 1000, "Previous age value was not below 1000") &&
							assert.Equal(t, 1000, updateDoc.Age, "Updated age is not 1000")
					})
			} else {
				return assert.NoError(t, err, "Failed find and update")
			}
		}, gen.Bool(),
	))
	properties.TestingRun(t)
}

func (m *suite) count(t *testing.T) {
	result, err := m.wrapper.Count(nil, bsonx.Doc{})
	require.NoError(t, err, "Count failed")
	require.Equal(t, m.inserted, int(result), "Count is not as expected")
}

func (m *suite) updateAll(t *testing.T) {
	var document bsonx.Doc
	document = document.Append("date", bsonx.Time(time.Now()))
	update := bson.M{
		"$set": bson.M{
			"date": time.Now(),
		},
	}
	matched, modified, err := m.wrapper.UpdateAll(nil, bsonx.Doc{}, update)
	require.NoError(t, err, "Failed to update All")
	require.Equal(t, int64(m.inserted), matched, "All should have match")
	require.Equal(t, int64(m.inserted), modified, "All should have modified")
}

func (m *suite) updateOne(t *testing.T) {
	update := bson.M{
		"$set": bson.M{
			"date": time.Now().Add(10 * time.Hour),
		},
	}
	_, err := m.wrapper.UpdateOne(nil, bsonx.Doc{}, update, false)
	require.NoError(t, err, "Failed to update one document")
}

func (m *suite) remove(t *testing.T) {
	properties := gopter.NewProperties(nil)
	properties.Property("Remove documents", prop.ForAll(
		func(many bool) bool {
			if many {
				query := bson.M{
					"age": bson.M{
						"$lt": 1000,
					},
				}
				removedCount, err := m.wrapper.RemoveMany(nil, query)
				m.inserted = m.inserted - int(removedCount)
				return assert.NoError(t, err, "Failed to remove documents")
			} else {
				query := bson.M{
					"age": 1000,
				}
				removedCount, err := m.wrapper.RemoveOne(nil, query)
				m.inserted--
				return assert.NoError(t, err, "Remove one document failed") &&
					assert.Equal(t, int64(1), removedCount, "Should have remove 1 document")
			}
		}, gen.Bool(),
	))
	properties.TestingRun(t)
}

func (m *suite) aggregate(t *testing.T) {
	var docs []Document
	matchStage := bson.M{
		"$match": bson.M{
			"age": bson.M{
				"$gte": 0,
			},
		},
	}
	sortStage := bson.M{
		"$sort": bson.M{
			"age": 1,
		},
	}
	err := m.wrapper.Aggregate(nil, []interface{}{matchStage, sortStage}, &docs)
	require.NoError(t, err, "Aggregation failed")
	require.Len(t, docs, m.inserted, "Not all documents retrieved")
	require.Condition(t, func() bool {
		for _, doc := range docs {
			require.True(t, doc.Age >= 0, "Age is negative")
		}
		return true
	}, "Not all documents have a proper age")
}

func (m *suite) internalCollection(t *testing.T) {
	require.IsType(t, &mongo.Collection{}, m.wrapper.InternalCollection(), "Not mongo type")
}

func (m *suite) errors(t *testing.T) {
	err := m.wrapper.Find(context.Background(), nil, nil, nil, 0, 0, nil)
	require.Error(t, err, "can't unmarshal nil")
	var numbers []int
	err = m.wrapper.Find(context.Background(), nil, nil, nil, 0, 0, &numbers)
	require.Error(t, err, "wrong result type")
}
