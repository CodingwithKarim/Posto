package utils

const (
	UserExistsQuery = "SELECT EXISTS(SELECT 1 FROM Users WHERE Username = ?)"
)

const (
	GetUserCredentialsQuery = "SELECT ID, Password FROM Users WHERE Username = ?"
)

const (
	InsertUserQuery = "INSERT INTO Users (Username, Password) VALUES (?, ?)"
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
	SelectPublicPostsByUsernameQuery = `
        SELECT ID, Title, Content, CreatedAt, IsPublic
        FROM Posts
        WHERE UserID = (SELECT ID FROM Users WHERE Username = ?) AND IsPublic = 1
		ORDER BY CreatedAt DESC LIMIT ? OFFSET ?
    `
)

const (
	SelectAllPostsByUsernameQuery = `
        SELECT ID, Title, Content, CreatedAt, IsPublic
        FROM Posts
        WHERE UserID = (SELECT ID FROM Users WHERE Username = ?)
		ORDER BY CreatedAt DESC LIMIT ? OFFSET ?
    `
)

const (
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
