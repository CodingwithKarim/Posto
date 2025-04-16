package blogservice

import (
	"App/internal/types"
	"App/internal/utils"
	"database/sql"
	"fmt"
	"log"
)

func InsertBlogPostIntoDB(db *sql.DB, postData *types.CreateBlogPost) error {
	// Encrypt blog content if needed
	title, content, err := EncryptBlogPost(postData.Title, postData.Content, postData.UserID, postData.IsPublic)

	if err != nil {
		log.Printf("Failed to encrypt blog post title and content: %v", err)
		return fmt.Errorf("encryption error: failed to encrypt blog post title and content")
	}

	// Execute the SQL query
	if result, err := db.Exec(utils.InsertPostQuery, title, content, postData.UserID, postData.IsPublic); err != nil {
		log.Printf("SQL execution error while inserting blog post: %v", err)
		return fmt.Errorf("database error: failed to insert blog post")

	} else if rowsAffected, err := result.RowsAffected(); err != nil {
		log.Printf("Error retrieving affected rows for blog post insertion: %v", err)
		return fmt.Errorf("database error: unable to confirm blog post creation")

	} else if rowsAffected == 0 {
		log.Printf("No rows affected while inserting blog post. .Post data: %+v", postData)
		return fmt.Errorf("blog post creation unsuccessful")
	}

	// Return nil if inserting post into DB was successful
	return nil
}

func UpdateBlogPostInDB(db *sql.DB, postData *types.UpdateBlogPost) error {
	// Encrypt blog content if needed
	title, content, err := EncryptBlogPost(postData.Title, postData.Content, postData.UserID, postData.IsPublic)

	if err != nil {
		return fmt.Errorf("encryption error: failed to encrypt blog post title and content")
	}

	// Execute the SQL query
	if result, err := db.Exec(utils.UpdatePostQuery, title, content, postData.IsPublic, postData.ID, postData.UserID); err != nil {
		log.Printf("SQL execution error while updating blog post ID %d: %v", postData.ID, err)
		return fmt.Errorf("database error: failed to update blog post")

	} else if _, err := result.RowsAffected(); err != nil {
		log.Printf("Error retrieving affected rows for blog post update ID %d: %v", postData.ID, err)
		return fmt.Errorf("database error: unable to confirm blog post update")
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

func GetBlogPostsByUser(db *sql.DB, username string, isOwner bool, page, userID int) ([]*types.BlogPostData, int, error) {
	// Check if user exists in the database
	var exists bool
	if err := db.QueryRow(utils.UserExistsQuery, username).Scan(&exists); err != nil || !exists {
		return nil, 0, fmt.Errorf("user %s does not exist or an error occurred", username)
	}

	// Calculate pagination offset based on the post limit
	limit := utils.POST_LIMIT_PER_PAGE
	offset := (page - 1) * limit

	// Execute the query to retrieve blog posts from the user
	rows, err := db.Query(utils.SelectPostsByUsername, username, userID, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("error querying posts for user %s: %w", username, err)
	}

	defer rows.Close()

	// Prepare the slice for the results
	var posts []*types.BlogPostData
	var totalCount int

	// Iterate over the rows to build the posts slice
	for rows.Next() {
		post := &types.BlogPostData{}
		var createdAt []byte

		// Scan the row into the post struct
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &createdAt, &post.IsPublic, &totalCount); err != nil {
			return nil, 0, fmt.Errorf("error scanning post for user %s: %w", username, err)
		}

		// Decrypt the content and title if needed
		title, content, err := DecryptBlogPost(post.Title, post.Content, userID, post.IsPublic)

		if err != nil {
			return nil, 0, fmt.Errorf("encryption error: failed to decrypt blog post title and content")
		}

		// Assign decrypted title and content to the post
		post.Content = TruncateString(content, utils.BLOG_POST_PREVIEW_LENGTH)
		post.Title = title

		// Format the creation date for the UI
		post.CreatedAt = FormatDate(createdAt)

		// Append the post to the results slice
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating posts for user %s: %w", username, err)
	}

	return posts, totalCount, nil
}

func GetBlogPostData(db *sql.DB, postID int, userID int, isLoggedIn bool) (*types.BlogPostPageData, error) {
	var pageData = &types.BlogPostPageData{
		Post: &types.BlogPostData{},
	}
	var createdAt []byte
	var postUserID int

	// Execute the query to retrieve blog post by ID
	if err := db.QueryRow(utils.SelectPostDetailsQuery, postID, userID).Scan(
		&pageData.Post.ID, &pageData.Post.Title, &pageData.Post.Content,
		&createdAt, &pageData.Post.IsPublic, &postUserID, &pageData.Username,
	); err != nil {
		return nil, fmt.Errorf("post not found or access denied")
	}

	// Decrypt the content if needed
	title, content, err := DecryptBlogPost(pageData.Post.Title, pageData.Post.Content, userID, pageData.Post.IsPublic)

	if err != nil {
		return nil, fmt.Errorf("encryption error: failed to decrypt blog post title and content")
	}

	// Assign decrypted title and content to the post data
	pageData.Post.Title = title
	pageData.Post.Content = content

	// Format the username & created date
	pageData.Username = utils.CapitalizeFirstLetter(pageData.Username)
	pageData.Post.CreatedAt = FormatDate(createdAt)

	// Determine ownership and login status
	pageData.IsOwner = isLoggedIn && userID == postUserID
	pageData.IsLoggedIn = isLoggedIn

	// Get likes count for the blog post
	likesCount, err := GetLikesCount(db, postID)

	if err != nil {
		log.Printf("Error fetching likes count for post %d: %v", postID, err)
		return nil, fmt.Errorf("database error: failed to retrieve likes count")
	}

	// Assign likes count to the page data
	pageData.LikesCount = likesCount

	if isLoggedIn {
		// Check if the user has liked the post
		hasLiked, err := HasUserLikedPost(db, postID, userID)

		if err != nil {
			log.Printf("Error checking if user %d has liked post %d: %v", userID, postID, err)
			return nil, fmt.Errorf("database error: failed to check like status")
		}

		// Assign like status to the page data
		pageData.HasUserLiked = hasLiked
	}

	// Get comments for the blog post if any
	comments, err := GetCommentsForBlogPost(db, postID)

	if err != nil {
		log.Printf("Error fetching comments for post %d: %v", postID, err)
		return nil, fmt.Errorf("database error: failed to retrieve comments")
	}

	// Assign comments to the page data
	pageData.Comments = comments

	// Return the page data if successful
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

	// Decrypt the content if needed
	title, content, err := DecryptBlogPost(formData.Title, formData.Content, userID, formData.IsPublic)

	if err != nil {
		return fmt.Errorf("encryption error: failed to decrypt blog post title and content")
	}

	// Assign decrypted/raw title and content to the form data
	formData.Title = title
	formData.Content = content

	// Return nil if retrieving post data for edit was successful
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

func GetCommentsForBlogPost(db *sql.DB, postID int) ([]*types.Comment, error) {
	// Query to get comments for a post, joined with user table to get usernames
	rows, err := db.Query(utils.SelectCommentsForPostQuery, postID)
	if err != nil {
		return nil, fmt.Errorf("error querying comments for post: %d", postID)
	}
	defer rows.Close()

	// Collect the results
	var comments []*types.Comment
	for rows.Next() {
		comment := &types.Comment{}
		var createdAt []byte
		var username string

		// Scan the row into the comment struct
		if err := rows.Scan(&comment.ID, &comment.Content, &createdAt, &username); err != nil {
			return nil, fmt.Errorf("error scanning comment for post: %d", postID)
		}

		// Format the creation date
		comment.CreatedAt = FormatDate(createdAt)
		comment.Username = utils.CapitalizeFirstLetter(username)

		// Add comment to array of comments
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments for post: %d", postID)
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
			return false, fmt.Errorf("database error: failed to check like status")
		}
	}

	var result sql.Result

	if !exists {
		// If not liked, add a like
		result, err = db.Exec(utils.InsertLikeQuery, userID, postID)
		if err != nil {
			log.Printf("Error adding like for post %d by user %d: %v", postID, userID, err)
			return false, fmt.Errorf("database error: failed to add like")
		}
	} else {
		// If already liked, remove the like
		result, err = db.Exec(utils.DeleteLikeQuery, userID, postID)
		if err != nil {
			log.Printf("Error removing like for post %d by user %d: %v", postID, userID, err)
			return false, fmt.Errorf("database error: failed to remove like")
		}
	}

	// Validate that the operation affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error retrieving affected rows for like operation on post %d by user %d: %v", postID, userID, err)
		return false, fmt.Errorf("database error: unable to confirm like operation")
	}
	if rowsAffected == 0 {
		log.Printf("No rows affected during like operation for post %d by user %d", postID, userID)
		return false, fmt.Errorf("like operation unsuccessful")
	}

	return !exists, nil
}

func GetLikesCount(db *sql.DB, postID int) (int, error) {
	var count int

	// Execute the SQL query to count likes for the post
	if err := db.QueryRow(utils.CountLikesQuery, postID).Scan(&count); err != nil {
		return 0, fmt.Errorf("error counting likes")
	}

	return count, nil
}

func HasUserLikedPost(db *sql.DB, postID, userID int) (bool, error) {
	var exists bool

	// Execute the SQL query to check if the user has liked the post
	if err := db.QueryRow(utils.CheckUserLikedQuery, userID, postID).Scan(&exists); err != nil {
		return false, fmt.Errorf("error checking like status")
	}

	return exists, nil
}

func ToggleFollowUser(db *sql.DB, followerID int, followingUsername string) error {
	// Retrieve the user ID of the user being followed
	var followingID int
	err := db.QueryRow(utils.GetUserIDQuery, followingUsername).Scan(&followingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with username '%s' not found", followingUsername)
		}
		log.Printf("Database error: Failed to retrieve user ID for username '%s': %v", followingUsername, err)
		return fmt.Errorf("database error: Failed to retrieve user ID")
	}

	// Check if the user is already following the other user
	var exists int
	err = db.QueryRow(utils.CheckFollowQuery, followerID, followingID).Scan(&exists)
	if err != nil {
		log.Printf("Database error: Failed to check follow status for user %d -> %d: %v", followerID, followingID, err)
		return fmt.Errorf("database error: Failed to check follow status")
	}

	var result sql.Result

	if exists == 0 {
		// If not following, add a follow
		result, err = db.Exec(utils.InsertFollowQuery, followerID, followingID)
		if err != nil {
			log.Printf("Database error: Failed to add follow for user %d -> %d: %v", followerID, followingID, err)
			return fmt.Errorf("database error: Failed to add follow")
		}
	} else {
		// If already following, remove the follow
		result, err = db.Exec(utils.DeleteFollowQuery, followerID, followingID)
		if err != nil {
			log.Printf("Database error: Failed to remove follow for user %d -> %d: %v", followerID, followingID, err)
			return fmt.Errorf("database error: Failed to remove follow")
		}
	}

	// Validate that the operation affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Database error: Failed to retrieve affected rows for follow operation: %v", err)
		return fmt.Errorf("database error: unable to confirm follow operation")
	}
	if rowsAffected == 0 {
		log.Printf("No rows affected during follow operation for user %d -> %d", followerID, followingID)
		return fmt.Errorf("follow operation unsuccessful: no matching follow or unauthorized")
	}

	return nil
}

func IsFollowingUser(db *sql.DB, followerID int, followingUsername string) (bool, error) {
	var followingID int

	// Execute the SQL query to retrieve the user ID of the user being followed
	db.QueryRow(utils.GetUserIDQuery, followingUsername).Scan(&followingID)

	var exists int

	// Execute the SQL query to check if the user is following the other user
	err := db.QueryRow(utils.CheckFollowQuery, followerID, followingID).Scan(&exists)

	if err != nil {
		log.Printf("Database error: Failed to check follow status for user %d -> %s: %v", followerID, followingUsername, err)
		return false, fmt.Errorf("database error: Failed to check follow status")
	}

	// Return true if the user is following, false otherwise
	return exists == 1, nil
}

func GetHomeFeedPosts(db *sql.DB, userID int, page int) ([]*types.HomeFeedData, int, error) {
	// Execute the query to retrieve blog posts from user
	limit := utils.POST_LIMIT_PER_PAGE
	offset := (page - 1) * limit

	rows, err := db.Query(utils.SelectHomeFeedPostsQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying posts for user %d: %w", userID, err)
	}
	defer rows.Close()

	// Collect the results
	var posts []*types.HomeFeedData
	var totalCount int = 0
	for rows.Next() {
		post := &types.HomeFeedData{}
		var createdAt []byte

		// Scan the row into the post struct
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &createdAt, &post.IsPublic, &post.Username, &totalCount); err != nil {
			return nil, 0, fmt.Errorf("error scanning post for user %d: %w", userID, err)
		}

		// Limit content length to 100 characters and add dots
		if len(post.Content) > 100 {
			post.Content = post.Content[:100] + utils.DOTS_STRING
		}

		post.Username = utils.CapitalizeFirstLetter(post.Username)

		// Format the creation date for the UI
		post.CreatedAt = FormatDate(createdAt)

		// Add posts to array of posts
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating posts for user %d: %w", userID, err)
	}

	// Return arr of posts & null if successful
	return posts, totalCount, nil
}
