package types

import (
	"database/sql"

	"github.com/gorilla/sessions"
)

type App struct {
	SessionStore *sessions.CookieStore
	Database     *sql.DB
}

type User struct {
	Username string
	ID       int
}

type ErrorPageData struct {
	StatusCode   int
	ErrorMessage string
}
