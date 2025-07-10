package types

type BlogPageData struct {
	Username    string
	Posts       []*BlogPostData
	IsOwner     bool
	IsLoggedIn  bool
	IsFollowing bool
	Tabs        int
	CurrentPage int
}

type BlogPostData struct {
	ID int
	BlogPostBase
	CreatedAt string
}

type HomeFeedData struct {
	BlogPostData
	Username string
}

type BlogPostPageData struct {
	Post         *BlogPostData
	Username     string
	IsLoggedIn   bool
	IsOwner      bool
	Comments     []*Comment
	LikesCount   int
	HasUserLiked bool
}

type CreateComment struct {
	PostID  int    `json:"postId"`
	UserID  int    `json:"userId"`
	Comment string `json:"comment"`
}

type Comment struct {
	ID        int    `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
	Username  string `json:"username"`
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

type BlogPreview struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
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

type HomeFeedPage struct {
	Posts       []*HomeFeedData
	CurrentPage int
	Tabs        int
}
