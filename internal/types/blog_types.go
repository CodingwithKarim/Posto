package types

type BlogPageData struct {
	Username   string
	Posts      []BlogPostData
	IsOwner    bool
	IsLoggedIn bool
}

type BlogPostData struct {
	ID int
	BlogPostBase
	CreatedAt string
}

type BlogPostPageData struct {
	Post       BlogPostData
	Username   string
	IsLoggedIn bool
	IsOwner    bool
}

type BlogPostFormData struct {
	Username  string
	IsEditing bool
	PostID    int
	BlogPostBase
}

type BlogPostBase struct {
	Title    string
	IsPublic bool
	Content  string
}

type CreateBlogPost struct {
	BlogPostBase
	UserID int
}

type UpdateBlogPost struct {
	BlogPostBase
	UserID int
	ID     int
}
