package userservice

import (
	"App/internal/types"
	"App/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetUserStatus(context *gin.Context, username string) (bool, bool) {
	// Cheeck if user is logged in based on context
	user, isValidUser := IsUserLoggedIn(context)

	// if valid login status, check if usernames are identical
	if isValidUser {
		return true, (user.Username == username)
	}

	// return false if user not logged in
	return false, false
}

func IsUserLoggedIn(ctx *gin.Context) (types.User, bool) {
	// Retrieve user from context
	user := GetUserFromContext(ctx)

	// Check if user is valid and return
	return user, (IsValidUser(user))
}

func GetUserFromContext(ctx *gin.Context) types.User {
	// Get User from context and cast to User type
	user, _ := ctx.Value(utils.USER).(types.User)

	// Return user
	return user
}

func IsValidUser(user types.User) bool {
	// Check if username is not empty and ID is > 0
	return user.Username != "" && user.ID > 0
}

func IsValidInputLength(inputString string, min, max int) bool {
	// Check if user info is between 3 and 40 characters long
	return len(inputString) >= min && len(inputString) <= max
}
