package data

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	// NoSQL: module containing Mongo api client
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// TODO "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type AccommodationRepository struct {
	cli    *mongo.Client
	logger *log.Logger
}

func NewAccommodationRepository(ctx context.Context, logger *log.Logger) (*AccommodationRepository, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &AccommodationRepository{
		cli:    client,
		logger: logger,
	}, nil
}

func (ar *AccommodationRepository) Disconnect(ctx context.Context) error {
	err := ar.cli.Disconnect(ctx)
	if err != nil {
		ar.logger.Fatal(err.Error())
		return err
	}
	return nil
}

func (ar *AccommodationRepository) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ar.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		ar.logger.Println(err)
	}

	databases, err := ar.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ar.logger.Println(err)
	}
	fmt.Println(databases)
}

func (ar *AccommodationRepository) CreateAccommodation(ctx context.Context, accommodation *Accommodation) error {
	collection := ar.getAccommodationCollection()

	_, err := collection.InsertOne(ctx, accommodation)
	if err != nil {
		ar.logger.Println(err)
		return err
	}

	return nil
}

func (ar *AccommodationRepository) GetAllAccommodations(ctx context.Context) ([]*Accommodation, error) {
	collection := ar.getAccommodationCollection()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var accommodations []*Accommodation
	if err := cursor.All(ctx, &accommodations); err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return accommodations, nil
}

func (ar *AccommodationRepository) GetAccommodation(ctx context.Context, id primitive.ObjectID) (*Accommodation, error) {
	collection := ar.getAccommodationCollection()

	var accommodation Accommodation
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&accommodation)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return &accommodation, nil
}

func (ar *AccommodationRepository) UpdateAccommodation(ctx context.Context, accommodation *Accommodation) error {
	collection := ar.getAccommodationCollection()

	filter := bson.M{"_id": accommodation.ID}
	update := bson.M{"$set": accommodation}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ar.logger.Println(err)
		return err
	}

	return nil
}

func (ar *AccommodationRepository) DeleteAccommodation(ctx context.Context, id primitive.ObjectID) error {
	collection := ar.getAccommodationCollection()

	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		ar.logger.Println(err)
		return err
	}

	return nil
}

func (ar *AccommodationRepository) getAccommodationCollection() *mongo.Collection {
	patientDatabase := ar.cli.Database("mongoDemo")
    patientsCollection := patientDatabase.Collection("accommodations")
	return patientsCollection
}
