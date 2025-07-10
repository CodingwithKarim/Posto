package api

import (
	"App/internal/cache"
	"App/internal/types"
	"App/internal/userservice"
	"App/internal/utils"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
)

var blockedIPs = make(map[string]time.Time)        // In-memory blocklist
var ipLimiters = make(map[string]*limiter.Limiter) // rate limiter store per IP

func RequireAuth(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Attempt to authenticate user
		user, err := authenticateUser(app, context)

		if err != nil {
			userservice.HandleAuthenticationError(context, err)
			return
		}

		// Check if the user key exists in the cache
		if validKey := cache.HasUserKey(user.ID); !validKey {
			// User key not found in cache, log out the user session
			userservice.LogoutUserSession(context, app.SessionStore)

			userservice.HandleAuthenticationError(context, fmt.Errorf("user key not found in cache for user ID: %d", user.ID))
			return
		}

		// Store user in context to be retrieved for future handlers
		context.Set(utils.USER, user)

		// Proceed to next handler if authentication was successful
		context.Next()
	}
}

func OptionalAuth(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Attempt to authenticate user
		if user, err := authenticateUser(app, context); err == nil {
			// Check if the user key exists in the cache
			if validKey := cache.HasUserKey(user.ID); !validKey {
				// User key not found in cache, log out the user session
				if err := userservice.LogoutUserSession(context, app.SessionStore); err == nil {
					log.Printf("OptionalAuth: encryption key missing for user %d; session logged out", user.ID)
				}

			} else {
				// User key found in cache, store user in context
				context.Set(utils.USER, user)
			}
		}

		// Proceed to next handler regardless
		context.Next()
	}
}

func authenticateUser(app *types.App, context *gin.Context) (types.User, error) {
	// Retrieve session
	session, err := app.SessionStore.Get(context.Request, utils.COOKIE_SESSION)

	if err != nil {
		return types.User{}, fmt.Errorf("error accessing session")
	}

	// Extract user from session
	user, ok := session.Values[utils.USER].(types.User)

	if !ok || !userservice.IsValidUser(user) {
		return types.User{}, fmt.Errorf("error validating user, username: %s", user.Username)
	}

	// Validate user in the database
	if !userservice.CheckUserExists(user, app.Database) {
		return types.User{}, fmt.Errorf("error validating user when checking session stored user")
	}

	return user, nil
}

func BlockSuspiciousIPsAndRateLimit(c *gin.Context) {
	// Grab client IP
	ip := c.ClientIP()

	// Check if the IP is currently blocked
	if blockTime, blocked := blockedIPs[ip]; blocked {
		// Check if the block has expired
		if time.Now().Before(blockTime) {
			// Block the IP from further processing
			c.JSON(403, gin.H{"error": "Access denied. Your IP is blocked."})
			c.Abort()
			return
		} else {
			// If Block expired remove out of blocked list
			delete(blockedIPs, ip)
		}
	}

	// Retrieve or create a limiter for this IP
	lim, exists := ipLimiters[ip]

	// If a limiter doesn't exist create one and set to Client IP
	if !exists {
		// Create a new limiter set for a minute for designated number of requests
		lim = tollbooth.NewLimiter(utils.REQUEST_LIMIT, &limiter.ExpirableOptions{
			DefaultExpirationTTL: time.Minute,
		})

		// Store the limiter for future requests
		ipLimiters[ip] = lim
	}

	// Check rate limit for this IP
	if httpError := tollbooth.LimitByRequest(lim, c.Writer, c.Request); httpError != nil {
		// Log and block the IP if rate limit exceeded
		log.Printf("Suspicious activity detected from IP: %s (rate limit exceeded)", ip)

		// Add to in-memory block list with expiration time
		blockedIPs[ip] = time.Now().Add(utils.EXPIRATION_TIME * time.Hour)

		c.JSON(httpError.StatusCode, gin.H{"error": "Access denied. Rate limit exceeded."})
		c.Abort()
		return
	}

	// Proceed with the request if within rate limits
	c.Next()
}

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// if you want to allow all, skip this check and use "*"
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Vary", "Origin") // tell caches that we vary on origin
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Handle preflight
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
