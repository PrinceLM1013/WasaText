/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net/http"

	"github.com/PrinceLM1013/WasaText/service/api/reqcontext"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error

	Ping() error

	// Session and User management
	CreateUser(userName string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	Authorize(username string, token string, w http.ResponseWriter, ctx reqcontext.RequestContext) (is_valid bool)
	ChangeUserName(newName string, username string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	UpPhoto(username string, fileBytes []byte, w http.ResponseWriter, ctx reqcontext.RequestContext)

	// Conversations
	GetMyConversations(username string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	GetConversation(username string, conversationID string, w http.ResponseWriter, ctx reqcontext.RequestContext)

	// Messages
	SendMessage(username string, content string, conversationID string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	ForwardMessage(username string, messageID string, conversationID string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	CommentMessage(username string, messageID string, comment string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	UncommentMessage(username string, messageID string, commentID string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	DeleteMessage(username string, messageID string, w http.ResponseWriter, ctx reqcontext.RequestContext)

	// Group management
	LeaveGroup(username string, groupID string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	SetGroupPhoto(username string, groupID string, fileBytes []byte, w http.ResponseWriter, ctx reqcontext.RequestContext)
	AddToGroup(username string, groupID string, newMember string, w http.ResponseWriter, ctx reqcontext.RequestContext)
	SetGroupName(username string, groupID string, newName string, w http.ResponseWriter, ctx reqcontext.RequestContext)
}

var writeErr = "error writing response"
var rowErr = `{"error": "Failed to scan row", "ERR": "`

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Check if table exists. If not, the database is empty, and we need to create the structure
	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='example_table';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE example_table (id INTEGER NOT NULL PRIMARY KEY, name TEXT);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	query := `
			CREATE TABLE IF NOT EXISTS user
	`

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
