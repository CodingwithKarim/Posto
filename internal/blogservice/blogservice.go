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

func GetBlogPostData(db *sql.DB, postID int, userID int, isLoggedIn bool) (*types.BlogPostPageData, error) {
	var pageData = &types.BlogPostPageData{}
	var createdAt []byte
	var postUserID int

	// Execute the query to retrieve blog post by ID
	if err := db.QueryRow(utils.SelectPostDetailsQuery, postID, userID).Scan(
		&pageData.Post.ID, &pageData.Post.Title, &pageData.Post.Content,
		&createdAt, &pageData.Post.IsPublic, &postUserID, &pageData.Username,
	); err != nil {
		return nil, fmt.Errorf("post not found or access denied")
	}

	// Format the username & created date
	pageData.Username = utils.CapitalizeFirstLetter(pageData.Username)
	pageData.Post.CreatedAt = FormatDate(createdAt)

	// Determine ownership and login status
	pageData.IsOwner = isLoggedIn && userID == postUserID
	pageData.IsLoggedIn = isLoggedIn

	likesCount, err := GetLikesCount(db, postID)

	if err != nil {
		log.Printf("Error fetching likes count for post %d: %v", postID, err)
		return nil, err
	}

	pageData.LikesCount = likesCount

	if isLoggedIn {
		hasLiked, err := HasUserLikedPost(db, postID, userID)

		if err != nil {
			log.Printf("Error checking if user %d has liked post %d: %v", userID, postID, err)
			return nil, fmt.Errorf("database error: failed to check like status")
		}

		pageData.HasUserLiked = hasLiked
	}

	if pageData.LikesCount > 0 {
		likedUsers, err := GetUsersWhoLiked(db, postID)
		if err != nil {
			log.Printf("Error fetching users who liked post %d: %v", postID, err)
			return nil, fmt.Errorf("database error: failed to retrieve users who liked")
		} else {
			pageData.LikedUsers = likedUsers
		}
	}

	// Get comments for the blog post if any
	comments, err := GetCommentsForBlogPost(db, postID)

	if err != nil {
		log.Printf("Error fetching comments for post %d: %v", postID, err)
		return nil, fmt.Errorf("database error: failed to retrieve comments")
	}

	pageData.Comments = comments

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

func InsertCommentIntoDB(db *sql.DB, commentData *types.CreateComment) error {
	// Execute the SQL query to insert a comment
	if result, err := db.Exec(utils.InsertCommentQuery, commentData.PostID, commentData.UserID, commentData.Comment); err != nil {
		log.Printf("SQL execution error while inserting comment: %v", err)
		return fmt.Errorf("database error: failed to insert comment")

	} else if rowsAffected, err := result.RowsAffected(); err != nil {
		log.Printf("Error retrieving affected rows for comment insertion: %v", err)
		return fmt.Errorf("database error: unable to confirm comment creation")

	} else if rowsAffected == 0 {
		log.Printf("No rows affected while inserting comment. Comment data: %+v", commentData)
		return fmt.Errorf("comment creation unsuccessful")
	}

	// Return nil if inserting comment into DB was successful
	return nil
}

func GetCommentsForBlogPost(db *sql.DB, postID int) ([]types.Comment, error) {
	// Query to get comments for a post, joined with user table to get usernames
	rows, err := db.Query(utils.SelectCommentsForPostQuery, postID)
	if err != nil {
		return nil, fmt.Errorf("error querying comments for post %d: %w", postID, err)
	}
	defer rows.Close()

	// Collect the results
	var comments []types.Comment
	for rows.Next() {
		var comment types.Comment
		var createdAt []byte
		var username string

		// Scan the row into the comment struct
		if err := rows.Scan(&comment.ID, &comment.Content, &createdAt, &username); err != nil {
			return nil, fmt.Errorf("error scanning comment for post %d: %w", postID, err)
		}

		// Format the creation date
		comment.CreatedAt = FormatDate(createdAt)
		comment.Username = utils.CapitalizeFirstLetter(username)

		// Add comment to array of comments
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments for post %d: %w", postID, err)
	}

	return comments, nil
}

func ToggleLikeOnPost(db *sql.DB, postID int, userID int) (bool, error) {
	// Check if the user has already liked the post
	var exists bool
	err := db.QueryRow(utils.CheckUserLikedQuery, userID, postID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			exists = false
		} else {
			log.Printf("Error checking if user %d has liked post %d: %v", userID, postID, err)
			return false, fmt.Errorf("database error: failed to check like status: %w", err)
		}
	}

	var result sql.Result

	if !exists {
		// If not liked, add a like
		result, err = db.Exec(utils.InsertLikeQuery, userID, postID)
		if err != nil {
			log.Printf("Error adding like for post %d by user %d: %v", postID, userID, err)
			return false, fmt.Errorf("database error: failed to add like: %w", err)
		}
	} else {
		// If already liked, remove the like
		result, err = db.Exec(utils.DeleteLikeQuery, userID, postID)
		if err != nil {
			log.Printf("Error removing like for post %d by user %d: %v", postID, userID, err)
			return false, fmt.Errorf("database error: failed to remove like: %w", err)
		}
	}

	// Validate that the operation affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error retrieving affected rows for like operation on post %d by user %d: %v", postID, userID, err)
		return false, fmt.Errorf("database error: unable to confirm like operation: %w", err)
	}
	if rowsAffected == 0 {
		log.Printf("No rows affected during like operation for post %d by user %d", postID, userID)
		return false, fmt.Errorf("like operation unsuccessful")
	}

	return !exists, nil
}

func GetLikesCount(db *sql.DB, postID int) (int, error) {
	var count int

	if err := db.QueryRow(utils.CountLikesQuery, postID).Scan(&count); err != nil {
		return 0, fmt.Errorf("error counting likes: %w", err)
	}

	return count, nil
}

func HasUserLikedPost(db *sql.DB, postID, userID int) (bool, error) {
	var exists bool

	if err := db.QueryRow(utils.CheckUserLikedQuery, postID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("error checking like status: %w", err)
	}

	return exists, nil
}

// GetUsersWhoLiked retrieves a list of users who liked a post
func GetUsersWhoLiked(db *sql.DB, postID int) ([]types.LikeUser, error) {
	query := `
		SELECT u.Username 
		FROM Likes l
		JOIN Users u ON l.UserID = u.ID
		WHERE l.PostID = ?
		ORDER BY u.Username ASC
	`

	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("error querying users who liked: %w", err)
	}
	defer rows.Close()

	var users []types.LikeUser
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, fmt.Errorf("error scanning username: %w", err)
		}

		users = append(users, types.LikeUser{
			Username: utils.CapitalizeFirstLetter(username),
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through users: %w", err)
	}

	return users, nil
}
