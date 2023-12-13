package data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type NotificationsRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// Constructor
func New(ctx context.Context, logger *log.Logger) (*NotificationsRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	nr := &NotificationsRepo{
		cli:    client,
		logger: logger,
	}

	return nr, nil
}

// Disconnect
func (ar *NotificationsRepo) Disconnect(ctx context.Context) error {
	err := ar.cli.Disconnect(ctx)
	if err != nil {
		ar.logger.Fatal(err.Error())
		return err
	}
	return nil
}

// Check database connection
func (ar *NotificationsRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := ar.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		ar.logger.Println(err)
	}

	// Print available databases
	databases, err := ar.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ar.logger.Println(err)
	}
	ar.logger.Println(databases)
}

// Repo methods

func (ar *NotificationsRepo) AddRating(rating *RatingAccommodation) error {
	ratingsCollection := ar.getRatingsCollection()

	_, err := ratingsCollection.InsertOne(context.Background(), rating)
	if err != nil {
		return err
	}

	return nil
}

func (ar *NotificationsRepo) AddHostRating(rating *RatingHost) error {
	ratingsCollection := ar.getHostRatingsCollection()

	_, err := ratingsCollection.InsertOne(context.Background(), rating)
	if err != nil {
		return err
	}

	return nil
}

func (ar *NotificationsRepo) UpdateHostRating(id primitive.ObjectID, newRating *RatingHost) error {
	ratingsCollection := ar.getHostRatingsCollection()
	filter := bson.M{"_id": id}

	update := bson.M{
		"$set": bson.M{
			"time": newRating.Time,
			"rate": newRating.Rate,
		},
	}

	_, err := ratingsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (ar *NotificationsRepo) DeleteHostRating(id primitive.ObjectID, idUser primitive.ObjectID) error {
	ratingsCollection := ar.getHostRatingsCollection()
	filter := bson.M{"_id": id}

	var rating RatingHost
	err := ratingsCollection.FindOne(context.Background(), filter).Decode(&rating)
	if err != nil {
		return err
	}

	if rating.ID == primitive.NilObjectID {
		return errors.New("no rating found with this ID")
	}

	if rating.GuestID != idUser {
		return errors.New("user did not create this rating")
	}

	result, err := ratingsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no documents deleted")
	}

	return nil
}

func (ar *NotificationsRepo) FindRatingById(ctx context.Context, id primitive.ObjectID) (*RatingAccommodation, error) {
	collection := ar.getRatingsCollection()

	var rating RatingAccommodation
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rating)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return &rating, nil
}

func (ar *NotificationsRepo) FindAccommodationRatingByGuest(ctx context.Context, accommodationId primitive.ObjectID, guestId primitive.ObjectID) (*RatingAccommodation, error) {
	collection := ar.getRatingsCollection()

	query := bson.M{"idAccommodation": accommodationId, "idGuest": guestId}
	var rating RatingAccommodation
	err := collection.FindOne(ctx, query).Decode(&rating)

	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return &rating, nil
}

func (ar *NotificationsRepo) FindHostRatingByGuest(ctx context.Context, idHost primitive.ObjectID, guestId primitive.ObjectID) (*RatingHost, error) {
	collection := ar.getHostRatingsCollection()

	query := bson.M{"idHost": idHost, "idGuest": guestId}
	fmt.Println(query)
	var rating RatingHost
	err := collection.FindOne(ctx, query).Decode(&rating)

	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return &rating, nil
}

func (ar *NotificationsRepo) GetAllAccommodationRatings(ctx context.Context) ([]RatingAccommodation, error) {
	ratingsCollection := ar.getRatingsCollection()

	cursor, err := ratingsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingAccommodation
	for cursor.Next(ctx) {
		var rating RatingAccommodation
		if err := cursor.Decode(&rating); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (ar *NotificationsRepo) GetAllAccommodationRatingsByUser(ctx context.Context, userID primitive.ObjectID) ([]RatingAccommodation, error) {
	ratingsCollection := ar.getRatingsCollection()

	cursor, err := ratingsCollection.Find(ctx, bson.M{"idGuest": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingAccommodation
	for cursor.Next(ctx) {
		var rating RatingAccommodation
		if err := cursor.Decode(&rating); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (ar *NotificationsRepo) FindHostRatingById(ctx context.Context, id primitive.ObjectID) (*RatingHost, error) {
	collection := ar.getHostRatingsCollection()

	var rating RatingHost
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rating)
	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	return &rating, nil
}

func (ar *NotificationsRepo) GetAllHostRatings(ctx context.Context) ([]RatingHost, error) {
	ratingsCollection := ar.getHostRatingsCollection()

	cursor, err := ratingsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingHost
	for cursor.Next(ctx) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (ar *NotificationsRepo) GetAllHostRatingsByUser(ctx context.Context, userID primitive.ObjectID) ([]RatingHost, error) {
	ratingsCollection := ar.getHostRatingsCollection()

	cursor, err := ratingsCollection.Find(ctx, bson.M{"idHost": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingHost
	for cursor.Next(ctx) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (ar *NotificationsRepo) GetHostRatings(ctx context.Context, hostUsername string) ([]RatingHost, error) {
	ratingsCollection := ar.getHostRatingsCollection()

	filter := bson.M{"hostUsername": hostUsername}

	cursor, err := ratingsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingHost
	for cursor.Next(ctx) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (ar *NotificationsRepo) GetRatingsByAccommodationID(accommodationID primitive.ObjectID) ([]RatingAccommodation, error) {
	ratingsCollection := ar.getRatingsCollection()
	var ratings []RatingAccommodation

	filter := bson.M{"idAccommodation": accommodationID}

	cursor, err := ratingsCollection.Find(context.TODO(), filter)
	if err != nil {
		return ratings, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var rating RatingAccommodation
		if err := cursor.Decode(&rating); err != nil {
			return ratings, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (ar *NotificationsRepo) GetRatingsByHostUsername(username string) ([]RatingHost, error) {
	ratingsCollection := ar.getHostRatingsCollection()
	var ratings []RatingHost

	filter := bson.M{"hostUsername": username}

	cursor, err := ratingsCollection.Find(context.TODO(), filter)
	if err != nil {
		return ratings, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			return ratings, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (ar *NotificationsRepo) UpdateRatingAccommodationByID(id primitive.ObjectID, newRate int) error {
	ratingsCollection := ar.getRatingsCollection()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"rate": newRate,
			"time": time.Now(),
		},
	}

	result, err := ratingsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("no documents updated")
	}

	return nil
}

func (ar *NotificationsRepo) DeleteRatingAccommodationByID(id primitive.ObjectID, idUser primitive.ObjectID) error {
	ratingsCollection := ar.getRatingsCollection()

	filter := bson.M{"_id": id}

	var rating RatingAccommodation
	err := ratingsCollection.FindOne(context.Background(), filter).Decode(&rating)
	if err != nil {
		return err
	}

	if rating.ID == primitive.NilObjectID {
		return errors.New("no rating found with this ID")
	}

	if rating.GuestID != idUser {
		return errors.New("user did not create this rating")
	}

	result, err := ratingsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no documents deleted")
	}

	return nil
}

// Getting DB collections

func (ar *NotificationsRepo) getNotificationsCollection() *mongo.Collection {
	notificationDatabase := ar.cli.Database("notificationDB")
	notificationsCollection := notificationDatabase.Collection("notifications")
	return notificationsCollection
}

func (ar *NotificationsRepo) getRatingsCollection() *mongo.Collection {
	notificationDatabase := ar.cli.Database("notificationDB")
	ratingsCollection := notificationDatabase.Collection("ratings")
	return ratingsCollection
}

func (ar *NotificationsRepo) getHostRatingsCollection() *mongo.Collection {
	notificationDatabase := ar.cli.Database("notificationDB")
	hostRatingsCollection := notificationDatabase.Collection("hostRatings")
	return hostRatingsCollection
}
