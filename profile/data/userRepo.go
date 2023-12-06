package data

import (
	"context"
	"errors"
	"fmt"
	"log"

	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type UserRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

type UsernameExistsError struct {
	Message string
}

func (e UsernameExistsError) Error() string {
	return e.Message
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func (ur *UserRepo) GetUser(ctx context.Context, username string) (*NewUser, error) {
	collection := ur.getUserCollection()

	var user NewUser
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepo) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	collection := ur.getUserCollection()
	filter := bson.M{"username": username}

	err := collection.FindOne(ctx, filter).Err()

	return errors.Is(err, mongo.ErrNoDocuments), nil
}

func (ur *UserRepo) CheckEmailAvailability(ctx context.Context, email string) (bool, error) {
	collection := ur.getUserCollection()
	filter := bson.M{"email": email}

	err := collection.FindOne(ctx, filter).Err()

	return errors.Is(err, mongo.ErrNoDocuments), nil
}

func (ur *UserRepo) UpdateUser(ctx context.Context, username string, user *NewUser) error {
	usernameAvailable, err := ur.CheckUsernameAvailability(ctx, user.Username)
	if err != nil {
		ur.logger.Println("Error checking username availability:", err)
		return err
	}

	if !usernameAvailable && username != user.Username {
		return fmt.Errorf("username %s is already taken", user.Username)
	}

	emailAvailable, err := ur.CheckEmailAvailability(ctx, user.Email)
    if err != nil {
        ur.logger.Println("Error checking email availability:", err)
        return err
    }

    if !emailAvailable {
        return fmt.Errorf("email %s is already taken", user.Email)
    }

	collection := ur.getUserCollection()

	filter := bson.M{"username": username}
	update := bson.M{"$set": user}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ur.logger.Println("Error updating user in profile service:", err)
		return err
	}

	ur.logger.Printf("User updated in profile service")

	return nil
}

func (ur *UserRepo) DeleteUser(ctx context.Context, username string) error {
	collection := ur.getUserCollection()

	filter := bson.M{"username": username}
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
