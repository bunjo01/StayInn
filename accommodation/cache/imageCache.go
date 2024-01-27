package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

const (
	defaultExpiration = 30 * time.Second
	longExpiration    = 120 * time.Second
)

type ImageCache struct {
	cli *redis.Client
}

// Constructs Redis Client
func New() *ImageCache {
	redisHost := os.Getenv("REDIS_HOST")
	redistPort := os.Getenv("REDIS_PORT")
	redisAddress := fmt.Sprintf("%s:%s", redisHost, redistPort)

	client := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	return &ImageCache{
		cli: client,
	}
}

// Check connection
func (ic *ImageCache) Ping() {
	val, _ := ic.cli.Ping().Result()
	log.Info(fmt.Sprintf("[acco-cache]acc#1 Redis ping: %s", val))
}

// Set key-value pair with default expiration
func (ic *ImageCache) Post(image *Image) error {
	key := constructKey(image.AccID, image.ID)

	value, err := json.Marshal(image)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-cache]acc#2 Failed to encode image: %v", err))
		return err
	}

	err = ic.cli.Set(key, value, defaultExpiration).Err()

	return err
}

// Get single image by key
func (ic *ImageCache) Get(accID, imageID string) (*Image, error) {
	key := constructKey(accID, imageID)

	val, err := ic.cli.Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	image := &Image{}
	err = json.Unmarshal(val, &image)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-cache]acc#3 Failed to decode image: %v", err))
		return nil, err
	}

	log.Info(fmt.Printf("[acco-cache]acc#4 Cache hit"))
	return image, nil
}

// Get all images for accommodation
func (ic *ImageCache) GetAll(accID string) (Images, error) {
	key := constructKey(accID, "")

	vals, err := ic.cli.Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	var images Images
	err = json.Unmarshal(vals, &images)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-cache]acc#5 Failed to decode image: %v", err))
		return nil, err
	}

	log.Info(fmt.Printf("[acco-cache]acc#6 Cache hit"))
	return images, nil
}

// Post all images for accommodation
func (ic *ImageCache) PostAll(accID string, images Images) error {
	key := constructKey(accID, "")

	value, err := json.Marshal(images)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-cache]acc#7 Failed to encode image: %v", err))
		return err
	}

	err = ic.cli.Set(key, value, longExpiration).Err()

	return err
}

// Check if given key exists
func (ic *ImageCache) Exists(accID, imageID string) (bool, error) {
	key := constructKey(accID, imageID)

	exists, err := ic.cli.Exists(key).Result()
	if err != nil {
		log.Error(fmt.Sprintf("[acco-cache]acc#8 Failed to check key existence: %v", err))
		return false, err
	}

	return exists > 0, nil
}
