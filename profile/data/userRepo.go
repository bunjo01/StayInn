package data

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type UserRepo struct {
	cli *mongo.Client
}

type UsernameExistsError struct {
	Message string
}

func (e UsernameExistsError) Error() string {
	return e.Message
}

// Constructor
func New(ctx context.Context) (*UserRepo, error) {
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
		cli: client,
	}, nil
}

// Disconnect
func (ur *UserRepo) Disconnect(ctx context.Context) error {
	err := ur.cli.Disconnect(ctx)
	if err != nil {
		log.Fatal(fmt.Sprintf("[prof-repo]#pr1 Failed to disconnect: %v", err))
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
		log.Error(fmt.Sprintf("[prof-repo]#pr2 Failed to ping: %v", err))
	}

	// Print available databases
	databases, err := ur.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Error(fmt.Sprintf("[prof-repo]#pr3 Failed to list databases: %v", err))
	}
	fmt.Println(databases)
}

// Repo methods

func (ur *UserRepo) CreateProfileDetails(ctx context.Context, user *NewUser) error {
	collection := ur.getUserCollection()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(fmt.Sprintf("[prof-repo]#pr4 Failed to create profile: %v", err))
		return err
	}

	return nil
}

func (ur *UserRepo) GetAllUsers(ctx context.Context) ([]*NewUser, error) {
	collection := ur.getUserCollection()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(fmt.Sprintf("[prof-repo]#pr5 Failed to get all users: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*NewUser
	if err := cursor.All(ctx, &users); err != nil {
		log.Error(fmt.Sprintf("[prof-repo]#pr6 Failed to iterate over all users: %v", err))
		return nil, err
	}

	return users, nil
}

func (ur *UserRepo) GetUser(ctx context.Context, username string) (*NewUser, error) {
	collection := ur.getUserCollection()

	var user NewUser
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		log.Fatal(fmt.Sprintf("[prof-repo]#pr7 Failed to get user by username '%s': %v", username, err))
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepo) GetUserById(ctx context.Context, id primitive.ObjectID) (*NewUser, error) {
	collection := ur.getUserCollection()
	var user NewUser
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		log.Fatal(fmt.Sprintf("[prof-repo]#pr8 Failed to get user by id '%s': %v", id, err))
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

func (ur *UserRepo) UpdateUser(ctx context.Context, username string, user *NewUser, oldEmail string) error {
	usernameAvailable, err := ur.CheckUsernameAvailability(ctx, user.Username)
	if err != nil {
		log.Fatal(fmt.Sprintf("[prof-repo]#pr9 Failed to check username availability: %v", err))
		return err
	}

	if !usernameAvailable && username != user.Username {
		// return fmt.Errorf("username %s is already taken", user.Username)
		log.Error(fmt.Sprintf("[prof-repo]#pr10 Username '%s' is already taken", user.Username))
		return err
	}

	if oldEmail != user.Email {
		emailAvailable, err := ur.CheckEmailAvailability(ctx, user.Email)
		if err != nil {
			log.Error(fmt.Sprintf("[prof-repo]#pr11 Error checking email availability: %v", err))
			return err
		}

		if !emailAvailable {
			log.Error(fmt.Sprintf("[prof-repo]#pr12 Email '%s' is already taken", user.Email))
			return err
		}
	}

	collection := ur.getUserCollection()

	filter := bson.M{"username": username}
	update := bson.M{"$set": user}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-repo]#pr13 Error updating user in profile service: %v", err))
		return err
	}

	return nil
}

func (ur *UserRepo) DeleteUser(ctx context.Context, username string) error {
	collection := ur.getUserCollection()

	filter := bson.M{"username": username}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-repo]#pr14 Failed to delete user '%s': %v", username, err))
		return err
	}

	return nil
}

func (ur *UserRepo) getUserCollection() *mongo.Collection {
	profileDatabase := ur.cli.Database("profileDB")
	usersCollection := profileDatabase.Collection("users")
	return usersCollection
}
