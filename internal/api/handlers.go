package api

import (
	"App/internal/blogservice"
	"App/internal/types"
	"App/internal/userservice"
	"App/internal/utils"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetLoginPageHandler(context *gin.Context) {
	// Render the login page template without template data
	context.HTML(http.StatusOK, utils.LOGIN_PAGE, nil)
}

func PostLoginHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Extract username and password from the request
		username, password := strings.ToLower(context.PostForm(utils.USERNAME)), context.PostForm(utils.PASSWORD)

		// Validate the username & password form inputs
		if err := userservice.ValidateAuthInputLength(username, password); err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		// Authenticate user credentials
		user, err := userservice.VerifyUserCredentials(username, password, app.Database)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusUnauthorized, err.Error())
			return
		}

		// Retrieve session data from the request
		if err := userservice.SaveUserSession(context, app, user); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to the user's profile page after successful login
		context.Redirect(http.StatusFound, "/"+user.Username)
	}
}

func PostLogoutHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Call LogoutSession to log the user out and handle any errors
		if err := userservice.LogoutUserSession(context, app.SessionStore); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to homepage after successful logout
		context.Redirect(http.StatusFound, "/")
	}
}

func GetSignupPageHandler(context *gin.Context) {
	// Render the login page template without template data
	context.HTML(http.StatusOK, utils.SIGNUP_PAGE, nil)
}

func PostSignupHandler(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Extract username and password from the form data
		username, password := strings.ToLower(context.PostForm("username")), context.PostForm("password")

		// Validate the username & password form inputs
		if err := userservice.ValidateAuthInputLength(username, password); err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		// Attempt to create a new user in the database
		if err := userservice.RegisterUser(username, password, database); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to the login page upon successful signup
		context.Redirect(http.StatusFound, "/login")
	}
}

func GetHomePageHandler(context *gin.Context) {
	// Check if the user is logged
	user, isLoggedIn := userservice.IsUserLoggedIn(context)

	if isLoggedIn {
		// If the user is logged in, redirect them to their profile page
		context.Redirect(http.StatusFound, "/"+user.Username)
	} else {
		// If the user is not logged in, render the default homepage
		context.HTML(http.StatusOK, utils.ROOT_PAGE, nil)
	}
}

func UpdatePostHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Retrieve form values
		title := context.PostForm("title")
		isPublic := context.DefaultPostForm("isPublic", "false")
		message := context.PostForm("message")

		// Validate form values
		if err := blogservice.ValidatePostInputs(title, isPublic, message); err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
		}

		id, err := blogservice.ValidatePostIDInput(context)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		if err := blogservice.UpdateBlogPostInDB(app.Database, &types.UpdateBlogPost{
			BlogPostBase: types.BlogPostBase{
				Title:    title,
				Content:  message,
				IsPublic: blogservice.ConvertIsPublicToBool(isPublic),
			},
			UserID: (userservice.GetUserFromContext(context)).ID,
			ID:     id,
		}); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
		}

		// Redirect to the updated post page
		context.Redirect(http.StatusFound, "/blogpost/"+strconv.Itoa(id))
	}
}

func CreatePostHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Retrieve form values
		title := context.PostForm("title")
		isPublic := context.DefaultPostForm("isPublic", "false") // Default to "0" if not selected
		message := context.PostForm("message")

		// Validate form values
		if err := blogservice.ValidatePostInputs(title, isPublic, message); err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		// Get user info from context
		user := userservice.GetUserFromContext(context)

		// Execute the query with parameterized values
		if err := blogservice.InsertBlogPostIntoDB(app.Database, &types.CreateBlogPost{
			BlogPostBase: types.BlogPostBase{
				Title:    title,
				IsPublic: blogservice.ConvertIsPublicToBool(isPublic),
				Content:  message,
			},
			UserID: user.ID,
		}); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to the user's page after creating the post
		context.Redirect(http.StatusFound, "/"+user.Username)
	}
}

func DeletePostHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Get user info from context (set in middleware)
		user := userservice.GetUserFromContext(context)

		// Validate the post ID
		id, err := blogservice.ValidatePostIDInput(context)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		if err := blogservice.DeleteBlogPostFromDB(app.Database, id, user.ID); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to the user's page after successful deletion
		context.Redirect(http.StatusFound, "/"+user.Username)
	}
}

func GetUserBlogPostsHandler(db *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Retrieve the username from the URL parameter
		username := strings.ToLower(context.Param(utils.USERNAME))

		// Check if the user is logged in and if the requested user is the owner of the blog
		isLoggedIn, isOwner := userservice.GetUserStatus(context, username)

		// Handle pagination to determine which posts to retrieve
		page := blogservice.GetPageQuery(context)

		// Fetch the blog posts from the database
		posts, err := blogservice.GetBlogPostsByUser(db, username, isOwner, page, utils.POST_LIMIT_PER_PAGE)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusNotFound, err.Error())
			return
		}

		// Check if the request is an AJAX request by inspecting the "X-Requested-With" header
		// If it's an AJAX request, return the posts in JSON format for client-side handling
		if context.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			context.JSON(http.StatusOK, posts)
			return
		}

		// If not AJAX, We pass the page data to render the HTML page
		context.HTML(http.StatusOK, utils.USER_PROFILE_PAGE, &types.BlogPageData{
			Username:   utils.CapitalizeFirstLetter(username),
			Posts:      posts,
			IsOwner:    isOwner,
			IsLoggedIn: isLoggedIn,
		})
	}
}

func GetBlogPostPageHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		id, err := blogservice.ValidatePostIDInput(context)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		// Check if the user is logged in
		user, isLoggedIn := userservice.IsUserLoggedIn(context)

		// Get blog post data from the database
		pageData, err := blogservice.GetBlogPostData(app.Database, id, user.ID, isLoggedIn)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusNotFound, err.Error())
			return
		}

		// Render the blog post
		context.HTML(http.StatusOK, utils.BLOG_POST_PAGE, pageData)
	}
}

func GetCreateOrEditPostPageHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Get postID and edit mode
		postID, isEditMode, err := blogservice.GetPostIDAndMode(context)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		// Get user info from the context (set in middleware)
		user := userservice.GetUserFromContext(context)

		log.Println(user.ID, user.Username)

		// Initialize form data with default values
		formData := &types.BlogPostFormData{
			Username:  utils.CapitalizeFirstLetter(user.Username),
			IsEditing: isEditMode,
			PostID:    postID,
			BlogPostBase: types.BlogPostBase{
				IsPublic: true, // Default for new posts
			},
		}

		// Populate form data if editing a post
		if isEditMode {
			if err := blogservice.GetPostDataOnEdit(app.Database, formData, postID, user.ID); err != nil {
				utils.SendErrorResponse(context, http.StatusNotFound, err.Error())
				return
			}
		}

		// Render the create or edit post page
		context.HTML(http.StatusOK, utils.CREATE_POST_PAGE, formData)
	}
}
