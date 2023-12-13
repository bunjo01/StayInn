package data

import (
	"context"
	"errors"
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
func (nr *NotificationsRepo) Disconnect(ctx context.Context) error {
	err := nr.cli.Disconnect(ctx)
	if err != nil {
		nr.logger.Fatal(err.Error())
		return err
	}
	return nil
}

// Check database connection
func (nr *NotificationsRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := nr.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		nr.logger.Println(err)
	}

	// Print available databases
	databases, err := nr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		nr.logger.Println(err)
	}
	nr.logger.Println(databases)
}

// Repo methods

func (nr *NotificationsRepo) AddRating(rating *RatingAccommodation) error {
	ratingsCollection := nr.getRatingsCollection()

	_, err := ratingsCollection.InsertOne(context.Background(), rating)
	if err != nil {
		return err
	}

	return nil
}

func (nr *NotificationsRepo) AddHostRating(rating *RatingHost) error {
	ratingsCollection := nr.getHostRatingsCollection()

	_, err := ratingsCollection.InsertOne(context.Background(), rating)
	if err != nil {
		return err
	}

	return nil
}

func (nr *NotificationsRepo) UpdateHostRating(id primitive.ObjectID, newRating *RatingHost) error {
	ratingsCollection := nr.getHostRatingsCollection()
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

func (nr *NotificationsRepo) DeleteHostRating(id primitive.ObjectID, idUser primitive.ObjectID) error {
	ratingsCollection := nr.getHostRatingsCollection()
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

func (nr *NotificationsRepo) GetAllAccommodationRatings(ctx context.Context) ([]RatingAccommodation, error) {
	ratingsCollection := nr.getRatingsCollection()

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

func (nr *NotificationsRepo) GetAllAccommodationRatingsByUser(ctx context.Context, userID primitive.ObjectID) ([]RatingAccommodation, error) {
	ratingsCollection := nr.getRatingsCollection()

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

func (nr *NotificationsRepo) GetAllHostRatings(ctx context.Context) ([]RatingHost, error) {
	ratingsCollection := nr.getHostRatingsCollection()

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

func (nr *NotificationsRepo) GetAllHostRatingsByUser(ctx context.Context, userID primitive.ObjectID) ([]RatingHost, error) {
	ratingsCollection := nr.getHostRatingsCollection()

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

func (nr *NotificationsRepo) GetHostRatings(ctx context.Context, hostUsername string) ([]RatingHost, error) {
	ratingsCollection := nr.getHostRatingsCollection()

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

func (r *NotificationsRepo) GetRatingsByAccommodationID(accommodationID primitive.ObjectID) ([]RatingAccommodation, error) {
	ratingsCollection := r.getRatingsCollection()
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

func (r *NotificationsRepo) GetRatingsByHostUsername(username string) ([]RatingHost, error) {
	ratingsCollection := r.getHostRatingsCollection()
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

func (nr *NotificationsRepo) UpdateRatingAccommodationByID(id primitive.ObjectID, newRate int) error {
	ratingsCollection := nr.getRatingsCollection()

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

func (nr *NotificationsRepo) DeleteRatingAccommodationByID(id primitive.ObjectID, idUser primitive.ObjectID) error {
	ratingsCollection := nr.getRatingsCollection()

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

func (nr *NotificationsRepo) getNotificationsCollection() *mongo.Collection {
	notificationDatabase := nr.cli.Database("notificationDB")
	notificationsCollection := notificationDatabase.Collection("notifications")
	return notificationsCollection
}

func (nr *NotificationsRepo) getRatingsCollection() *mongo.Collection {
	notificationDatabase := nr.cli.Database("notificationDB")
	ratingsCollection := notificationDatabase.Collection("ratings")
	return ratingsCollection
}

func (nr *NotificationsRepo) getHostRatingsCollection() *mongo.Collection {
	notificationDatabase := nr.cli.Database("notificationDB")
	hostRatingsCollection := notificationDatabase.Collection("hostRatings")
	return hostRatingsCollection
}
