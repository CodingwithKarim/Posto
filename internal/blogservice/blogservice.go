package blogservice

import (
	"App/internal/types"
	"App/internal/utils"
	"database/sql"
	"fmt"
	"log"
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
		log.Printf("No rows affected while updating blog post ID %d.", postData.ID)
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

func GetBlogPostsByUser(db *sql.DB, username string, isOwner bool, page int) ([]types.BlogPostData, error) {
	var exists bool

	// check if user is present in the database
	if err := db.QueryRow(utils.UserExistsQuery, username).Scan(&exists); err != nil || !exists {
		return nil, fmt.Errorf("user %s does not exist or error occurred: %w", username, err)
	}

	// Calculate pagination offset based on post limit
	limit := utils.POST_LIMIT_PER_PAGE
	offset := (page - 1) * limit

	// Assign SQL query to fetch blog posts based on ownership
	var query string
	if !isOwner {
		query = utils.SelectPublicPostsByUsernameQuery
	} else {
		query = utils.SelectAllPostsByUsernameQuery
	}

	// Execute the query to retrieve blog posts from user
	rows, err := db.Query(query, username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error querying posts for user %s: %w", username, err)
	}
	defer rows.Close()

	// Collect the results
	var posts []types.BlogPostData
	for rows.Next() {
		var post types.BlogPostData
		var createdAt []byte

		// Scan the row into the post struct
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &createdAt, &post.IsPublic); err != nil {
			return nil, fmt.Errorf("error scanning post for user %s: %w", username, err)
		}

		// Limit content length to 100 characters and add dots
		if len(post.Content) > 100 {
			post.Content = post.Content[:100] + utils.DOTS_STRING
		}

		// Format the creation date for the UI
		post.CreatedAt = FormatDate(createdAt)

		// Add posts to array of posts
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts for user %s: %w", username, err)
	}

	// Return arr of posts & null if successful
	return posts, nil
}

func GetBlogPostData(db *sql.DB, postID int, userID int, isLoggedIn bool) (types.BlogPostPageData, error) {
	var pageData types.BlogPostPageData
	var createdAt []byte
	var postUserID int

	// Execute the query to retrieve blog post by ID
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
	pageData.LikesCount = 0
	pageData.HasUserLiked = false
	pageData.Comments = append(pageData.Comments, types.Comment{
		Content:   "Wow what a sweet post this is great stuff man",
		CreatedAt: "January 1st 2025",
	}, types.Comment{
		Content:   "Woo thats intetesting",
		CreatedAt: "January 2nd 2025",
	})

	return pageData, nil
}

func GetPostDataOnEdit(db *sql.DB, formData *types.BlogPostFormData, postID, userID int) error {
	// Execute SQL query to retrieve existing post data for edit page
	if err := db.QueryRow(utils.SelectEditPostQuery, postID, userID).Scan(&formData.Title, &formData.Content, &formData.IsPublic); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("post not found or unauthorized")
		}

		return fmt.Errorf("database error occurred while accessing the post")
	}
	return nil
}
