package mongocom

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"path/filepath"
	"time"
)

type FileStruct struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	FileParent  string             `bson:"fileParent,omitempty"`
	FileName    string             `bson:"fileName,omitempty"`
	FileContent string             `bson:"fileContent,omitempty"`
	IsFolder    bool               `bson:"isFolder,omitempty"`
}

func InitiateMongoClient() *mongo.Client {
	connectionString := "" // Use it from somewhere
	connectionOption := options.Client()
	connectionOption.ApplyURI(connectionString)
	connectionOption.SetMaxPoolSize(60)

	mongoClient, err := mongo.Connect(context.Background(), connectionOption)
	if err != nil {
		log.Fatalf("Cannot connection mongodb: %s\n", err.Error())
	}

	return mongoClient
}

func FindFileByName(fileName string) FileStruct {
	dir, base := filepath.Split(fileName)

	mongoClient := InitiateMongoClient()

	dbObject := mongoClient.Database("griddata")
	collection := dbObject.Collection("testCollection")

	mongoContext, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var result FileStruct
	findCursor := collection.FindOne(mongoContext, bson.M{
		"fileName":   base,
		"fileParent": dir,
	})
	findCursor.Decode(&result)

	return result
}

func ListDirectory(parentDirectory string) []FileStruct {
	mongoClient := InitiateMongoClient()

	dbObject := mongoClient.Database("griddata")
	collection := dbObject.Collection("testCollection")

	mongoContext, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var result []FileStruct
	findCursor, _ := collection.Find(mongoContext, bson.M{
		"fileParent": parentDirectory,
	})

	findCursor.All(mongoContext, &result)

	return result
}

func WriteFile(f *FileStruct) {
	mongoClient := InitiateMongoClient()

	dbObject := mongoClient.Database("griddata")
	collection := dbObject.Collection("testCollection")

	mongoContext, _ := context.WithTimeout(context.Background(), 10*time.Second)

	collection.DeleteOne(mongoContext, &bson.M{"fileName": f.FileName})
	collection.InsertOne(mongoContext, *f, nil)
}

func RemoveFile(fileParent string, fileName string) {
	mongoClient := InitiateMongoClient()

	dbObject := mongoClient.Database("griddata")
	collection := dbObject.Collection("testCollection")

	mongoContext, _ := context.WithTimeout(context.Background(), 10*time.Second)

	collection.DeleteOne(mongoContext, &bson.M{"fileParent": fileParent, "fileName": fileName})
}

func CreateDirectory(parent string, name string) {
	mongoClient := InitiateMongoClient()

	dbObject := mongoClient.Database("griddata")
	collection := dbObject.Collection("testCollection")

	mongoContext, _ := context.WithTimeout(context.Background(), 10*time.Second)

	fileStruct := &FileStruct{
		FileParent: parent,
		FileName:   name,
		IsFolder:   true,
	}

	collection.DeleteOne(mongoContext, &bson.M{"fileName": name, "fileParent": parent})
	collection.InsertOne(mongoContext, fileStruct, nil)
}
