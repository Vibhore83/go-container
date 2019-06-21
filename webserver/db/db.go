/*
 * db.go handles interaction with MongoDB. All CRUD operations are implemented.
 *
 * API version: 1.0.0
 * Author Credits - Arun K
 */

package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"webserver/logging"
)

var client *mongo.Client

//ErrNoMatchDocument is returned when no matching document is found
var ErrNoMatchDocument = errors.New("No matching document")

//ErrMultipleDocExist is returned when multiple meta docs exist
var ErrMultipleDocExist = errors.New("More than expected number of documents")

const dbName = "infrabuilder"
const tbColl = "testbed"
const tbMetaColl = "testbedmeta"

func init() {
	var err error

	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

//getTestBedCollection returns testbed collection
func getTestBedCollection() *mongo.Collection {
	return client.Database(dbName).Collection(tbColl)
}

// getTestBedMetaCollection returns testbedMeta collection
func getTestBedMetaCollection() *mongo.Collection {
        return client.Database(dbName).Collection(tbMetaColl)
}

//InsertTestBed inserts testbed data into MongoDB
func InsertTestBed(ctx context.Context, tb *TestBed) (*mongo.InsertOneResult, error) {
	insertResult, err := getTestBedCollection().InsertOne(ctx, tb)
	return insertResult, err
}

//UpdateTestBedStatus updates status field in testbed collection for a document
func UpdateTestBedStatus(ctx context.Context, id, status string) (*mongo.UpdateResult, error) {
	colQuerier := bson.M{"_id": id}
	change := bson.M{"$set": bson.M{"status": status}}

	updateResult, err := getTestBedCollection().UpdateOne(ctx, colQuerier, change)
	return updateResult, err
}

//UpdateContainerProperty updates property for a container in TestBed document
func UpdateContainerProperty(ctx context.Context, id, container, property string, value interface{}) (*mongo.UpdateResult, error) {
	colQuerier := bson.M{"_id": id, "container": bson.M{"$elemMatch": bson.M{"image": container}}}
	field := fmt.Sprintf("container.$.%v", property)
	change := bson.M{"$set": bson.M{field: value}}

	updateResult, err := getTestBedCollection().UpdateOne(ctx, colQuerier, change)
	return updateResult, err
}

//GetTestBedFromID returns document corresponding to a testbed
func GetTestBedFromID(ctx context.Context, id string) (TestBed, error) {
	tb := TestBed{}
	colQuerier := bson.M{"_id": id}
	err := getTestBedCollection().FindOne(ctx, colQuerier).Decode(&tb)
	if err != nil {
		logging.Error.Println(err)
	}
	return tb, err
}

//GetContainerProperty returns container property for a test bed
func GetContainerProperty(ctx context.Context, id, container string) (*ContainerProp, error) {
	type containerStruct struct {
		Container []ContainerProp
	}
	cntr := containerStruct{}
	colQuerier := bson.M{"_id": id}
	projection := bson.M{"container": bson.M{"$elemMatch": bson.M{"image": container}}}
	err := getTestBedCollection().FindOne(ctx, colQuerier, options.FindOne().SetProjection(projection)).Decode(&cntr)

	if len(cntr.Container) > 0 {
		return &cntr.Container[0], err
	} else {
		log.Println("Could not find container")
		return nil, ErrNoMatchDocument
	}
}

//DeleteTestBed removes a test bed document from the collection
func DeleteTestBed(ctx context.Context, id string) (*mongo.DeleteResult, error) {
	colQuerier := bson.M{"_id": id}
	deleteResult, err := getTestBedCollection().DeleteOne(ctx, colQuerier)
	return deleteResult, err
}

//InitTestBedMetaCollection initialize testbedmeta collection
func InitTestBedMetaCollection(ctx context.Context) error {
	colQuerier := bson.M{}
	count, err := getTestBedMetaCollection().CountDocuments(ctx, colQuerier)
	if err != nil {
		return err
	}
	if count == 0 {
		rec := NewTestBedMeta()
		_, err := getTestBedMetaCollection().InsertOne(ctx, rec)
		if err != nil {
			return err
		}
	} else if count > 1 {
		log.Println("More than one document exist")
		return ErrMultipleDocExist
	}
	return nil
}

//GetTestBedMeta returns TestBedMeta document
func GetTestBedMeta(ctx context.Context) (TestBedMeta, error) {
	tbm := TestBedMeta{}
	colQuerier := bson.M{}
	err := getTestBedMetaCollection().FindOne(ctx, colQuerier).Decode(&tbm)
	return tbm, err
}

//AddPortToMeta appends a port to Allocated ports list
func AddPortToMeta(ctx context.Context, port int) error {
	colQuerier := bson.M{}
	change := bson.M{"$addToSet": bson.M{"allocatedPorts": port}}
	res := getTestBedMetaCollection().FindOneAndUpdate(ctx, colQuerier, change)
	return res.Err()
}

//DeletePortFromMeta appends a port to Allocated ports list
func DeletePortFromMeta(ctx context.Context, port int) error {
	colQuerier := bson.M{}
	change := bson.M{"$pull": bson.M{"allocatedPorts": port}}
	res := getTestBedMetaCollection().FindOneAndUpdate(ctx, colQuerier, change)
	return res.Err()
}
