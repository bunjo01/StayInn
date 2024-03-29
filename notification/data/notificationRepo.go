package data

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type NotificationsRepo struct {
	cli *mongo.Client
}

// Constructor
func New(ctx context.Context) (*NotificationsRepo, error) {
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
		cli: client,
	}

	return nr, nil
}

// Disconnect
func (nr *NotificationsRepo) Disconnect(ctx context.Context) error {
	err := nr.cli.Disconnect(ctx)
	if err != nil {
		log.Fatal(fmt.Sprintf("[noti-repo]nr#1 Failed to disconnect: %v", err))
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
		log.Error(fmt.Sprintf("[noti-repo]nr#2 Failed to ping connection: %v", err))
	}

	// Print available databases
	databases, err := nr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#3 Failed to list available database: %v", err))
	}
	log.Info(fmt.Sprintf("[noti-repo]nr#4 Show available database: %v", databases))
}

// Repo methods

func (nr *NotificationsRepo) CreateNotification(ctx context.Context, notification *Notification) error {
	collection := nr.getNotificationsCollection()

	_, err := collection.InsertOne(ctx, notification)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#5 Failed to create notification: %v", err))
		return err
	}

	return nil
}

func (nr *NotificationsRepo) GetAllNotifications(ctx context.Context, username string) ([]Notification, error) {
	notificationsCollection := nr.getNotificationsCollection()

	cursor, err := notificationsCollection.Find(ctx, bson.M{"hostUsername": username})
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#6 Failed to find all notification: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []Notification
	for cursor.Next(ctx) {
		var notification Notification
		if err := cursor.Decode(&notification); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#7 Failed to decode body: %v", err))
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, nil
}

func (nr *NotificationsRepo) AddRating(rating *RatingAccommodation) error {
	ratingsCollection := nr.getRatingsCollection()

	_, err := ratingsCollection.InsertOne(context.Background(), rating)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#9 Failed to create rating: %v", err))
		return err
	}
	return nil
}

func (nr *NotificationsRepo) AddHostRating(rating *RatingHost) error {
	ratingsCollection := nr.getHostRatingsCollection()

	_, err := ratingsCollection.InsertOne(context.Background(), rating)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#11 Failed to create host rating: %v", err))
		return err
	}
	return nil
}

func (nr *NotificationsRepo) GetRatingsByHostID(hostID primitive.ObjectID) ([]RatingHost, error) {
	ratingsCollection := nr.getHostRatingsCollection()

	filter := bson.M{"idHost": hostID}

	cursor, err := ratingsCollection.Find(context.Background(), filter)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#13 Failed to find rating by hostID: %v", err))
		return nil, err
	}
	defer cursor.Close(context.Background())

	var ratings []RatingHost
	for cursor.Next(context.Background()) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#14 Failed to decode body: %v", err))
			return nil, err
		}
		ratings = append(ratings, rating)
	}
	return ratings, nil
}

func (nr *NotificationsRepo) UpdateHostRating(id, idUser primitive.ObjectID, newRating *RatingHost) error {
	ratingsCollection := nr.getHostRatingsCollection()

	filter := bson.M{"_id": id}

	var rating RatingHost

	err := ratingsCollection.FindOne(context.Background(), filter).Decode(&rating)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#16 Failed to find rating: %v", err))
		return err
	}

	if rating.GuestID != idUser {
		log.Error("[noti-repo]nr#17 User does not match the rating creator")
		return errors.New("cannot update rating: user does not match the rating creator")
	}

	update := bson.M{
		"$set": bson.M{
			"time": newRating.Time,
			"rate": newRating.Rate,
		},
	}

	_, err = ratingsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#18 Failed to update rating: %v", err))
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
		log.Error(fmt.Sprintf("[noti-repo]nr#20 Failed to find rating with id: %v", err))
		return err
	}

	if rating.ID == primitive.NilObjectID {
		log.Error("[noti-repo]nr#21 No rating found with this ID")
		return errors.New("no rating found with this ID")
	}

	if rating.GuestID != idUser {
		log.Error("[noti-repo]nr#22 User did not create this rating")
		return errors.New("user did not create this rating")
	}

	result, err := ratingsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#23 Failed to deleted host rating: %v", err))
		return err
	}

	if result.DeletedCount == 0 {
		log.Error("[noti-repo]nr#24 No documents deleted")
		return errors.New("no documents deleted")
	}

	return nil
}

func (nr *NotificationsRepo) FindAccommodationRatingByGuest(ctx context.Context, accommodationId primitive.ObjectID, guestId primitive.ObjectID) (*RatingAccommodation, error) {
	collection := nr.getRatingsCollection()

	query := bson.M{"idAccommodation": accommodationId, "idGuest": guestId}
	var rating RatingAccommodation
	err := collection.FindOne(ctx, query).Decode(&rating)

	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#26 Failed to find accommodation rating with guestID: %v", err))
		return nil, err
	}

	return &rating, nil
}

func (nr *NotificationsRepo) FindHostRatingByGuest(ctx context.Context, idHost primitive.ObjectID, guestId primitive.ObjectID) (*RatingHost, error) {
	collection := nr.getHostRatingsCollection()

	query := bson.M{"idHost": idHost, "idGuest": guestId}
	fmt.Println(query)
	var rating RatingHost
	err := collection.FindOne(ctx, query).Decode(&rating)

	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#28 Failed to find host rating with guestID: %v", err))
		return nil, err
	}

	return &rating, nil
}

func (nr *NotificationsRepo) GetAllAccommodationRatings(ctx context.Context) ([]RatingAccommodation, error) {
	ratingsCollection := nr.getRatingsCollection()

	cursor, err := ratingsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#30 Failed to find ratings for accommodation: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingAccommodation
	for cursor.Next(ctx) {
		var rating RatingAccommodation
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#31 Failed to decode body: %v", err))
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
		log.Error("[noti-repo]nr#33 Failed to find ratings for accommodation with guestID")
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingAccommodation
	for cursor.Next(ctx) {
		var rating RatingAccommodation
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#34 Failed to decode body %v", err))
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (nr *NotificationsRepo) GetAllHostRatingsByUser(ctx context.Context, userID primitive.ObjectID) ([]RatingHost, error) {
	ratingsCollection := nr.getHostRatingsCollection()

	cursor, err := ratingsCollection.Find(ctx, bson.M{"idGuest": userID})
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#36 Failed to find ratings for host by guestID: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingHost
	for cursor.Next(ctx) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#37 Failed to decode body: %v", err))
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (nr *NotificationsRepo) GetAllAccommodationRatingsForLoggedHost(ctx context.Context, userID primitive.ObjectID) ([]RatingAccommodation, error) {
	ratingsCollection := nr.getRatingsCollection()

	cursor, err := ratingsCollection.Find(ctx, bson.M{"idHost": userID})
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#39 Failed to find ratings for accommodation for logged host: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingAccommodation
	for cursor.Next(ctx) {
		var rating RatingAccommodation
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#40 Failed to decode body: %v", err))
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (nr *NotificationsRepo) GetAllHostRatings(ctx context.Context) ([]RatingHost, error) {
	ratingsCollection := nr.getHostRatingsCollection()

	cursor, err := ratingsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#42 Failed to find ratings for host: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingHost
	for cursor.Next(ctx) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#43 Failed to decode body: %v", err))
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (nr *NotificationsRepo) GetRatingsByAccommodationID(accommodationID primitive.ObjectID) ([]RatingAccommodation, error) {
	ratingsCollection := nr.getRatingsCollection()
	var ratings []RatingAccommodation

	filter := bson.M{"idAccommodation": accommodationID}

	cursor, err := ratingsCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#45 Failed to find ratings for accommodation with accommodationID: %v", err))
		return ratings, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var rating RatingAccommodation
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#46 Failed to decode body: %v", err))
			return ratings, err
		}
		ratings = append(ratings, rating)
	}
	return ratings, nil
}

func (nr *NotificationsRepo) GetRatingsByHostUsername(username string) ([]RatingHost, error) {
	ratingsCollection := nr.getHostRatingsCollection()
	var ratings []RatingHost

	filter := bson.M{"hostUsername": username}

	cursor, err := ratingsCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#48 Failed to find ratings with host username: %v", err))
		return ratings, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#49 Failed to decode body %v", err))
			return ratings, err
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
		log.Error(fmt.Sprintf("[noti-repo]nr#50: Failed to find ratings wiht hostUsername: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []RatingHost
	for cursor.Next(ctx) {
		var rating RatingHost
		if err := cursor.Decode(&rating); err != nil {
			log.Error(fmt.Sprintf("[noti-repo]nr#51 Failed to decode body: %v", err))
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func (nr *NotificationsRepo) UpdateRatingAccommodationByID(id, idUser primitive.ObjectID, newRate int) error {
	ratingsCollection := nr.getRatingsCollection()

	filter := bson.M{"_id": id}
	var rating RatingAccommodation
	err := ratingsCollection.FindOne(context.Background(), filter).Decode(&rating)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#53 Failed to find rating for accommodation by id: %v", err))
		return err
	}

	if rating.GuestID != idUser {
		log.Error("[noti-repo]nr#54 User does not match the rating creator")
		return errors.New("cannot update rating: user does not match the rating creator")
	}

	update := bson.M{
		"$set": bson.M{
			"rate": newRate,
			"time": time.Now(),
		},
	}

	result, err := ratingsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#54 Failed to upfate rating accommodation: %v", err))
		return err
	}

	if result.ModifiedCount == 0 {
		log.Error("[noti-repo]nr#55 No documents updated")
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
		log.Error(fmt.Sprintf("[noti-repo]nr#56 Failed to find rating for accommodation %v", err))
		return err
	}

	if rating.ID == primitive.NilObjectID {
		log.Error("[noti-repo]nr#57 No rating found with this ID")
		return errors.New("no rating found with this ID")
	}

	if rating.GuestID != idUser {
		log.Error("[noti-repo]nr#58 User did not create this rating")
		return errors.New("user did not create this rating")
	}

	result, err := ratingsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-repo]nr#59 Failed to delete rating: %v", err))
		return err
	}

	if result.DeletedCount == 0 {
		log.Error("[noti-repo]nr#60 No documents deleted")
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
