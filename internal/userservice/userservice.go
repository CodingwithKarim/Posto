package userservice

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"App/internal/cache"
	"App/internal/types"
	"App/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUserAndSaveSession(username string, password string, context *gin.Context, app *types.App) error {
	// Declare variable to check if username exists
	var exists bool

	// Execute SQL query & store result in exists variable
	if err := app.Database.QueryRow(utils.UserExistsQuery, username).Scan(&exists); err != nil {
		log.Println("Error checking if username exists:", err)
		return fmt.Errorf("error checking username availability")
	}

	// If username exists return an error
	if exists {
		return fmt.Errorf("username already exists")
	}

	// Hash the provided password using bcrypt package
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	if err != nil {
		log.Println("Failed to generate password hash:", err)
		return fmt.Errorf("failed to hash password")
	}

	// Generate a random encryption salt
	encryptionSalt := make([]byte, 16)

	if _, err := rand.Read(encryptionSalt); err != nil {
		log.Println("Failed to generate encryption salt:", err)
		return fmt.Errorf("failed to generate encryption salt")
	}

	// Execute sql query, passing the username & hashed password
	result, err := app.Database.Exec(utils.InsertUserQuery, username, passwordHash, encryptionSalt)

	if err != nil {
		log.Println("Error inserting new user into database:", err)
		return fmt.Errorf("failed to insert new user")
	}

	// Get the last inserted ID
	userID, err := result.LastInsertId()

	if err != nil {
		log.Println("Error getting last inserted ID:", err)
		return fmt.Errorf("failed to get last inserted ID")
	}

	id := int(userID)

	// Cache the user key using the derived salt
	cache.DeriveAndCacheUserKey(id, password, encryptionSalt)

	// Save the user session using the session store
	if err := SaveUserSession(context, app.SessionStore, &types.User{
		ID:       id,
		Username: username,
	}); err != nil {
		log.Println("Failed to save user session:", err)
		return fmt.Errorf("failed to save session after registration")
	}

	// Return nil if user registration & session saving is successful
	return nil
}

func VerifyUserCredentialsAndSaveSession(username, password string, context *gin.Context, app *types.App) error {
	// Declare variables to store id & password hash from SQL query
	var id int
	var passwordHash []byte
	var encryptionSalt []byte

	// Run SQL query against the database & return the SQL row
	row := app.Database.QueryRow(utils.GetUserCredentialsQuery, username)

	// Scan the row data into id and passwordHash
	if err := row.Scan(&id, &passwordHash, &encryptionSalt); err != nil {
		log.Println("Error fetching user from database:", err)
		return fmt.Errorf("user not found")
	}

	// Compare password from user with the hashed password in the database
	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(password)); err != nil {
		log.Println("Error comparing password:")
		return fmt.Errorf("invalid password credentials")
	}

	cache.DeriveAndCacheUserKey(id, password, encryptionSalt)

	if err := SaveUserSession(context, app.SessionStore, &types.User{
		ID:       id,
		Username: username,
	}); err != nil {
		log.Println("Failed to save user session:", err)
		return fmt.Errorf("failed to save session after registration")
	}

	return nil
}

func CheckUserExists(user types.User, database *sql.DB) bool {
	// Execute the query using the constant and check if the user exists
	var exists bool
	if err := database.QueryRow(utils.UserExistsQuery, user.Username).Scan(&exists); err != nil {
		// If no rows are found, we return false (normal case)
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		// Log any other database-related errors
		log.Printf("Database query error: %v", err)
		return false
	}

	// Return the result
	return exists
}

func SaveUserSession(context *gin.Context, cookieStore *sessions.CookieStore, user *types.User) error {
	// Retrieve session data from the request
	session, err := cookieStore.Get(context.Request, utils.COOKIE_SESSION)
	if err != nil {
		log.Printf("Failed to retrieve session data for user: %s, Error: %v", user.Username, err)
		return fmt.Errorf("failed to retrieve session data: %w", err)
	}

	// Store user in the session
	session.Values[utils.USER] = user

	// Save session data to ensure persistence
	if err := session.Save(context.Request, context.Writer); err != nil {
		log.Printf("Failed to save session data for user: %s, Error: %v", user.Username, err)
		return fmt.Errorf("failed to save session data: %w", err)
	}

	return nil
}

func LogoutUserSession(context *gin.Context, store *sessions.CookieStore) error {
	// Retrieve session data from the store
	session, err := store.Get(context.Request, utils.COOKIE_SESSION)

	if err != nil {
		log.Printf("Failed to retrieve session data: %v", err)
		return fmt.Errorf("failed to retrieve session data")
	}

	// Expire the session by setting MaxAge to -1
	session.Options.MaxAge = -1

	// Save session data to ensure persistence
	if err := session.Save(context.Request, context.Writer); err != nil {
		log.Printf("Failed to save session data during logout: %v", err)
		return fmt.Errorf("failed to save session data")
	}

	return nil
}

func ValidateAuthInputLength(username, password string) error {
	// Validate username length
	if !utils.IsValidInputLength(username, utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH) {
		return fmt.Errorf("username must be between %d and %d characters", utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH)
	}

	// Validate password length
	if !utils.IsValidInputLength(password, utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH) {
		return fmt.Errorf("password must be between %d and %d characters", utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH)
	}

	// Return nil if username & password length validation passed
	return nil
}

func HandleAuthenticationError(context *gin.Context, err error) {
	// Log error
	log.Println(err.Error())

	// Redirect the user to the login page
	context.Redirect(http.StatusFound, "/login")

	// Abort further processing of the request
	context.Abort()
}
