package userservice

import (
	"App/internal/types"
	"App/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetUserStatus(ctx *gin.Context, username string) (bool, bool) {
	user, isValidUser := IsUserLoggedIn(ctx)

	if isValidUser {
		return true, (user.Username == username)
	}

	return false, false
}

func IsUserLoggedIn(ctx *gin.Context) (types.User, bool) {
	user := GetUserFromContext(ctx)

	return user, (IsValidUser(user))
}

func GetUserFromContext(ctx *gin.Context) types.User {
	user, _ := ctx.Value(utils.USER).(types.User)
	return user
}

func IsValidUser(user types.User) bool {
	return user.Username != "" && user.ID > 0
}

func IsValidInputLength(inputString string, min, max int) bool {
	// Check if user info is between 3 and 40 characters long
	return len(inputString) >= min && len(inputString) <= max
}
