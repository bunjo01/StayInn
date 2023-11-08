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
	// TODO "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type CredentialsRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// Constructor
func New(ctx context.Context, logger *log.Logger) (*CredentialsRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &CredentialsRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect
func (pr *CredentialsRepo) Disconnect(ctx context.Context) error {
	err := pr.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Check database connection
func (pr *CredentialsRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := pr.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		pr.logger.Println(err)
	}

	// Print available databases
	databases, err := pr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		pr.logger.Println(err)
	}
	fmt.Println(databases)
}

// TODO Repo methods

func (cr *CredentialsRepo) ValidateCredentials(username, password string) error {
    filter := bson.M{"username": username}

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    options := options.FindOne()

    var foundUser Credentials
    err := cr.cli.Database("authDB").Collection("credentials").FindOne(ctx, filter, options).Decode(&foundUser)
    if err != nil {
        return err
    }

    if foundUser.Password != password {
        return errors.New("Invalid password")
    }

    return nil
}

func (cr *CredentialsRepo) AddCredentials(username, password string) error {
    newCredentials := Credentials{
        Username: username,
        Password: password,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := cr.cli.Database("authDB").Collection("credentials").InsertOne(ctx, newCredentials)
    if err != nil {
        return err
    }

    return nil
}


