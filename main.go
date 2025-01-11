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

	// Create a router to map incoming requests to handler functions
	router := gin.Default()

	// Construct the connection string for MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlDB)

	database, err := sql.Open("mysql", dsn)
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
	cookieStore.Options.Domain = "postoblog.duckdns.org" // Updated to HTTPS domain
	cookieStore.Options.MaxAge = 604800
	cookieStore.Options.Secure = true

	// Load HTML templates
	router.LoadHTMLGlob("./internal/templates/*.html")

	// Create app struct for accessing session & database
	app := &types.App{SessionStore: cookieStore, Database: database}

	// Configure routes with associated handlers
	router.GET("/", api.OptionalAuth(app), api.GetHomePageHandler)
	router.GET("/:username", api.OptionalAuth(app), api.GetUserBlogPostsHandler(app.Database))
	router.GET("/:username/posts", api.OptionalAuth(app), api.GetUserBlogPostsHandler(app.Database))
	router.GET("/edit/:ID", api.RequireAuth(app), api.GetCreateOrEditPostPageHandler(app))
	router.GET("/blogpost/:ID", api.OptionalAuth(app), api.GetBlogPostPageHandler(app))
	router.POST("/createpost", api.RequireAuth(app), api.CreatePostHandler(app))
	router.POST("/edit/:ID", api.RequireAuth(app), api.UpdatePostHandler(app))
	router.GET("/login", api.GetLoginPageHandler)
	router.GET("/signup", api.GetSignupPageHandler)
	router.GET("/createpost", api.RequireAuth(app), api.GetCreateOrEditPostPageHandler(app))
	router.POST("/login", api.PostLoginHandler(app))
	router.POST("/signup", api.PostSignupHandler(app.Database))
	router.POST("/delete/:ID", api.RequireAuth(app), api.DeletePostHandler(app))
	router.POST("/logout", api.RequireAuth(app), api.PostLogoutHandler(app))

	// Configure public folder to be accessed from root directory
	router.Use(static.Serve("/", static.LocalFile("./public", false)))

	// Start the HTTPS server on port 443 using SSL certificates
	if err := router.RunTLS(":443", certFile, keyFile); err != nil {
		log.Fatal("Error starting HTTPS server:", err)
	}
}
