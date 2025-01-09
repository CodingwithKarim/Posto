package userservice

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"App/internal/types"
	"App/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(username string, password string, database *sql.DB) error {
	// Declare variable to check if username exists
	var exists bool

	// Execute SQL query & store result in exists variable
	if err := database.QueryRow(utils.UserExistsQuery, username).Scan(&exists); err != nil {
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

	// Execute sql query, passing the username & hashed password
	_, err = database.Exec(utils.InsertUserQuery, username, passwordHash)

	if err != nil {
		log.Println("Error inserting new user into database:", err)
		return fmt.Errorf("failed to insert new user")
	}

	// Return nil if no errors occurred
	return nil
}

func VerifyUserCredentials(username, password string, database *sql.DB) (types.User, error) {
	// Declare variables to store id & password hash from SQL query
	var id int
	var passwordHash []byte

	// Run SQL query against the database & return the SQL row
	row := database.QueryRow(utils.GetUserCredentialsQuery, username)

	// Scan the row data into id and passwordHash
	if err := row.Scan(&id, &passwordHash); err != nil {
		log.Println("Error fetching user from database:", err)
		return types.User{}, fmt.Errorf("user not found")
	}

	// Compare password from user with the hashed password in the database
	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(password)); err != nil {
		log.Println("Error comparing password:")
		return types.User{}, fmt.Errorf("invalid password credentials")
	}

	// Return user struct if authentication is successful
	return types.User{
		ID:       id,
		Username: username,
	}, nil
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

func SaveUserSession(context *gin.Context, app *types.App, user types.User) error {
	// Retrieve session data from the request
	session, err := app.SessionStore.Get(context.Request, "cookieSession")
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
	session, err := store.Get(context.Request, "cookieSession")

	if err != nil {
		log.Printf("Failed to retrieve session data: %v", err)
		return fmt.Errorf("failed to retrieve session data")
	}

	// Expire the session by setting MaxAge to -1
	session.Options.MaxAge = -1
	if err := session.Save(context.Request, context.Writer); err != nil {
		log.Printf("Failed to save session data during logout: %v", err)
		return fmt.Errorf("failed to save session data")
	}

	return nil
}

func ValidateAuthInputLength(username, password string) error {
	// Validate username and password length
	if !utils.IsValidInputLength(username, utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH) {
		return fmt.Errorf("username must be between %d and %d characters", utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH)
	}

	if !utils.IsValidInputLength(password, utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH) {
		return fmt.Errorf("password must be between %d and %d characters", utils.AUTH_MIN_LENGTH, utils.AUTH_MAX_LENGTH)
	}

	return nil
}

func HandleAuthenticationError(context *gin.Context) {
	// Redirect the user to the login page
	context.Redirect(http.StatusFound, "/login")

	// Abort further processing of the request
	context.Abort()
}
