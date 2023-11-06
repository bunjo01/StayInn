package data

import (
	"context"
	"errors"
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

type ReservationRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// Constructor
func New(ctx context.Context, logger *log.Logger) (*ReservationRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &ReservationRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect
func (rr *ReservationRepo) Disconnect(ctx context.Context) error {
	err := rr.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Check database connection
func (rr *ReservationRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := rr.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		rr.logger.Println(err)
	}

	// Print available databases
	databases, err := rr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
	}
	fmt.Println(databases)
}

func (rr *ReservationRepo) GetAll() (Reservations, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reservationCollection := rr.getCollection()

	var reservations Reservations
	reserationCursor, err := reservationCollection.Find(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	if err = reserationCursor.All(ctx, &reservations); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return reservations, nil
}

func (rr *ReservationRepo) GetById(id string) (*Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reservationCollection := rr.getCollection()

	var reservation Reservation
	objID, _ := primitive.ObjectIDFromHex(id)
	err := reservationCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&reservation)
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return &reservation, nil
}

func (rr *ReservationRepo) FindAvailablePeriodById(reservationID, periodID string) (*AvailabilityPeriod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservations := rr.getCollection()

	objReservationID, err := primitive.ObjectIDFromHex(reservationID)
	if err != nil {
		return nil, err
	}

	periodObjId, _ := primitive.ObjectIDFromHex(periodID)

	filter := bson.M{
		"_id":                     objReservationID,
		"availabilityPeriods._id": periodObjId,
	}

	var reservation Reservation
	err = reservations.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("Reservation not found")
		}
		return nil, err
	}

	// Find the specific AvailabilityPeriod within the Reservation.
	for _, period := range *reservation.AvailabilityPeriods {
		if period.ID.Hex() == periodID {
			return &period, nil
		}
	}

	return nil, errors.New("AvailabilityPeriod not found in Reservation")
}

func (rr *ReservationRepo) PostReservation(reservation *Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservationCollection := rr.getCollection()

	//	reservation.ID = primitive.NewObjectID()

	if reservation.AvailabilityPeriods == nil {
		reservation.AvailabilityPeriods = &[]AvailabilityPeriod{}
	}

	result, err := reservationCollection.InsertOne(ctx, &reservation)

	if err != nil {
		rr.logger.Println(err)
		return err
	}
	rr.logger.Printf("Reservation Id: %v\n", result.InsertedID)
	return nil
}

func (rr *ReservationRepo) CreateReservation(ctx context.Context, reservationID primitive.ObjectID, periodID primitive.ObjectID) error {
	reservationCollection := rr.getCollection()
	filter := bson.M{"_id": reservationID, "availabilityPeriods._id": periodID}
	update := bson.M{
		"$set": bson.M{"availabilityPeriods.$.isAvailable": true},
	}
	updateOptions := options.Update().SetUpsert(false)

	_, err := reservationCollection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		log.Printf("Failed to reserve availability period: %v\n", err)
		return err
	}

	return nil
}

func (rr *ReservationRepo) AddAvaiablePeriod(id string, period *AvailabilityPeriod) error {
	period.IsAvailable = true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservations := rr.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}

	update := bson.M{"$push": bson.M{
		"availabilityPeriods": period,
	}}
	result, err := reservations.UpdateOne(ctx, filter, update)
	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

func (rr *ReservationRepo) UpdateAvailablePeriod(reservationId string, newPeriod *AvailabilityPeriod) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservations := rr.getCollection()

	objId, _ := primitive.ObjectIDFromHex(reservationId)
	periodObjId := newPeriod.ID.Hex()

	period, _ := rr.FindAvailablePeriodById(reservationId, periodObjId)

	if period != nil {
		if newPeriod.PriceConfiguration.PricePerAccommodation > 0 {
			period.PriceConfiguration.PricePerAccommodation = newPeriod.PriceConfiguration.PricePerAccommodation
		}
		if newPeriod.PriceConfiguration.PricePerGuest > 0 {
			period.PriceConfiguration.PricePerGuest = newPeriod.PriceConfiguration.PricePerGuest
		}
		if !newPeriod.StartDate.IsZero() && newPeriod.StartDate.After(time.Now()) && newPeriod.StartDate.Before(period.EndDate) {
			period.StartDate = newPeriod.StartDate
		}
		if !newPeriod.EndDate.IsZero() && newPeriod.EndDate.After(time.Now()) && newPeriod.EndDate.After(period.StartDate) {
			period.EndDate = newPeriod.EndDate
		}
		period.PriceConfiguration.UsePricePerGuest = newPeriod.PriceConfiguration.UsePricePerGuest

		filter := bson.M{"_id": objId}
		update := bson.M{
			"$set": bson.M{
				"availabilityPeriods": []*AvailabilityPeriod{period},
			},
		}

		result, err := reservations.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Fatal(err)
		}

		rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
		rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)
	}

}

func (rr *ReservationRepo) DeleteById(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservationCollection := rr.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := reservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	rr.logger.Printf("Reservation deleted: %v\n", result.DeletedCount)
	return nil
}

func (rr *ReservationRepo) getCollection() *mongo.Collection {
	patientDatabase := rr.cli.Database("reservationDB")
	patientCollection := patientDatabase.Collection("reservation")
	return patientCollection
}

func (rr *ReservationRepo) ReservePeriod(reservationId string, periodId string) error {
	reservations := rr.getCollection()

	reservationObjID, _ := primitive.ObjectIDFromHex(reservationId)
	availabilityPeriodObjID, _ := primitive.ObjectIDFromHex(periodId)

	filter := bson.M{
		"_id":                     reservationObjID,
		"availabilityPeriods._id": availabilityPeriodObjID,
	}
	update := bson.M{
		"$set": bson.M{
			"availabilityPeriods.$.isAvailable": false,
		},
	}

	result, err := reservations.UpdateOne(context.Background(), filter, update)

	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
