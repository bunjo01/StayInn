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
			"guestUsername": newRating.GuestUsername,
			"hostUsername":  newRating.HostUsername,
			"time":          newRating.Time,
			"rate":          newRating.Rate,
		},
	}

	_, err := ratingsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (nr *NotificationsRepo) DeleteHostRating(id primitive.ObjectID) error {
	ratingsCollection := nr.getHostRatingsCollection()
	filter := bson.M{"_id": id}

	_, err := ratingsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
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

func (nr *NotificationsRepo) DeleteRatingAccommodationByID(id primitive.ObjectID) error {
	ratingsCollection := nr.getRatingsCollection()

	filter := bson.M{"_id": id}

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
