package utils

import (
	"App/internal/types"
	"strings"

	"github.com/gin-gonic/gin"
)

func IsValidInputLength(inputString string, min, max int) bool {
	// Check if user info is between 3 and 40 characters long
	return len(inputString) >= min && len(inputString) <= max
}

func CapitalizeFirstLetter(s string) string {
	// Check if string is invaild
	if len(s) == 0 {
		return s
	}

	// Concat the first uppercased letter with rest of string & return
	return strings.ToUpper(string(s[0])) + s[1:]
}

func SendErrorResponse(context *gin.Context, statusCode int, errorMessage string) {
	// Render error page with passed in status code & error message
	context.HTML(statusCode, ERROR_PAGE, types.ErrorPageData{
		StatusCode:   statusCode,
		ErrorMessage: errorMessage,
	})
}

func TruncateChars(s string, maxRunes int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes]) + "…"
}
