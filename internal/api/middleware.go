package api

import (
	"App/internal/types"
	"App/internal/userservice"
	"App/internal/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func RequireAuth(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Authenticate user
		user, isValidUser := authenticateUser(app, context)
		if !isValidUser {
			userservice.HandleAuthenticationError(context)
			return
		}

		// Store authenticated user in context
		context.Set(utils.USER, user)
		context.Next() // Proceed to the next handler
	}
}

func OptionalAuth(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Attempt to authenticate user
		user, isValidUser := authenticateUser(app, context)
		if isValidUser {
			// Store user in context if authenticated
			context.Set(utils.USER, user)
		}

		// Proceed to next handler regardless
		context.Next()
	}
}

func authenticateUser(app *types.App, context *gin.Context) (types.User, bool) {
	// Retrieve session
	session, err := app.SessionStore.Get(context.Request, "cookieSession")
	if err != nil {
		log.Printf("Error accessing session: %+v", err.Error())
		return types.User{}, false
	}

	// Extract user from session
	user, ok := session.Values[utils.USER].(types.User)
	if !ok || !userservice.IsValidUser(user) {
		return types.User{}, false
	}

	// Validate user in the database
	if !userservice.CheckUserExists(user, app.Database) {
		return types.User{}, false
	}

	return user, true
}
