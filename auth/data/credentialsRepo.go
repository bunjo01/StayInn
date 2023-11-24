package data

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// NoSQL: module containing Mongo api client
	"go.mongodb.org/mongo-driver/bson"
	// TODO "go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type CredentialsRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// Custom errors for better handling

type UsernameExistsError struct {
	Message string
}

func (e UsernameExistsError) Error() string {
	return e.Message
}

type PasswordCheckError struct {
	Message string
}

func (e PasswordCheckError) Error() string {
	return e.Message
}

const jwtSecret = "stayinn_secret"

var secretKey = []byte("stayinn_secret")

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
func (cr *CredentialsRepo) Disconnect(ctx context.Context) error {
	err := cr.cli.Disconnect(ctx)
	if err != nil {
		cr.logger.Fatal(err.Error())
		return err
	}
	return nil
}

// Check database connection
func (cr *CredentialsRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := cr.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		cr.logger.Println(err)
	}

	// Print available databases
	databases, err := cr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		cr.logger.Println(err)
	}
	fmt.Println(databases)
}

// TODO Repo methods

func (cr *CredentialsRepo) ValidateCredentials(username, password string) error {
	collection := cr.getCredentialsCollection()
	filter := bson.M{"username": username}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	options := options.FindOne()

	var foundUser Credentials
	err := collection.FindOne(ctx, filter, options).Decode(&foundUser)
	if err != nil {
		cr.logger.Fatal(err.Error())
		return err
	}

	if foundUser.Password != password {
		return errors.New("invalid password")
	}

	return nil
}

func (cr *CredentialsRepo) AddCredentials(username, password, email string) error {
	collection := cr.getCredentialsCollection()

	newCredentials := Credentials{
		Username: username,
		Password: password,
		Email:    email,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, newCredentials)
	if err != nil {
		cr.logger.Fatal(err.Error())
		return err
	}

	return nil
}

// Checks if username already exists in database.
// Returns true if username is unique, else returns false
func (cr *CredentialsRepo) CheckUsername(username string) bool {
	collection := cr.getCredentialsCollection()
	filter := bson.M{"username": username}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	options := options.FindOne()

	var foundUser Credentials
	err := collection.FindOne(ctx, filter, options).Decode(&foundUser)

	return errors.Is(err, mongo.ErrNoDocuments)
}

func (cr *CredentialsRepo) FindUserByUsername(username string) (NewUser, error) {
	collection := cr.getCredentialsCollection()
	filter := bson.M{"username": username}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	options := options.FindOne()
	var foundUser NewUser
	err := collection.FindOne(ctx, filter, options).Decode(&foundUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// User not found
			return NewUser{}, errors.New("user not found")
		}
		// Other error
		return NewUser{}, err
	}

	// Convert the found user to the NewUser type
	newUser := NewUser{
		ID:        foundUser.ID,
		Username:  foundUser.Username,
		Password:  "",
		FirstName: foundUser.FirstName,
		LastName:  foundUser.LastName,
		Email:     foundUser.Email,
		Address:   foundUser.Address,
		Role:      foundUser.Role,
	}

	return newUser, nil
}

// Checks if password is contained in the blacklist.
// Returns true if it passes the check, else returns false
func (cr *CredentialsRepo) CheckPassword(password string) (bool, error) {
	file, err := os.Open("security/blacklist.txt")
	if err != nil {
		log.Printf("error while opening blacklist.txt: %v", err)
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == password {
			return false, nil
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("error while scanning blacklist.txt: %v", err)
		return false, err
	}

	return true, nil
}

// Registers a new user to the system.
// Saves credentials to auth service and passes rest of info to profile service
func (cr *CredentialsRepo) RegisterUser(username, password, firstName, lastName, email, address string, role Role) error {
	usernameOK := cr.CheckUsername(username)
	passwordOK, err := cr.CheckPassword(strings.ToLower(password))
	if err != nil {
		return err
	}

	if usernameOK && passwordOK {
		err := cr.AddCredentials(username, password, email)
		if err != nil {
			cr.logger.Fatal(err.Error())
			return err
		}

		// pass info to profile service

		// err = cr.passInfoToProfileService(username, firstName, lastName, email, address)
		// if err != nil {
		//     return err
		// }

	} else if !usernameOK {
		return UsernameExistsError{Message: "username already exists"}
	} else if !passwordOK {
		return PasswordCheckError{Message: "choose a more secure password"}
	}

	return nil
}

func (cr *CredentialsRepo) passInfoToProfileService(username, firstName, lastName, email, address string) error {
	newUser := NewUser{
		Username:  username,
		Password:  "",
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Address:   address,
	}

	httpClient := &http.Client{}

	profileServiceURL := "http://profile_service:8083"

	requestBody, err := json.Marshal(newUser)
	if err != nil {
		return fmt.Errorf("Failed to marshal user data: %v", err)
	}

	log.Printf("Sending HTTP POST request to %s with payload: %s", profileServiceURL, requestBody)

	resp, err := httpClient.Post(profileServiceURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("HTTP POST request to profile service failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP POST request to profile service failed with status: %d", resp.StatusCode)
	}

	log.Println("HTTP POST request successful")

	return nil
}

// GenerateToken generates a JWT token with the specified username and role.
func (cr *CredentialsRepo) GenerateToken(username string) (string, error) {
	//userRole, err := cr.FindUserByUsername(username)
	//
	//if err != nil {
	//	if errors.Is(err, mongo.ErrNoDocuments) {
	//		return "", errors.New("user not found")
	//	}
	//	return "", err
	//}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"role":     "HOST",
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (cr *CredentialsRepo) getCredentialsCollection() *mongo.Collection {
	authDatabase := cr.cli.Database("authDB")
	credentialsCollection := authDatabase.Collection("credentials")
	return credentialsCollection
}
