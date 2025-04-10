package blogservice

import (
	"App/internal/utils"
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

	// Parse the byte slice to a time.Time object in UTC
	timeUTC, err := time.Parse("2006-01-02 15:04:05", string(createdAt))
	if err != nil {
		return ""
	}

	// Load the America/New_York location for EST/EDT
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return ""
	}

	// Convert the time to EST
	timeEST := timeUTC.In(loc)

	// Return a formatted string in EST/EDT
	return timeEST.Format("January 2, 2006 03:04 PM")
}
