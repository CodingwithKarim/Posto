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
)

func main() {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Open or create the log file
	logFile, err := os.OpenFile("/home/ec2-user/logs/posto.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		log.Fatal("Error opening log file:", err)
	}

	defer logFile.Close() // Ensure the log file is closed when done

	// Set the log output to the log file
	log.SetOutput(logFile)

	// Get required system environment variables
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDB := os.Getenv("MYSQL_DB")
	cookieStoreKey := os.Getenv("COOKIE_STORE_KEY")
	certFile := os.Getenv("CERT_FILE_PATH")
	keyFile := os.Getenv("KEY_FILE_PATH")

	// Ensure all necessary environment variables are present
	if mysqlUser == "" || mysqlPassword == "" || mysqlHost == "" || mysqlDB == "" || cookieStoreKey == "" {
		log.Fatal("Required environment variables are missing.")
	}

	// Ensure the SSL certificate and key files are present
	if certFile == "" || keyFile == "" {
		log.Fatal("SSL certificate or key file paths are missing.")
	}

	// Connect to database through formatted connection string
	database, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlDB))
	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}

	// Ensure the database connection is valid
	if err := database.Ping(); err != nil {
		log.Fatal("Error pinging the database:", err)
	}
	defer database.Close()

	// Register the User type for encoding/decoding
	gob.Register(types.User{})

	// Create a cookie store for session management
	cookieStore := sessions.NewCookieStore([]byte(cookieStoreKey))
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.SameSite = http.SameSiteStrictMode
	cookieStore.Options.Domain = "postoblog.duckdns.org"
	cookieStore.Options.MaxAge = 604800
	cookieStore.Options.Secure = true

	// Create app struct for accessing session & database
	app := &types.App{SessionStore: cookieStore, Database: database}

	// Create a router to map incoming requests to handler functions
	router := gin.New()

	// no reverse proxy to trust
	err = router.SetTrustedProxies(nil)

	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err) // Log fatal error if setting trusted proxies fails
	}

	// Load HTML templates
	router.LoadHTMLGlob("./internal/templates/*.html")

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
	}

	// Configure public folder to be accessed from root directory
	router.Use(static.Serve("/", static.LocalFile("./public", false)))

	// Start the HTTPS server on port 443 using SSL certificates
	if err := router.RunTLS(":443", certFile, keyFile); err != nil {
		log.Fatal("Error starting HTTPS server:", err)
	}
}
