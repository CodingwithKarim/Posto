package api

import (
	"App/internal/blogservice"
	"App/internal/types"
	"App/internal/userservice"
	"App/internal/utils"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetNotFoundHandler(context *gin.Context) {
	// If client attempts to access a route that doesn't exist, render an error page
	utils.SendErrorResponse(context, http.StatusNotFound, utils.INVALID_REQUEST_MESSAGE)
}

func GetHomePageHandler(context *gin.Context) {
	// Check if the user is logged in and redirect accordingly
	if user, isLoggedIn := userservice.IsUserLoggedIn(context); isLoggedIn {
		context.Redirect(http.StatusFound, "/profile/"+user.Username)
		return
	}

	// Render the default homepage if the user is not logged in
	context.HTML(http.StatusOK, utils.ROOT_PAGE, nil)
}

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

		// Authenticate user credentials and save session
		if err := userservice.VerifyUserCredentialsAndSaveSession(username, password, context, app); err != nil {
			utils.SendErrorResponse(context, http.StatusUnauthorized, err.Error())
			return
		}

		// Redirect to the user's profile page after successful login
		context.Redirect(http.StatusFound, "/profile/"+username)
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

func PostSignupHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Extract username and password from the form data
		username, password := strings.ToLower(context.PostForm("username")), context.PostForm("password")

		// Validate the username & password form inputs
		if err := userservice.ValidateAuthInputLength(username, password); err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		// Attempt to create a new user in the database
		if err := userservice.RegisterUserAndSaveSession(username, password, context, app); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to the user's profile page after successful signup
		context.Redirect(http.StatusFound, "/profile/"+username)
	}
}

func RenderUserProfilePageHandler(db *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Retrieve the username from the URL parameter
		username := strings.ToLower(context.Param(utils.USERNAME))

		// Check if the user is logged in and if the requested user is the owner of the blog
		user, isLoggedIn, isOwner := userservice.GetUserAndStatus(context, username)

		// Handle pagination to determine which posts to retrieve
		page := blogservice.GetPageQuery(context)

		// Fetch the blog posts from the database
		posts, totalCount, err := blogservice.GetBlogPostsByUser(db, username, isOwner, page, user.ID)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusNotFound, err.Error())
			return
		}

		previews := make([]types.BlogPreview, len(posts))

		for i, p := range posts {
			previews[i] = types.BlogPreview{
				ID:      p.ID,
				Title:   p.Title,
				Content: p.Content,
			}
		}

		isFollowing := false

		if isLoggedIn {
			if f, err := blogservice.IsFollowingUser(db, user.ID, username); err == nil {
				isFollowing = f
			} else {
				utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
				return
			}
		}

		htmlPayload := &types.BlogPageData{
			Username:    utils.CapitalizeFirstLetter(username),
			Posts:       posts,
			IsOwner:     isOwner,
			IsLoggedIn:  isLoggedIn,
			IsFollowing: isFollowing,
			CurrentPage: page,
			Tabs:        (totalCount + utils.POST_LIMIT_PER_PAGE - 1) / utils.POST_LIMIT_PER_PAGE,
		}

		context.Negotiate(http.StatusOK, gin.Negotiate{
			Offered:  []string{gin.MIMEJSON, gin.MIMEHTML},
			HTMLName: utils.USER_PROFILE_PAGE,
			JSONData: gin.H{"posts": previews},
			Data:     htmlPayload,
		})
	}
}

func RenderSingleBlogPostHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Validate Post ID from context
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
		context.Redirect(http.StatusFound, "/profile/"+user.Username)
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
			return
		}

		// Validate Post ID from context
		id, err := blogservice.ValidatePostIDInput(context)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		// Update Blog Post Data
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
			return
		}

		// Redirect to the updated post page
		context.Redirect(http.StatusFound, "/blogpost/"+strconv.Itoa(id))
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

		// Delete Blog Post
		if err := blogservice.DeleteBlogPostFromDB(app.Database, id, user.ID); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to the user's page after successful deletion
		context.Redirect(http.StatusFound, "/profile/"+user.Username)
	}
}

func PostCommentHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		postID, err := blogservice.ValidatePostIDInput(context)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		comment := context.PostForm("content")

		if err := blogservice.IsValidComment(comment); err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		user := userservice.GetUserFromContext(context)

		if err := blogservice.InsertCommentIntoDB(app.Database, &types.CreateComment{
			PostID:  postID,
			UserID:  user.ID,
			Comment: comment,
		}); err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		context.Redirect(http.StatusFound, "/blogpost/"+strconv.Itoa(postID))
	}
}

func PostLikeHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		postID, err := blogservice.ValidatePostIDInput(context)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusBadRequest, err.Error())
			return
		}

		user := userservice.GetUserFromContext(context)

		liked, err := blogservice.ToggleLikeOnPost(app.Database, postID, user.ID)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		context.JSON(http.StatusOK, gin.H{"liked": liked, "username": utils.CapitalizeFirstLetter(user.Username)})
	}
}

func PostFollowHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Get the username to follow from the request body
		username := strings.ToLower(context.Param("username"))

		if username == "" {
			utils.SendErrorResponse(context, http.StatusBadRequest, utils.INVALID_USERNAME_MESSAGE)
			return
		}

		// Get the current user from the context
		user := userservice.GetUserFromContext(context)

		// Attempt to toggle follow
		err := blogservice.ToggleFollowUser(app.Database, user.ID, username)
		if err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// No response needed, just send 200 OK
		context.Status(http.StatusOK)
	}
}

func GetHomeFeedHandler(app *types.App) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Get the current user from the context
		user := userservice.GetUserFromContext(context)

		// Handle pagination to determine which posts to retrieve
		page := blogservice.GetPageQuery(context)

		// Get the user's feed
		posts, totalCount, err := blogservice.GetHomeFeedPosts(app.Database, user.ID, page)

		if err != nil {
			utils.SendErrorResponse(context, http.StatusInternalServerError, err.Error())
			return
		}

		// If not AJAX, We pass the page data to render the HTML page
		context.HTML(http.StatusOK, "feed.html", &types.HomeFeedPage{
			Posts:       posts,
			CurrentPage: page,
			Tabs:        (totalCount + utils.POST_LIMIT_PER_PAGE - 1) / utils.POST_LIMIT_PER_PAGE,
		})
	}
}
