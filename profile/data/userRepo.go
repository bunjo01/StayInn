package data

import (
	"context"
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

type UserRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// Constructor
func New(ctx context.Context, logger *log.Logger) (*UserRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &UserRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect
func (ur *UserRepo) Disconnect(ctx context.Context) error {
	err := ur.cli.Disconnect(ctx)
	if err != nil {
		ur.logger.Fatal(err.Error())
		return err
	}
	return nil
}

// Check database connection
func (ur *UserRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := ur.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		ur.logger.Println(err)
	}

	// Print available databases
	databases, err := ur.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ur.logger.Println(err)
	}
	fmt.Println(databases)
}

// Repo methods

func (ur *UserRepo) CreateProfileDetails(ctx context.Context, user *NewUser) error {
    collection := ur.getUserCollection()

    _, err := collection.InsertOne(ctx, user)
    if err != nil {
        ur.logger.Println(err)
        return err
    }

    return nil
}

func (ur *UserRepo) GetAllUsers(ctx context.Context) ([]*NewUser, error) {
	collection := ur.getUserCollection()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*NewUser
	if err := cursor.All(ctx, &users); err != nil {
		ur.logger.Println(err)
		return nil, err
	}

	return users, nil
}

func (ur *UserRepo) GetUser(ctx context.Context, id primitive.ObjectID) (*NewUser, error) {
	collection := ur.getUserCollection()

	var user NewUser
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepo) UpdateUser(ctx context.Context, user *NewUser) error {
	collection := ur.getUserCollection()

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ur.logger.Println(err)
		return err
	}

	return nil
}

func (ur *UserRepo) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	collection := ur.getUserCollection()

	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		ur.logger.Println(err)
		return err
	}

	return nil
}

func (ur *UserRepo) getUserCollection() *mongo.Collection {
	profileDatabase := ur.cli.Database("profileDB")
    usersCollection := profileDatabase.Collection("users")
	return usersCollection
}