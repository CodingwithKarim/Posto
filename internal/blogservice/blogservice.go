package blogservice

import (
	"App/internal/types"
	"App/internal/utils"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func InsertBlogPostIntoDB(db *sql.DB, postData *types.CreateBlogPost) error {
	// Execute the SQL query
	if result, err := db.Exec(utils.InsertPostQuery, postData.Title, postData.Content, postData.UserID, postData.IsPublic); err != nil {
		log.Printf("SQL execution error while inserting blog post: %v", err)
		return fmt.Errorf("database error: failed to insert blog post")

	} else if rowsAffected, err := result.RowsAffected(); err != nil {
		log.Printf("Error retrieving affected rows for blog post insertion: %v", err)
		return fmt.Errorf("database error: unable to confirm blog post creation")

	} else if rowsAffected == 0 {
		log.Printf("No rows affected while inserting blog post. Post data: %+v", postData)
		return fmt.Errorf("blog post creation unsuccessful")
	}

	// Return nil if inserting post into DB was successful
	return nil
}

func UpdateBlogPostInDB(db *sql.DB, postData *types.UpdateBlogPost) error {
	// Execute the SQL query
	if result, err := db.Exec(utils.UpdatePostQuery, postData.Title, postData.Content, postData.IsPublic, postData.ID, postData.UserID); err != nil {
		log.Printf("SQL execution error while updating blog post ID %d: %v", postData.ID, err)
		return fmt.Errorf("database error: failed to update blog post")

	} else if rowsAffected, err := result.RowsAffected(); err != nil {
		log.Printf("Error retrieving affected rows for blog post update ID %d: %v", postData.ID, err)
		return fmt.Errorf("database error: unable to confirm blog post update")

	} else if rowsAffected == 0 {
		log.Printf("No rows affected while updating blog post ID %d. Post data: %+v", postData.ID, postData)
		return fmt.Errorf("blog post update unsuccessful")
	}

	// Return nil if updating post in DB was successful
	return nil
}

func DeleteBlogPostFromDB(db *sql.DB, postID int, userID int) error {
	// Execute the SQL query
	if result, err := db.Exec(utils.DeletePostQuery, postID, userID); err != nil {
		log.Printf("SQL execution error while deleting blog post ID %d by UserID %d: %v", postID, userID, err)
		return fmt.Errorf("database error: failed to delete blog post")

	} else if rowsAffected, err := result.RowsAffected(); err != nil {
		log.Printf("Error retrieving affected rows for blog post deletion ID %d by UserID %d: %v", postID, userID, err)
		return fmt.Errorf("database error: unable to confirm blog post deletion")

	} else if rowsAffected == 0 {
		log.Printf("No rows affected while deleting blog post ID %d by UserID %d", postID, userID)
		return fmt.Errorf("blog post deletion unsuccessful: no matching post or unauthorized")
	}

	// Return nil if deletion was successful
	return nil
}

func GetBlogPostsByUser(db *sql.DB, username string, isOwner bool, page, limit int) ([]types.BlogPostData, error) {
	// Check if the user exists
	var exists bool

	if err := db.QueryRow(utils.UserExistsQuery, username).Scan(&exists); err != nil {
		return nil, fmt.Errorf("error checking if user exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user %s does not exist", username)
	}

	// Calculate pagination offset
	offset := (page - 1) * limit

	// SQL query to fetch blog posts
	var query string

	if !isOwner {
		query = utils.SelectPublicPostsByUsernameQuery
	} else {
		query = utils.SelectAllPostsByUsernameQuery
	}

	// Execute the query
	rows, err := db.Query(query, username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error querying posts for user %s: %w", username, err)
	}
	defer rows.Close()

	// Collect the results
	var posts []types.BlogPostData
	for rows.Next() {
		var post types.BlogPostData
		var createdAt []byte // Use []byte for the CreatedAt field, instead of sql.NullTime

		// Scan the row into the post struct
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &createdAt, &post.IsPublic); err != nil {
			return nil, fmt.Errorf("error scanning post for user %s: %w", username, err)
		}

		// Limit content length to 100 characters
		if len(post.Content) > 100 {
			post.Content = post.Content[:100] + utils.DOTS_STRING
		}

		// Format the creation date
		post.CreatedAt = FormatDate(createdAt)

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts for user %s: %w", username, err)
	}

	return posts, nil
}

func GetBlogPostData(db *sql.DB, postID int, userID int, isLoggedIn bool) (types.BlogPostPageData, error) {
	var pageData types.BlogPostPageData
	var createdAt []byte
	var postUserID int

	// Execute the query and populate the page data
	if err := db.QueryRow(utils.SelectPostDetailsQuery, postID, userID).Scan(
		&pageData.Post.ID, &pageData.Post.Title, &pageData.Post.Content,
		&createdAt, &pageData.Post.IsPublic, &postUserID, &pageData.Username,
	); err != nil {
		return pageData, fmt.Errorf("post not found or access denied")
	}

	// Format the username & created date
	pageData.Username = utils.CapitalizeFirstLetter(pageData.Username)
	pageData.Post.CreatedAt = FormatDate(createdAt)

	// Determine ownership and login status
	pageData.IsOwner = isLoggedIn && userID == postUserID
	pageData.IsLoggedIn = isLoggedIn

	return pageData, nil
}

func GetPostDataOnEdit(db *sql.DB, formData *types.BlogPostFormData, postID, userID int) error {
	if err := db.QueryRow(utils.SelectEditPostQuery, postID, userID).Scan(&formData.Title, &formData.Content, &formData.IsPublic); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("post not found or unauthorized")
		}

		return fmt.Errorf("database error occurred while accessing the post")
	}
	return nil
}

func GetPostIDAndMode(context *gin.Context) (int, bool, error) {
	paramPostID := context.Param(utils.ID)

	// If no ID is provided, return default values
	if paramPostID == "" {
		return 0, false, nil
	}

	// Validate postID if provided
	if id, isValidID := IsValidPostID(paramPostID); isValidID {
		return id, true, nil
	}

	return 0, false, fmt.Errorf("invalid post ID")
}

func GetPageQuery(ctx *gin.Context) int {
	const maxPage = utils.BLOG_POST_PAGE_MAX // Maximum allowable page number

	// Parse the query parameter and default to 1 if invalid
	page, err := strconv.Atoi(ctx.DefaultQuery(utils.PAGE, utils.DEFAULT_PAGE))

	// Ensure the page is within the valid range
	if err != nil || page < 1 || page > maxPage {
		if page > maxPage {
			return maxPage
		}
		return 1
	}

	return page
}

func ValidatePostIDInput(context *gin.Context) (int, error) {
	id, isValidID := IsValidPostID(context.Param(utils.ID))

	if !isValidID {
		return id, fmt.Errorf("invalid post ID")
	}

	return id, nil
}

func ValidatePostInputs(title string, isPublic string, content string) error {
	// Validate public status (isPublic)
	if !IsValidPriority(isPublic) {
		return fmt.Errorf("invalid public status: must be 'true' or 'false'")
	}

	// Validate title length
	if !utils.IsValidInputLength(title, utils.BLOG_POST_MIN_LENGTH, utils.BLOG_TITLE_MAX_LENGTH) {
		log.Printf("invalid title length: must be between %d and %d", utils.BLOG_POST_MIN_LENGTH, utils.BLOG_TITLE_MAX_LENGTH)
		return fmt.Errorf("invalid title length: must be between %d and %d", utils.BLOG_POST_MIN_LENGTH, utils.BLOG_TITLE_MAX_LENGTH)
	}

	// Validate content length
	if !utils.IsValidInputLength(content, utils.BLOG_POST_MIN_LENGTH, utils.BLOG_CONTENT_MAX_LENGTH) {
		return fmt.Errorf("invalid content length: must be between %d and %d", utils.BLOG_POST_MIN_LENGTH, utils.BLOG_CONTENT_MAX_LENGTH)
	}

	// Return nil if all validations pass
	return nil
}
