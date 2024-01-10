package data

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	// "strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

type CredentialsRepo struct {
	cli       *mongo.Client
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
func New(ctx context.Context) (*CredentialsRepo, error) {
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
		cli: client,
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
		log.Fatal(fmt.Sprintf("[auth-repo]#ar1 Failed to disconnect: %v", err))
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
		log.Error(fmt.Sprintf("[auth-repo]#ar2 Failed to ping: %v", err))
	}

	// Print available databases
	databases, err := cr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Error(fmt.Sprintf("[auth-repo]#ar3 Failed to list databases: %v", err))
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
		log.Fatal(fmt.Sprintf("[auth-repo]#ar4 Failed to find user %s: %v", username, err))
		return err
	}

	// checks sent password and hashed password in db
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))
	if err != nil {
		log.Warning(fmt.Sprintf("[auth-repo]#ar5 User '%s' entered wrong password: %v", username, err))
		return errors.New("invalid password")
	}

	if !foundUser.IsActivated {
		log.Warning(fmt.Sprintf("[auth-repo]#ar6 Unactivated user '%s' tried to login: %v", username, err))
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
		log.Fatal(fmt.Sprintf("[auth-repo]#ar7 Failed to add credentials: %v", err))
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
		log.Fatal(fmt.Sprintf("[auth-repo]#ar8 Failed to add activation details: %v", err))
		return err
	}

	return nil
}

func (cr *CredentialsRepo) AddRecovery(recoveryUUID, username string, confirmed bool) error {
	collection := cr.getRecoveryCollection()

	currentTime := time.Now()

	newRecovery := RecoveryModel{
		RecoveryUUID: recoveryUUID,
		Time:         currentTime,
		Confirmed:    confirmed,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, newRecovery)
	if err != nil {
		log.Fatal(fmt.Sprintf("[auth-repo]#ar9 Failed to add recovery details: %v", err))
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

func (cr *CredentialsRepo) CheckEmail(email string) bool {
	collection := cr.getCredentialsCollection()
	filter := bson.M{"email": email}

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
		log.Warning(fmt.Sprintf("[auth-repo]#ar10 User '%s' not found: %v", username, err))
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
		log.Error(fmt.Sprintf("[auth-repo]#ar11 Failed to get all credentials: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var credentialsList []Credentials

	if err := cursor.All(ctx, &credentialsList); err != nil {
		log.Error(fmt.Sprintf("[auth-repo]#ar12 Failed to iterate over all credentials: %v", err))
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
		log.Error(fmt.Sprintf("[auth-repo]#ar13 Failed to change username for user '%s': %v", oldUsername, err))
		return err
	}

	return nil
}

func (cr *CredentialsRepo) ChangeEmail(ctx context.Context, oldEmail, email string) error {
	collection := cr.getCredentialsCollection()

	filter := bson.M{"email": oldEmail}

	update := bson.M{"$set": bson.M{"email": email}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error(fmt.Sprintf("[auth-repo]#ar14 Failed to change email for user '%s': %v", oldEmail, err))
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
		log.Error(fmt.Sprintf("[auth-repo]#ar14 Blacklist file not found: %v", err))
		return
	}

	file, err := os.Open(blacklistFile)
	if err != nil {
		log.Error(fmt.Sprintf("[auth-repo]#ar15 Error while opening blacklist file: %v", err))
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	blacklistMap := make(map[string]struct{})

	for scanner.Scan() {
		blacklistMap[strings.TrimSpace(scanner.Text())] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		log.Error(fmt.Sprintf("[auth-repo]#ar16 Error while scanning blacklist file: %v", err))
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
			log.Fatal(fmt.Sprintf("[auth-repo]#ar17 Error while hashing password: %v", err))
			return err
		}

		err = cr.AddCredentials(username, hashedPassword, email, role)
		if err != nil {
			log.Fatal(fmt.Sprintf("[auth-repo]#ar18 Error while adding credentials to DB: %v", err))
			return err
		}

		activationUUID, err := cr.SendActivationEmail(email)
		if err != nil {
			log.Fatal(fmt.Sprintf("[auth-repo]#ar19 Failed to send activation email: %v", err))
		} else {
			log.Info(fmt.Sprintf("[auth-repo]#ar20 Activation email sent with UUID: %s", activationUUID))
		}

		err = cr.AddActivation(activationUUID, username, false)
		if err != nil {
			log.Error(fmt.Sprintf("[auth-repo]#ar21 Error while adding activation model to DB: %v", err))
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
		log.Warning(fmt.Sprintf("[auth-repo]#ar22 User '%s' entered wrong old password when changing it: %v", username, err))
		return errors.New("old password not correct")
	}

	passwordOK, err := ur.CheckPassword(strings.ToLower(newPassword))
	if err != nil {
		return err
	}

	if passwordOK {
		hashedPassword, err := hashPassword(newPassword)
		if err != nil {
			log.Fatal(fmt.Sprintf("[auth-repo]#ar23 Error while hashing password: %v", err))
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
		log.Error(fmt.Sprintf("[auth-repo]#ar24 Failed to insert recoveryUUID in DB: %v", err))
		return "", err
	}
	if result.ModifiedCount == 0 {
		log.Warning(fmt.Sprintf("[auth-repo]#ar25 User with email '%s' not found: %v", email, err))
		return "", errors.New("user with the given email was not found")
	}

	recoveryCollection := cr.getRecoveryCollection()

	recoveryModel := RecoveryModel{
		RecoveryUUID: recoveryUUID,
		Time:         time.Now(),
		Confirmed:    false,
	}
	_, err = recoveryCollection.InsertOne(ctx, recoveryModel)
	if err != nil {
		return "", err
	}

	_, err = SendEmail(email, recoveryUUID, "recovery")
	if err != nil {
		return "", err
	}

	return recoveryUUID, nil
}

func (cr *CredentialsRepo) IsRecoveryLinkExpired(recoveryUUID string) (bool, error) {
	recoveryCollection := cr.getRecoveryCollection()

	filter := bson.M{
		"recoveryUUID": recoveryUUID,
		"confirmed":    false,
	}

	var recoveryData RecoveryModel
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := recoveryCollection.FindOne(ctx, filter).Decode(&recoveryData)
	if err != nil {
		return true, err
	}

	elapsedTime := time.Since(recoveryData.Time)
	if elapsedTime > 1*time.Minute {
		return true, nil // Link has expired
	}

	return false, nil // Link is valid
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

	expired, err := cr.IsRecoveryLinkExpired(recoveryUUID)
	if err != nil {
		return err
	}

	if expired {
		return errors.New("recovery link has expired")
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		log.Fatal(fmt.Sprintf("[auth-repo]#ar27 Error while hashing password: %v", err))
		return errors.New("password hashing failed")
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
		log.Error(fmt.Sprintf("[auth-repo]#ar28 Failed to update password using recoveryUUID '%s': %v", recoveryUUID, err))
		return errors.New("failed to update password using recoveryUUID")
	}
	if result.ModifiedCount == 0 {
		log.Error(fmt.Sprintf("[auth-repo]#ar29 No user found with recoveryUUID '%s': %v", recoveryUUID, err))
		return errors.New("no user found with recoveryUUID")
	}

	return nil
}

func (cr *CredentialsRepo) DeleteUser(ctx context.Context, username string) error {
	collection := cr.getCredentialsCollection()

	filter := bson.M{"username": username}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("[auth-repo]#ar30 Failed to delete user '%s': %v", username, err))
		return err
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
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

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

func (cr *CredentialsRepo) getRecoveryCollection() *mongo.Collection {
	authDatabase := cr.cli.Database("authDB")
	credentialsCollection := authDatabase.Collection("recovery")
	return credentialsCollection
}
