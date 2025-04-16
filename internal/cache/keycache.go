package cache

import (
	"App/internal/utils"
	"fmt"
	"log"
	"sync"

	"golang.org/x/crypto/argon2"
)

var userKeyCache = &sync.Map{}

func DeriveAndCacheUserKey(userID int, password string, salt []byte) error {
	if len(salt) != 16 {
		return fmt.Errorf("salt must be exactly 16 bytes")
	}

	if len(password) < utils.AUTH_MIN_LENGTH || len(password) > utils.AUTH_MAX_LENGTH {
		return fmt.Errorf("password length must be between %d and %d characters",
			utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH)
	}

	// Derive a key using Argon2 using the provided password and salt
	key := argon2.IDKey([]byte(password), salt, utils.ArgonTime, utils.ArgonMemory, utils.ArgonThreads, utils.ArgonKeyLen)

	// Store the derived key system cache
	CacheUserKey(userID, key)

	// return no error if the key is successfully cached
	return nil
}

func CacheUserKey(userID int, key []byte) {
	// Store the derived key in the cache (ID : Key)
	userKeyCache.Store(userID, key)
}

func GetUserKey(userID int) ([]byte, error) {
	// Retrieve the key from the cache using the user ID
	value, ok := userKeyCache.Load(userID)

	if !ok {
		return nil, fmt.Errorf("user key not found in cache")
	}

	// Type assertion with safety check
	key, ok := value.([]byte)

	if !ok {
		log.Printf("Invalid key type found in cache for user ID: %d", userID)
		return nil, fmt.Errorf("invalid key type found in cache")
	}

	// Return the key
	return key, nil
}

func HasUserKey(userID int) bool {
	// Check if the user key exists in the cache
	value, ok := userKeyCache.Load(userID)

	if !ok {
		return false
	}

	// Check if the value is of type []byte
	_, valid := value.([]byte)

	return valid
}

func RemoveUserKey(userID int) {
	userKeyCache.Delete(userID)
}
