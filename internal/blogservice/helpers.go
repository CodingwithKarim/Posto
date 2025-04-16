package blogservice

import (
	"App/internal/cache"
	"App/internal/utils"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetPostIDAndMode(context *gin.Context) (int, bool, error) {
	// Retrieve postID from URL param
	paramPostID := context.Param(utils.ID)

	// If no ID is provided, return default values
	if paramPostID == "" {
		return 0, false, nil
	}

	// Validate postID if provided
	if id, isValidID := IsValidPostID(paramPostID); isValidID {
		return id, true, nil
	}

	return 0, false, fmt.Errorf("invalid post ID")
}

func GetPageQuery(ctx *gin.Context) int {
	// Parse the query parameter to int & default to 1 if invalid
	page, err := strconv.Atoi(ctx.DefaultQuery(utils.PAGE, utils.DEFAULT_PAGE))

	if err != nil || page < 1 {
		return 1
	}
	if page > utils.BLOG_POST_PAGE_MAX {
		return utils.BLOG_POST_PAGE_MAX
	}

	// Return page if valid
	return page
}

func ValidatePostIDInput(context *gin.Context) (int, error) {
	// Validate post ID from url params
	id, isValidID := IsValidPostID(context.Param(utils.ID))

	if !isValidID {
		return id, fmt.Errorf("invalid post ID")
	}

	// Return post ID if valid
	return id, nil
}

func ValidatePostInputs(title string, isPublic string, content string) error {
	// Validate public status (isPublic)
	if !IsValidPriority(isPublic) {
		return fmt.Errorf("invalid public status: must be 'true' or 'false'")
	}

	// Validate title length
	if !utils.IsValidInputLength(title, utils.BLOG_POST_MIN_LENGTH, utils.BLOG_TITLE_MAX_LENGTH) {
		log.Printf("invalid title length: must be between %d and %d", utils.BLOG_POST_MIN_LENGTH, utils.BLOG_TITLE_MAX_LENGTH)
		return fmt.Errorf("invalid title length: must be between %d and %d", utils.BLOG_POST_MIN_LENGTH, utils.BLOG_TITLE_MAX_LENGTH)
	}

	// Validate content length
	if !utils.IsValidInputLength(content, utils.BLOG_POST_MIN_LENGTH, utils.BLOG_CONTENT_MAX_LENGTH) {
		return fmt.Errorf("invalid content length: must be between %d and %d", utils.BLOG_POST_MIN_LENGTH, utils.BLOG_CONTENT_MAX_LENGTH)
	}

	// Return nil if all validations pass
	return nil
}

func ConvertIsPublicToBool(isPublic string) bool {
	// Convert IsPublic string to bool & return
	isPublicBool, _ := strconv.ParseBool(isPublic)

	return isPublicBool
}

func IsValidPriority(isPublic string) bool {
	// Check if isPublic string is valid
	return isPublic == "true" || isPublic == "false"
}

func IsValidPostID(postID string) (int, bool) {
	// Convert post ID to int
	id, err := strconv.Atoi(postID)

	// Check if ID is clean
	if err != nil || id <= 0 {
		log.Printf("Invalid post ID: %s", postID)
		return -1, false
	}

	// Return ID if valid
	return id, true
}

func IsValidComment(comment string) error {
	if !utils.IsValidInputLength(comment, utils.BLOG_POST_MIN_LENGTH, utils.BLOG_TITLE_MAX_LENGTH) {
		return fmt.Errorf("invalid comment length: must be between %d and %d", utils.BLOG_POST_MIN_LENGTH, utils.BLOG_TITLE_MAX_LENGTH)
	}

	return nil
}

func FormatDate(createdAt []byte) string {
	if createdAt == nil {
		return ""
	}

	// Parse the byte slice to a time.Time object
	timeUTC, err := time.Parse(time.RFC3339, string(createdAt))

	if err != nil {
		log.Printf("Failed to parse time: %v", err)
		return ""
	}

	// Load the America/New_York location for EST/EDT
	loc, err := time.LoadLocation("America/New_York")

	if err != nil {
		log.Printf("Failed to load location: %v", err)
		return ""
	}

	// Convert the time to EST
	timeEST := timeUTC.In(loc)

	// Return a formatted string in EST/EDT
	return timeEST.Format("January 2, 2006 03:04 PM")
}

func EncryptBlogPost(title string, content string, userID int, isPublic bool) (string, string, error) {
	// If post is public, return title and content as is
	if isPublic {
		return title, content, nil
	}

	// If post is private, encrypt the title or leave raw
	encryptedTitle, err := encryptContent(title, userID, isPublic)

	if err != nil {
		log.Printf("Failed to encrypt title: %v", err)
		return "", "", fmt.Errorf("failed to encrypt title")
	}

	// If post is private, encrypt the content or leave raw
	encyptedContent, err := encryptContent(content, userID, isPublic)

	if err != nil {
		log.Printf("Failed to encrypt content: %v", err)
		return "", "", fmt.Errorf("failed to encrypt content")
	}

	// Return encrypted title and content
	return encryptedTitle, encyptedContent, nil
}

func DecryptBlogPost(title string, content string, userID int, isPublic bool) (string, string, error) {
	// If post is public, return title and content as is
	if isPublic {
		return title, content, nil
	}

	// If post is private, decrypt the title or leave raw
	decryptedTitle, err := decryptContent(title, userID, isPublic)

	if err != nil {
		log.Printf("Failed to decrypt title: %v", err)
		return "", "", fmt.Errorf("failed to decrypt title")
	}

	// If post is private, decrypt the content or leave raw
	decryptedContent, err := decryptContent(content, userID, isPublic)

	if err != nil {
		log.Printf("Failed to decrypt content: %v", err)
		return "", "", fmt.Errorf("failed to decrypt content")
	}

	// Return decrypted title and content
	return decryptedTitle, decryptedContent, nil
}

func encryptContent(data string, userID int, isPublic bool) (string, error) {
	// Get the user's encryption key
	key, err := cache.GetUserKey(userID)

	if err != nil {
		log.Printf("Failed to retrieve user key: %v", err)
		return "", fmt.Errorf("failed to retrieve user key")
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)

	if err != nil {
		log.Printf("Failed to create AES cipher: %v", err)
		return "", fmt.Errorf("failed to create AES cipher")
	}

	// Create a GCM instance
	gcm, err := cipher.NewGCM(block)

	if err != nil {
		log.Printf("Failed to create GCM: %v", err)
		return "", fmt.Errorf("failed to create GCM")
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		log.Printf("Failed to generate nonce: %v", err)
		return "", fmt.Errorf("failed to generate nonce")
	}

	// Encrypt the content
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	// Convert the ciphertext to a base64 string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptContent(content string, userID int, isPublic bool) (string, error) {
	// Decode the base64 string back to bytes
	ciphertext, err := base64.StdEncoding.DecodeString(content)

	if err != nil {
		log.Printf("Failed to decode base64 content: %v", err)
		return "", fmt.Errorf("failed to decode encrypted content")
	}

	// Get the user's encryption key
	key, err := cache.GetUserKey(userID)

	if err != nil {
		log.Printf("Failed to retrieve user key: %v", err)
		return "", fmt.Errorf("failed to retrieve user key")
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)

	if err != nil {
		log.Printf("Failed to create AES cipher: %v", err)
		return "", fmt.Errorf("failed to create AES cipher")
	}

	// Create a GCM instance
	gcm, err := cipher.NewGCM(block)

	if err != nil {
		log.Printf("Failed to create GCM: %v", err)
		return "", fmt.Errorf("failed to create GCM")
	}

	// Extract the nonce from the ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		log.Printf("Ciphertext too short, expected at least %d bytes", nonceSize)
		return "", fmt.Errorf("invalid encrypted content")
	}

	// Split the nonce and the actual ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the content
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Printf("Failed to decrypt content: %v", err)
		return "", fmt.Errorf("failed to decrypt content")
	}

	// Return the decrypted string
	return string(plaintext), nil
}

func TruncateString(content string, max int) string {
	if len(content) <= max {
		return content
	}
	return content[:max] + utils.DOTS_STRING
}
