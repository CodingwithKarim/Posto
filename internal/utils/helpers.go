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
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(string(s[0])) + s[1:]
}

func SendErrorResponse(context *gin.Context, statusCode int, errorMessage string) {
	context.HTML(statusCode, ERROR_PAGE, types.ErrorPageData{
		StatusCode:   statusCode,
		ErrorMessage: errorMessage,
	})
}
