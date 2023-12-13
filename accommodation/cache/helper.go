package cache

import "fmt"

const (
	cacheImage                  = "accommodation:%s:image:%s"
	cacheImagesForAccommodation = "accommodation:%s"
	cacheAll                    = "accommodation"
)

// Constructs key for Redis image cache.
// If accommodation ID and image ID are specified, returns key for one image.
// If only accommodation ID is specified, returns key for all images for accomodation.
// If both params are empty string, returns key for all images
func constructKey(accID, imageID string) string {
	if accID != "" && imageID != "" {
		return fmt.Sprintf(cacheImage, accID, imageID)
	} else if accID != "" && imageID == "" {
		return fmt.Sprintf(cacheImagesForAccommodation, accID)
	}
	return cacheAll
}
