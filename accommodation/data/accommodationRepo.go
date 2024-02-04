package data

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	// NoSQL: module containing Mongo api client
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type AccommodationRepository struct {
	cli *mongo.Client
}

func NewAccommodationRepository(ctx context.Context) (*AccommodationRepository, error) {
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
		cli: client,
	}, nil
}

func (ar *AccommodationRepository) Disconnect(ctx context.Context) error {
	err := ar.cli.Disconnect(ctx)
	if err != nil {
		log.Fatal(fmt.Sprintf("[acco-repo]acr#1 Unable to disconnect: %v", err))
		return err
	}
	return nil
}

func (ar *AccommodationRepository) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ar.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(fmt.Sprintf("[acco-repo]acr#2 Ping failed: %v", err))
	}

	databases, err := ar.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(fmt.Sprintf("[acco-repo]acr#3 Unable to disconnect from repo: %v", err))
	}
	fmt.Println(databases)
}

func (ar *AccommodationRepository) CreateAccommodation(ctx context.Context, accommodation *Accommodation) error {
	collection := ar.getAccommodationCollection()

	_, err := collection.InsertOne(ctx, accommodation)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#4 Failed to create accommodation: %v", err))
		return err
	}

	return nil
}

func (ar *AccommodationRepository) GetAllAccommodations(ctx context.Context) ([]*Accommodation, error) {
	collection := ar.getAccommodationCollection()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#5 Failed to get all accommodations: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var accommodations []*Accommodation
	if err := cursor.All(ctx, &accommodations); err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#6 Failed to iterate over all accommodations: %v", err))
		return nil, err
	}

	return accommodations, nil
}

func (ar *AccommodationRepository) GetAccommodationsForUser(ctx context.Context, userID primitive.ObjectID) ([]*Accommodation, error) {
	collection := ar.getAccommodationCollection()

	cursor, err := collection.Find(ctx, bson.M{"hostID": userID})
	if err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#7 Failed to get accommodations for user '%v': %v", userID, err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var accommodations []*Accommodation
	if err := cursor.All(ctx, &accommodations); err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#8 Failed to iterate over all accommodations: %v", err))
		return nil, err
	}

	return accommodations, nil
}

func (ar *AccommodationRepository) GetAccommodation(ctx context.Context, id primitive.ObjectID) (*Accommodation, error) {
	collection := ar.getAccommodationCollection()

	var accommodation Accommodation
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&accommodation)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#9 Failed to get accommodation '%v': %v", id, err))
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
		log.Error(fmt.Sprintf("[acco-repo]acr#10 Failed to update accommodation '%v': %v", accommodation.ID, err))
		return err
	}

	return nil
}

func (ar *AccommodationRepository) DeleteAccommodation(ctx context.Context, id primitive.ObjectID) error {
	collection := ar.getAccommodationCollection()

	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#11 Failed to delete accommodation '%v': %v", id, err))
		return err
	}

	return nil
}

func (ar *AccommodationRepository) DeleteAccommodationsForUser(ctx context.Context, userID primitive.ObjectID) error {
	collection := ar.getAccommodationCollection()

	filter := bson.M{"hostID": userID}
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#12 Failed to delete accommodations for user '%v': %v", userID, err))
		return err
	}

	return nil
}

func (ar *AccommodationRepository) FindAccommodationsByIDs(ctx context.Context, ids []primitive.ObjectID) (*[]Accommodation, error) {
	var accommodations []Accommodation
	if len(ids) != 0 {
		collection := ar.getAccommodationCollection()
		filter := bson.M{"_id": bson.M{"$in": ids}}

		cur, err := collection.Find(ctx, filter)
		if err != nil {
			log.Error(fmt.Sprintf("[acco-repo]acr#13 Failed to get accommodations: %v", err))
		}
		defer cur.Close(ctx)

		if err := cur.All(ctx, &accommodations); err != nil {
			log.Error(fmt.Sprintf("[acco-repo]acr#14 Failed to iterate over all accommodations: %v", err))
			return nil, err
		}

		return &accommodations, nil
	}
	return &accommodations, nil
}

// Search part

func (ar *AccommodationRepository) GetFilteredAccommodations(ctx context.Context, filters bson.M) ([]*Accommodation, error) {
	collection := ar.getAccommodationCollection()

	// Log parameters
	log.Info(fmt.Sprintf("[acco-repo]acr#15 Filter parameters: %v", filters))

	cursor, err := collection.Find(ctx, filters)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#16 Failed to get accommodations: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var accommodations []*Accommodation
	if err := cursor.All(ctx, &accommodations); err != nil {
		log.Error(fmt.Sprintf("[acco-repo]acr#17 Failed to iterate over all accommodations: %v", err))
		return nil, err
	}

	return accommodations, nil
}

func (ar *AccommodationRepository) getAccommodationCollection() *mongo.Collection {
	patientDatabase := ar.cli.Database("mongoDemo")
	patientsCollection := patientDatabase.Collection("accommodations")
	return patientsCollection
}
