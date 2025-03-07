package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"

	"App/internal/api"
	"App/internal/types"
	"database/sql"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file in local environment
	if err := godotenv.Load("/env/posto.env"); err != nil {
		log.Fatal("posto.env file not found")
	}
}

func main() {
	// Get required system environment variables
	appPort := os.Getenv("APP_PORT")
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlPort := os.Getenv("MYSQL_PORT")
	mysqlDB := os.Getenv("MYSQL_DB")
	cookieStoreKey := os.Getenv("SESSION_SECRET")

	// Ensure all necessary environment variables are present
	if mysqlUser == "" || mysqlPassword == "" || mysqlHost == "" || mysqlPort == "" || mysqlDB == "" {
		log.Fatal("Required environment variables are missing.")
	}

	// Construct the connection string for MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDB)
	database, err := sql.Open("mysql", dsn)

	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}

	// Ensure the database connection is valid
	if err := database.Ping(); err != nil {
		log.Fatal("Error pinging the database:", err)
	}
	defer database.Close()

	// Create a router to map incoming requests to handler functions
	router := gin.Default()

	// No reverse proxy to handle for this server
	err = router.SetTrustedProxies(nil)

	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// Register the User type for encoding/decoding
	gob.Register(types.User{})

	// Create a cookie store for session management
	cookieStore := sessions.NewCookieStore([]byte(cookieStoreKey))
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.Secure = true
	cookieStore.Options.SameSite = http.SameSiteStrictMode
	cookieStore.Options.MaxAge = 604800

	// Load HTML templates
	router.LoadHTMLGlob("./internal/templates/*.html")

	// Create app struct for accessing session & database
	app := &types.App{SessionStore: cookieStore, Database: database}

	// Middleware for blocking suspicious IPs
	router.Use(api.BlockSuspiciousIPsAndRateLimit)

	// Invalid Routes
	router.NoRoute(api.GetNotFoundHandler)

	// Public Routes (No authentication required)
	router.GET("/", api.OptionalAuth(app), api.GetHomePageHandler)
	router.GET("/profile/:username", api.OptionalAuth(app), api.GetUserBlogPostsHandler(app.Database))
	router.GET("/blogpost/:ID", api.OptionalAuth(app), api.GetBlogPostPageHandler(app))
	router.GET("/login", api.GetLoginPageHandler)
	router.GET("/signup", api.GetSignupPageHandler)
	router.POST("/login", api.PostLoginHandler(app))
	router.POST("/signup", api.PostSignupHandler(app.Database))

	// Authenticated Routes (Require authentication)
	authRoutes := router.Group("/")
	authRoutes.Use(api.RequireAuth(app))
	{
		authRoutes.GET("/edit/:ID", api.GetCreateOrEditPostPageHandler(app))
		authRoutes.POST("/createpost", api.CreatePostHandler(app))
		authRoutes.POST("/edit/:ID", api.UpdatePostHandler(app))
		authRoutes.GET("/createpost", api.GetCreateOrEditPostPageHandler(app))
		authRoutes.POST("/delete/:ID", api.DeletePostHandler(app))
		authRoutes.POST("/logout", api.PostLogoutHandler(app))
		authRoutes.POST("/blogpost/:ID/comment", api.PostCommentHandler(app))
		authRoutes.POST("/blogpost/:ID/like", api.PostLikeHandler(app))
	}

	// Configure public folder to be accessed from root directory
	router.Use(static.Serve("/", static.LocalFile("./public", false)))

	// Start the HTTP server on port 8080
	if err := router.Run(appPort); err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}
}
