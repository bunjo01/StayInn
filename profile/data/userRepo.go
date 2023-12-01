package data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"net/http"
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

func (ur *UserRepo) UpdateUser(ctx context.Context, user *NewUser) error {

	usernameAvailable, err := ur.CheckUsernameAvailability(ctx, user.Username)
    if err != nil {
        ur.logger.Println("Error checking username availability:", err)
        return err
    }

    if !usernameAvailable {
        return fmt.Errorf("Username %s is already taken", user.Username)
    }

    collection := ur.getUserCollection()

    filter := bson.M{"_id": user.ID}
    update := bson.M{"$set": user}

    _, err = collection.UpdateOne(ctx, filter, update)
    if err != nil {
        ur.logger.Println("Error updating user in profile service:", err)
        return err
    }

    ur.logger.Printf("User updated in profile service")

    err = ur.passUsernameToAuthService(user.Email, user.Username)
    if err != nil {
        ur.logger.Println("Error passing username to auth service:", err)
        return err
    }

    return nil
}

func (ur *UserRepo) passUsernameToAuthService(email, username string) error {
    credentialsServiceURL := os.Getenv("AUTH_SERVICE_URI")

    reqBody := map[string]string{"username": username}
    requestBody, err := json.Marshal(reqBody)
    if err != nil {
        ur.logger.Println("Error marshaling request body:", err)
        return err
    }

    req, err := http.NewRequest("PUT", credentialsServiceURL+ "/update-username" + "/" + email + "/" + username, bytes.NewBuffer(requestBody))
    if err != nil {
        ur.logger.Println("Error creating HTTP PUT request:", err)
        return fmt.Errorf("failed to create HTTP PUT request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        ur.logger.Println("Error making HTTP PUT request to auth service:", err)
        return fmt.Errorf("HTTP PUT request to credentials service failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        ur.logger.Printf("HTTP PUT request to auth service failed with status: %d\n", resp.StatusCode)
        return fmt.Errorf("HTTP PUT request to credentials service failed with status: %d", resp.StatusCode)
    }

    ur.logger.Println("HTTP PUT request to auth service successful")

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
