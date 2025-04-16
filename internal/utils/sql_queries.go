package utils

const (
	UserExistsQuery         = "SELECT EXISTS(SELECT 1 FROM users WHERE Username = ?)"
	GetUserCredentialsQuery = "SELECT ID, Password, Encryption_Salt FROM users WHERE Username = ?"
	InsertUserQuery         = "INSERT INTO users (Username, Password, Encryption_Salt) VALUES (?, ?, ?)"
)

const (
	InsertPostQuery = `
		INSERT INTO posts (Title, Content, UserID, IsPublic) 
		VALUES (?, ?, ?, ?)
	`
	UpdatePostQuery = `
		UPDATE posts 
		SET Title = ?, Content = ?, IsPublic = ? 
		WHERE ID = ? AND UserID = ?
	`
	DeletePostQuery = "DELETE FROM posts WHERE ID = ? AND UserID = ?"

	SelectPostsByUsername = `
		SELECT ID, Title, Content, CreatedAt, IsPublic, Count(*) OVER() AS total_count
		FROM posts
		WHERE UserID = (SELECT ID FROM users WHERE Username = ?)
		AND (IsPublic = 1 OR UserID = ?)
		ORDER BY CreatedAt DESC
		LIMIT ? OFFSET ?
	`

	SelectPostDetailsQuery = `
		SELECT 
			p.ID, p.Title, p.Content, p.CreatedAt, 
			p.IsPublic, p.UserID, u.Username
		FROM posts p
		JOIN users u ON p.UserID = u.ID
		WHERE p.ID = ? AND (p.IsPublic = 1 OR p.UserID = ?)
	`

	SelectEditPostQuery = `
		SELECT Title, Content, IsPublic
		FROM posts
		WHERE ID = ? AND UserID = ?
	`
)

const (
	InsertCommentQuery = `
		INSERT INTO comments (PostID, UserID, Comment) VALUES (?, ?, ?)`

	SelectCommentsForPostQuery = `
		SELECT 
			c.ID,
			c.Comment, 
			c.CreatedAt, 
			u.Username 
		FROM 
			comments c
		JOIN 
			users u ON c.UserID = u.ID
		WHERE 
			c.PostID = ?
		ORDER BY 
			c.CreatedAt ASC`
)

const (
	InsertLikeQuery     = `INSERT INTO likes (UserID, PostID) VALUES (?, ?)`
	DeleteLikeQuery     = `DELETE FROM likes WHERE UserID = ? AND PostID = ?`
	CountLikesQuery     = `SELECT COUNT(*) FROM likes WHERE PostID = ?`
	CheckUserLikedQuery = `
		SELECT EXISTS (
			SELECT 1 FROM likes WHERE UserID = ? AND PostID = ?
		)
	`
)

const (
	GetUserIDQuery = `
		SELECT ID FROM users WHERE Username = ?`

	CheckFollowQuery = `
		SELECT EXISTS (
			SELECT 1 FROM user_follows 
			WHERE follower_id = ? AND following_id = ?
		)`

	InsertFollowQuery = `
		INSERT INTO user_follows (follower_id, following_id) 
		VALUES (?, ?)`

	DeleteFollowQuery = `
		DELETE FROM user_follows 
		WHERE follower_id = ? AND following_id = ?`
)

const SelectHomeFeedPostsQuery = `
	SELECT 
		posts.ID, 
		posts.Title, 
		posts.Content, 
		posts.CreatedAt, 
		posts.IsPublic, 
		users.Username AS AuthorUsername,
		Count(*) OVER() AS total_count
	FROM posts
	JOIN user_follows ON posts.UserID = user_follows.following_id
	JOIN users ON users.ID = posts.UserID
	WHERE user_follows.follower_id = ? 
	  AND posts.IsPublic = 1
	ORDER BY posts.CreatedAt DESC
	LIMIT ? OFFSET ?`

const (
	InsertUserKeyQuery       = `INSERT INTO user_keys (User_ID, Key) VALUES (?, ?)`
	SelectUserKeyExistsQuery = `SELECT EXISTS(SELECT 1 FROM user_keys WHERE User_ID = ?)`
)
