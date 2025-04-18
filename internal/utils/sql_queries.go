package utils

const (
	UserExistsQuery = "SELECT EXISTS(SELECT 1 FROM Users WHERE Username = ?)"
)

const (
	GetUserCredentialsQuery = "SELECT ID, Password, Encryption_Salt FROM Users WHERE Username = ?"
)

const (
	InsertUserQuery = "INSERT INTO Users (Username, Password, Encryption_Salt) VALUES (?, ?, ?)"
)

const (
	UpdatePostQuery = "UPDATE Posts SET Title = ?, Content = ?, IsPublic = ? WHERE ID = ? AND UserID = ?"
)

const (
	InsertPostQuery = "INSERT INTO Posts (Title, Content, UserID, IsPublic) VALUES (?, ?, ?, ?)"
)

const (
	DeletePostQuery = "DELETE FROM Posts WHERE ID = ? AND UserID = ?"
)

const (
	SelectPostsByUsername = `
		SELECT ID, Title, Content, CreatedAt, IsPublic, Count(*) OVER() AS total_count
		FROM Posts
		WHERE UserID = (SELECT ID FROM Users WHERE Username = ?)
		AND (IsPublic = 1 OR UserID = ?)
		ORDER BY CreatedAt DESC
		LIMIT ? OFFSET ?`

	SelectPostDetailsQuery = `
        SELECT 
            p.ID, p.Title, p.Content, p.CreatedAt, 
            p.IsPublic, p.UserID, u.Username
        FROM Posts p
        JOIN Users u ON p.UserID = u.ID
        WHERE p.ID = ? AND (p.IsPublic = 1 OR p.UserID = ?)
    `
)

const (
	SelectEditPostQuery = `
        SELECT Title, Content, IsPublic
        FROM Posts
        WHERE ID = ? AND UserID = ?
    `
)

const (
	InsertCommentQuery         = `INSERT INTO Comments (PostID, UserID, Comment) VALUES (?, ?, ?)`
	SelectCommentsForPostQuery = `
	SELECT 
    c.ID,
    c.Comment, 
    c.CreatedAt, 
    u.Username 
FROM 
    Comments c
JOIN 
    Users u ON c.UserID = u.ID
WHERE 
    c.PostID = ?
ORDER BY 
    c.CreatedAt ASC`
)

const (
	InsertLikeQuery     = `INSERT INTO Likes (UserID, PostID) VALUES (?, ?)`
	DeleteLikeQuery     = `DELETE FROM Likes WHERE UserID = ? AND PostID = ?`
	CountLikesQuery     = `SELECT COUNT(*) FROM Likes WHERE PostID = ?`
	CheckUserLikedQuery = `
		 SELECT EXISTS (
        SELECT 1 FROM Likes WHERE UserID = ? AND PostID = ?
    )
	`
)

const (
	GetUserIDQuery = `
        SELECT ID FROM Users WHERE Username = ?`

	CheckFollowQuery = `
        SELECT EXISTS (
            SELECT 1 FROM User_Follows 
            WHERE follower_id = ? AND following_id = ?
        )`

	InsertFollowQuery = `
        INSERT INTO User_Follows (follower_id, following_id) 
        VALUES (?, ?)`

	DeleteFollowQuery = `
        DELETE FROM User_Follows 
        WHERE follower_id = ? AND following_id = ?`
)

const SelectHomeFeedPostsQuery = `
    SELECT 
        Posts.ID, 
        Posts.Title, 
        Posts.Content, 
        Posts.CreatedAt, 
        Posts.IsPublic, 
        Users.Username AS AuthorUsername,
		Count(*) OVER() AS total_count
    FROM Posts
    JOIN User_Follows ON Posts.UserID = User_Follows.following_id
    JOIN Users ON Users.ID = Posts.UserID
    WHERE User_Follows.follower_id = ? 
      AND Posts.IsPublic = 1
    ORDER BY Posts.CreatedAt DESC
    LIMIT ? OFFSET ?`
