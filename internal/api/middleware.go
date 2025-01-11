package api

import (
	"App/internal/types"
	"App/internal/userservice"
	"App/internal/utils"
	"log"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
)

var blockedIPs = make(map[string]time.Time)        // In-memory blocklist
var ipLimiters = make(map[string]*limiter.Limiter) // Store rate limiters per IP

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
		log.Println("In optional auth, isvalid user is", isValidUser)
		if isValidUser {
			// Store user in context if authenticated
			context.Set(utils.USER, user)
		}

		// Proceed to next handler regardless
		context.Next()
	}
}

func BlockSuspiciousIPsAndRateLimit(c *gin.Context) {
	// Grab client IP
	ip := c.ClientIP()

	// Check if the IP is blocked indefinitely
	if blockTime, blocked := blockedIPs[ip]; blocked {
		// Check if the block has expired
		if time.Now().Before(blockTime) {
			// Block is still valid
			c.JSON(403, gin.H{"error": "Access denied. Your IP is blocked."})
			c.Abort()
			return
		} else {
			// Block has expired, remove from the block list
			delete(blockedIPs, ip)
		}
	}

	// Retrieve or create a limiter for this IP
	lim, exists := ipLimiters[ip]
	if !exists {
		// Create a new limiter set for a minute
		lim = tollbooth.NewLimiter(utils.REQUEST_LIMIT, &limiter.ExpirableOptions{
			DefaultExpirationTTL: time.Minute,
		})

		// Store the limiter for future requests
		ipLimiters[ip] = lim
	}

	// Check rate limit for this IP
	if httpError := tollbooth.LimitByRequest(lim, c.Writer, c.Request); httpError != nil {
		// Log and block the IP indefinitely on rate limit exceed
		log.Printf("Suspicious activity detected from IP: %s (rate limit exceeded)", ip)

		// Add to in-memory block list with expiration time (e.g., 24 hours from now)
		blockedIPs[ip] = time.Now().Add(utils.EXPIRATION_TIME)

		c.JSON(httpError.StatusCode, gin.H{"error": "Access denied. Rate limit exceeded."})
		c.Abort()
		return
	}

	// Proceed with the request if within rate limits
	c.Next()
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
		log.Println("Not a valid user", user.ID, user.Username)
		return types.User{}, false
	}

	// Validate user in the database
	if !userservice.CheckUserExists(user, app.Database) {
		log.Println("User doesnt exist", user.Username, user.ID)
		return types.User{}, false
	}

	return user, true
}
