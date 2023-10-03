package structs

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"text/template"

	"github.com/mattn/go-sqlite3"
)

type Home struct {
	Template *template.Template
	DB       *sql.DB
	Topics   []Topic
	User     User
	Logged   bool
}

type Topic struct {
	ID   int64
	Name string
}

type User struct {
	Mail           string
	Password       string
	Password_verif string
	Username       string
	uid            string
}

// ServeHTTP handles HTTP requests for the home page.
func (h Home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.FormValue("newTopic") != "" {
		h.CreateTopic(w, r)
	}
	if r.URL.Path == "/logout" {
		LogoutHandler(w, r)
		return
	}
	//authenticate user from the database
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.User.Username, h.Logged = session.Values["username"].(string)

	h.MigrateLikes(h.User.Username)

	allTopics, err := h.AllTopics()
	h.Topics = allTopics
	if err != nil {
		log.Fatal(err)
	}

	err = h.Template.Execute(w, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MigrateHome creates the topics table if it doesn't exist.
func (h *Home) MigrateHome() error {
	query := `
    CREATE TABLE IF NOT EXISTS topics(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );
    `
	_, err := h.DB.Exec(query)
	return err
}

// CreateTopic creates a new topic in the database.
func (h *Home) CreateTopic(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	if session.Values["username"] == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return errors.New("user not logged in")
	}

	_, err = h.DB.Exec("INSERT INTO topics(name) values(?)", r.FormValue("newTopic"))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return ErrDuplicate
			}
		}
		return err
	}
	return nil
}

// AllTopics retrieves all topics from the database.
func (h *Home) AllTopics() ([]Topic, error) {
	rows, err := h.DB.Query("SELECT * FROM topics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var all []Topic
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.ID, &topic.Name); err != nil {
			return nil, err
		}
		all = append(all, topic)
	}
	return all, nil
}

// AllUser retrieves all users from the database.
func (h *Home) AllUser() ([]User, error) {
	rows, err := h.DB.Query("SELECT * FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var all []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.uid, &user.Username); err != nil {
			return nil, err
		}
		all = append(all, user)
	}
	return all, nil
}

// LogoutHandler handles the logout process.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["username"] = nil
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/home", http.StatusFound)
}
