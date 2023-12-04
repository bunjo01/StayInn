package data

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

type CredentialsRepo struct {
	cli       *mongo.Client
	logger    *log.Logger
	blacklist map[string]struct{}
	once      sync.Once
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

	cr := &CredentialsRepo{
		cli:    client,
		logger: logger,
	}

	cr.once.Do(func() {
		cr.loadBlacklist()
	})

	return cr, nil
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	options := options.FindOne()

	var foundUser Credentials
	err := collection.FindOne(ctx, filter, options).Decode(&foundUser)
	if err != nil {
		cr.logger.Fatal(err.Error())
		return err
	}

	// checks sent password and hashed password in db
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))
	if err != nil {
		return errors.New("invalid password")
	}

	if !foundUser.IsActivated {
		cr.logger.Println("Account not activated!")
		return errors.New("account not activated")
	}

	return nil
}

func (cr *CredentialsRepo) AddCredentials(username, password, email, role string) error {
	collection := cr.getCredentialsCollection()

	newCredentials := Credentials{
		Username:    username,
		Password:    password,
		Email:       email,
		Role:        role,
		IsActivated: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, newCredentials)
	if err != nil {
		cr.logger.Fatal(err.Error())
		return err
	}

	return nil
}

func (cr *CredentialsRepo) AddActivation(activationUUID, username string, confirmed bool) error {
	collection := cr.getActivationCollection()

	currentTime := time.Now()

	newActivation := ActivatioModel{
		ActivationUUID: activationUUID,
		Username:       username,
		Time:           currentTime,
		Confirmed:      confirmed,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, newActivation)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	options := options.FindOne()

	var foundUser Credentials
	err := collection.FindOne(ctx, filter, options).Decode(&foundUser)

	return errors.Is(err, mongo.ErrNoDocuments)
}

func (cr *CredentialsRepo) FindUserByUsername(username string) (NewUser, error) {
	collection := cr.getCredentialsCollection()
	filter := bson.M{"username": username}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		ID:          foundUser.ID,
		Username:    foundUser.Username,
		Password:    "",
		FirstName:   foundUser.FirstName,
		LastName:    foundUser.LastName,
		Email:       foundUser.Email,
		Address:     foundUser.Address,
		Role:        foundUser.Role,
		IsActivated: foundUser.IsActivated,
	}

	return newUser, nil
}

func (cr *CredentialsRepo) GetAllCredentials(ctx context.Context) ([]Credentials, error) {
	collection := cr.getCredentialsCollection()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		cr.logger.Println(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var credentialsList []Credentials

	if err := cursor.All(ctx, &credentialsList); err != nil {
		cr.logger.Println(err)
		return nil, err
	}

	return credentialsList, nil
}

func (cr *CredentialsRepo) ChangeUsername(ctx context.Context, oldUsername, username string) error {
	collection := cr.getCredentialsCollection()

	filter := bson.M{"username": oldUsername}

	update := bson.M{"$set": bson.M{"username": username}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		cr.logger.Println(err)
		return err
	}

	return nil
}

// CheckPassword checks if the given password is contained in the blacklist.
// Returns true if it passes the check, else returns false.
func (cr *CredentialsRepo) CheckPassword(password string) (bool, error) {
	_, found := cr.blacklist[password]
	return !found, nil
}

// Loading blacklist into map for faster lookup
func (cr *CredentialsRepo) loadBlacklist() {
	blacklistFile := "security/blacklist.txt"
	if _, err := os.Stat(blacklistFile); os.IsNotExist(err) {
		log.Printf("Blacklist file not found: %v", err)
		return
	}

	file, err := os.Open(blacklistFile)
	if err != nil {
		log.Printf("Error while opening blacklist file: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	blacklistMap := make(map[string]struct{})

	for scanner.Scan() {
		blacklistMap[strings.TrimSpace(scanner.Text())] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error while scanning blacklist file: %v", err)
		return
	}

	cr.blacklist = blacklistMap
}

// Registers a new user to the system.
// Saves credentials to auth db
func (cr *CredentialsRepo) RegisterUser(username, password, firstName, lastName, email, address, role string) error {
	usernameOK := cr.CheckUsername(username)
	passwordOK, err := cr.CheckPassword(strings.ToLower(password))
	if err != nil {
		return err
	}

	if usernameOK && passwordOK {
		hashedPassword, err := hashPassword(password)
		if err != nil {
			cr.logger.Fatalf("error while hashing password: %v", err)
			return err
		}

		err = cr.AddCredentials(username, hashedPassword, email, role)
		if err != nil {
			cr.logger.Fatalf("error while adding credentials to db: %v", err)
			return err
		}

		activationUUID, err := cr.SendActivationEmail(email)
		if err != nil {
			cr.logger.Println("Failed to send activation email:", err)
		} else {
			cr.logger.Println("Activation email sent successfully with UUID:", activationUUID)
		}

		err = cr.AddActivation(activationUUID, username, false)
		if err != nil {
			cr.logger.Println("Failed to add activation model to collection:", err)
		}

	} else if !usernameOK {
		return UsernameExistsError{Message: "username already exists"}
	} else if !passwordOK {
		return PasswordCheckError{Message: "choose a more secure password"}
	}

	return nil
}

// ChangePassword je metoda koja menja lozinku određenog korisnika
func (ur *CredentialsRepo) ChangePassword(username, oldPassword, newPassword string) error {
	collection := ur.getCredentialsCollection()
	filter := bson.M{"username": username}
	var user Credentials

	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		ur.logger.Println("Old password not correct.")
		return errors.New("old password not correct")
	}

	passwordOK, err := ur.CheckPassword(strings.ToLower(newPassword))
	if err != nil {
		return err
	}

	if passwordOK {
		hashedPassword, err := hashPassword(newPassword)
		if err != nil {
			ur.logger.Fatalf("error while hashing password: %v", err)
			return err
		}
		_, err = collection.UpdateOne(context.Background(), filter, bson.M{"$set": bson.M{"password": hashedPassword}})
		if err != nil {
			return err
		}
	} else if !passwordOK {
		return PasswordCheckError{Message: "choose a more secure password"}
	}

	return nil
}

func (cr *CredentialsRepo) SendActivationEmail(email string) (string, error) {
	// Generiranje UUID-a za aktivaciju
	activationUUID := generateActivationUUID()

	// Slanje e-maila za aktivaciju
	_, err := SendEmail(email, activationUUID, "activation")
	if err != nil {
		return "", err
	}

	return activationUUID, nil
}

func (cr *CredentialsRepo) SendRecoveryEmail(email string) (string, error) {
	recoveryUUID := generateActivationUUID()
	collection := cr.getCredentialsCollection()
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"recoveryUUID": recoveryUUID,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		cr.logger.Println("Failed to insert recoveryUUID in Credentials Collection")
		return "", err
	}
	if result.ModifiedCount == 0 {
		cr.logger.Printf("user with email %s not found", email)
		return "", errors.New("user with the given email was not found")
	}

	_, err = SendEmail(email, recoveryUUID, "recovery")
	if err != nil {
		return "", err
	}

	return recoveryUUID, nil
}

func (cr *CredentialsRepo) ActivateUserAccount(activationUUID string) error {
	collection := cr.getActivationCollection()
	filter := bson.M{"activationUUID": activationUUID}

	update := bson.M{
		"$set": bson.M{
			"confirmed": true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to confirm activation account: %v", err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("activation with activationUUID %s not found", activationUUID)
	}
	var activationModel ActivatioModel

	err = collection.FindOne(ctx, filter).Decode(&activationModel)
	if err != nil {
		return fmt.Errorf("failed to find activation model: %v", err)
	}

	currentTime := time.Now()
	timeDifference := currentTime.Sub(activationModel.Time)

	if timeDifference.Minutes() > 1 {
		return fmt.Errorf("link for activation has expired")
	}

	collection = cr.getCredentialsCollection()

	filter = bson.M{"username": activationModel.Username}
	update = bson.M{
		"$set": bson.M{
			"isActivated": true,
		},
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to activate user account: %v", err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("account with username %s not found", activationModel.Username)
	}
	return nil
}

func (cr *CredentialsRepo) UpdatePasswordWithRecoveryUUID(recoveryUUID, newPassword string) error {
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		cr.logger.Println("Hashovanje lozinke pri resetovanju nije uspelo")
		return err
	}

	collection := cr.getCredentialsCollection()
	filter := bson.M{"recoveryUUID": recoveryUUID}
	update := bson.M{
		"$set": bson.M{
			"password":     hashedPassword,
			"recoveryUUID": "", // Delete recoveryUUID after password change
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		cr.logger.Println("Failed to update password using recoveryUUID")
		return err
	}
	if result.ModifiedCount == 0 {
		cr.logger.Printf("No user found with recoveryUUID: %s", recoveryUUID)
		return errors.New("no user found with recoveryUUID")
	}

	return nil
}

// BCrypt 12 hashing of password.
// Returns hash and nil if successful, else returns empty string and error
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// GenerateToken generates a JWT token with the specified username and role.
func (cr *CredentialsRepo) GenerateToken(username, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"role":     role,
			"exp":      strconv.FormatInt(time.Now().Add(time.Hour*24).Unix(), 10),
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

func (cr *CredentialsRepo) getActivationCollection() *mongo.Collection {
	authDatabase := cr.cli.Database("authDB")
	credentialsCollection := authDatabase.Collection("activation")
	return credentialsCollection
}
