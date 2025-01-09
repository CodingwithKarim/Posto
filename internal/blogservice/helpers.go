package blogservice

import (
	"log"
	"strconv"
	"time"
)

func ConvertIsPublicToBool(isPublic string) bool {
	isPublicBool, _ := strconv.ParseBool(isPublic)

	return isPublicBool
}

func IsValidPriority(priority string) bool {
	return priority == "true" || priority == "false"
}

func IsValidPostID(postID string) (int, bool) {
	id, err := strconv.Atoi(postID)

	if err != nil || id <= 0 {
		log.Printf("Invalid post ID: %s", postID)
		return -1, false
	}

	return id, true
}

func FormatDate(createdAt []byte) string {
	if createdAt == nil {
		return ""
	}

	// Parse the byte slice to a time.Time
	t, err := time.Parse("2006-01-02 15:04:05", string(createdAt))
	if err != nil {
		return "" // return empty string if parsing fails
	}

	return t.Format("January 2, 2006 03:04 PM")
}
