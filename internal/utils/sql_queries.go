package utils

const (
	UserExistsQuery         = "SELECT EXISTS(SELECT 1 FROM Users WHERE Username = ?)"
	GetUserCredentialsQuery = "SELECT ID, Password FROM Users WHERE Username = ?"
	InsertUserQuery         = "INSERT INTO Users (Username, Password) VALUES (?, ?)"
)

const (
	InsertPostQuery = `
		INSERT INTO Posts (Title, Content, UserID, IsPublic) 
		VALUES (?, ?, ?, ?)
	`
	UpdatePostQuery = `
		UPDATE Posts 
		SET Title = ?, Content = ?, IsPublic = ? 
		WHERE ID = ? AND UserID = ?
	`
	DeletePostQuery = "DELETE FROM Posts WHERE ID = ? AND UserID = ?"

	SelectPublicPostsByUsernameQuery = `
		SELECT ID, Title, Content, CreatedAt, IsPublic
		FROM Posts
		WHERE UserID = (SELECT ID FROM Users WHERE Username = ?) 
		  AND IsPublic = 1
		ORDER BY CreatedAt DESC 
		LIMIT ? OFFSET ?
	`

	SelectAllPostsByUsernameQuery = `
		SELECT ID, Title, Content, CreatedAt, IsPublic
		FROM Posts
		WHERE UserID = (SELECT ID FROM Users WHERE Username = ?)
		ORDER BY CreatedAt DESC 
		LIMIT ? OFFSET ?
	`

	SelectPostDetailsQuery = `
		SELECT 
			p.ID, p.Title, p.Content, p.CreatedAt, 
			p.IsPublic, p.UserID, u.Username
		FROM Posts p
		JOIN Users u ON p.UserID = u.ID
		WHERE p.ID = ? AND (p.IsPublic = 1 OR p.UserID = ?)
	`

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
