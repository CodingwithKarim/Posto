package utils

import "time"

const (
	AUTH_MIN_LENGTH = 3
	AUTH_MAX_LENGTH = 40
)

const (
	BLOG_POST_MIN_LENGTH    = 1
	BLOG_TITLE_MAX_LENGTH   = 100
	BLOG_CONTENT_MAX_LENGTH = 10000
)

const (
	REQUEST_LIMIT       = 60
	POST_LIMIT_PER_PAGE = 3
	BLOG_POST_PAGE_MAX  = 1000
)

const (
	EXPIRATION_TIME = 24 * time.Hour // Constant duration to block an IP (e.g., 24 hours)
)

const (
	PAGE         = "page"
	DEFAULT_PAGE = "1"
	DOTS_STRING  = "..."
)

const (
	ID       = "ID"
	USERNAME = "username"
	PASSWORD = "password"
	USER     = "user"
)

const (
	BLOG_POST_PAGE    = "blogpost.html"
	CREATE_POST_PAGE  = "createpost.html"
	ERROR_PAGE        = "error.html"
	ROOT_PAGE         = "index.html"
	LOGIN_PAGE        = "login.html"
	SIGNUP_PAGE       = "signup.html"
	USER_PROFILE_PAGE = "userprofile.html"
)

const (
	COOKIE_SESSION = "cookieSession"
)
