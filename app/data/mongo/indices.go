package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type IndexBuilder interface {
	Keys() keysBuilder
	Build() mongo.IndexModel
}

type keysBuilder interface {
	Asc(name string) keysBuilder
	Desc(name string) keysBuilder
	Text(name string) keysBuilder
	Sphere(name string) keysBuilder
	Options() optionsBuilder
	// There is a hash index, but I'm not sure when we will use it.
}

type optionsBuilder interface {
	ExpireAfterSeconds(expireAfterSeconds int32) optionsBuilder
	Name(name string) optionsBuilder
	Sparse(sparse bool) optionsBuilder
	StorageEngine(storageEngine bsonx.Doc) optionsBuilder
	Unique(unique bool) optionsBuilder
	Version(version int32) optionsBuilder
	DefaultLanguage(defaultLanguage string) optionsBuilder
	LanguageOverride(languageOverride string) optionsBuilder
	TextVersion(textVersion int32) optionsBuilder
	Weights(weights bsonx.Doc) optionsBuilder
	SphereVersion(sphereVersion int32) optionsBuilder
	Bits(bits int32) optionsBuilder
	Max(max float64) optionsBuilder
	Min(min float64) optionsBuilder
	BucketSize(bucketSize int32) optionsBuilder
	PartialFilterExpression(partialFilterExpression bsonx.Doc) optionsBuilder
	Collation(collation *options.Collation) optionsBuilder
	Done() IndexBuilder
}

type indexBuilderImpl struct {
	keys    bsonx.Doc
	options *options.IndexOptions
}

type keysBuilderImpl struct {
	indexBuilder *indexBuilderImpl
	keys         bsonx.Doc
}

type optionsBuilderImpl struct {
	indexBuilder        *indexBuilderImpl
	mongoOptionsBuilder *options.IndexOptions
}

func NewIndexBuilder() IndexBuilder {
	return new(indexBuilderImpl)
}

// #### Index builder

func (ib *indexBuilderImpl) Keys() keysBuilder {
	return &keysBuilderImpl{
		indexBuilder: ib,
	}
}

func (ib *indexBuilderImpl) setOptions() optionsBuilder {
	return &optionsBuilderImpl{
		indexBuilder:        ib,
		mongoOptionsBuilder: options.Index(),
	}
}

func (ib *indexBuilderImpl) Build() mongo.IndexModel {
	return mongo.IndexModel{
		Keys:    ib.keys,
		Options: ib.options,
	}
}

// #### Options builder

func (ob *optionsBuilderImpl) ExpireAfterSeconds(expireAfterSeconds int32) optionsBuilder {
	ob.mongoOptionsBuilder.SetExpireAfterSeconds(expireAfterSeconds)
	return ob
}

func (ob *optionsBuilderImpl) Name(name string) optionsBuilder {
	ob.mongoOptionsBuilder.SetName(name)
	return ob
}

func (ob *optionsBuilderImpl) Sparse(sparse bool) optionsBuilder {
	ob.mongoOptionsBuilder.SetSparse(sparse)
	return ob
}

func (ob *optionsBuilderImpl) StorageEngine(storageEngine bsonx.Doc) optionsBuilder {
	ob.mongoOptionsBuilder.SetStorageEngine(storageEngine)
	return ob
}

func (ob *optionsBuilderImpl) Unique(unique bool) optionsBuilder {
	ob.mongoOptionsBuilder.SetUnique(unique)
	return ob
}

func (ob *optionsBuilderImpl) Version(version int32) optionsBuilder {
	ob.mongoOptionsBuilder.SetVersion(version)
	return ob
}

func (ob *optionsBuilderImpl) DefaultLanguage(defaultLanguage string) optionsBuilder {
	ob.mongoOptionsBuilder.SetDefaultLanguage(defaultLanguage)
	return ob
}

func (ob *optionsBuilderImpl) LanguageOverride(languageOverride string) optionsBuilder {
	ob.mongoOptionsBuilder.SetLanguageOverride(languageOverride)
	return ob
}

func (ob *optionsBuilderImpl) TextVersion(textVersion int32) optionsBuilder {
	ob.mongoOptionsBuilder.SetTextVersion(textVersion)
	return ob
}

func (ob *optionsBuilderImpl) Weights(weights bsonx.Doc) optionsBuilder {
	ob.mongoOptionsBuilder.SetWeights(weights)
	return ob
}

func (ob *optionsBuilderImpl) SphereVersion(sphereVersion int32) optionsBuilder {
	ob.mongoOptionsBuilder.SetSphereVersion(sphereVersion)
	return ob
}

func (ob *optionsBuilderImpl) Bits(bits int32) optionsBuilder {
	ob.mongoOptionsBuilder.SetBits(bits)
	return ob
}

func (ob *optionsBuilderImpl) Max(max float64) optionsBuilder {
	ob.mongoOptionsBuilder.SetMax(max)
	return ob
}

func (ob *optionsBuilderImpl) Min(min float64) optionsBuilder {
	ob.mongoOptionsBuilder.SetMin(min)
	return ob
}

func (ob *optionsBuilderImpl) BucketSize(bucketSize int32) optionsBuilder {
	ob.mongoOptionsBuilder.SetBucketSize(bucketSize)
	return ob
}

func (ob *optionsBuilderImpl) PartialFilterExpression(partialFilterExpression bsonx.Doc) optionsBuilder {
	ob.mongoOptionsBuilder.SetPartialFilterExpression(partialFilterExpression)
	return ob
}

func (ob *optionsBuilderImpl) Collation(collation *options.Collation) optionsBuilder {
	ob.mongoOptionsBuilder.SetCollation(collation)
	return ob
}

func (ob *optionsBuilderImpl) Done() IndexBuilder {
	ob.indexBuilder.options = ob.mongoOptionsBuilder
	return ob.indexBuilder
}

// ##### Keys builder ######

func (kb *keysBuilderImpl) Asc(name string) keysBuilder {
	kb.keys = kb.keys.Append(name, bsonx.Int32(1))
	return kb
}

func (kb *keysBuilderImpl) Desc(name string) keysBuilder {
	kb.keys = kb.keys.Append(name, bsonx.Int32(-1))
	return kb
}

func (kb *keysBuilderImpl) Text(name string) keysBuilder {
	kb.keys = kb.keys.Append(name, bsonx.String("text"))
	return kb
}

func (kb *keysBuilderImpl) Sphere(name string) keysBuilder {
	kb.keys = kb.keys.Append(name, bsonx.String("2dsphere"))
	return kb
}

func (kb *keysBuilderImpl) Options() optionsBuilder {
	kb.indexBuilder.keys = kb.keys
	return kb.indexBuilder.setOptions()
}
