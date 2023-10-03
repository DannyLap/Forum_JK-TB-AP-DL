package structs

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// structure of the login page
type Login struct {
	Template *template.Template
	Username string
	Password string
	Error    string
	Logged   bool
}

// creation of a cookie store
var store = sessions.NewCookieStore([]byte("clé_secrète"))

// ServeHTTP handles the login page and http requests
func (h Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "POST" {
		h.LoginPostHandler(w, r)
		return
	}
	err := h.Template.Execute(w, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// LoginPostHandler handles the submission of the login form
func (l *Login) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		pseudo := r.FormValue("pseudo")
		mdp := r.FormValue("password")

		// creation of the db
		db, err := sql.Open("sqlite3", "./data/db.db")
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		defer db.Close()

		// information of the user from the db
		row := db.QueryRow("SELECT username, password FROM user WHERE username = ?", pseudo)
		var (
			username  string
			hashedPwd string
		)
		err = row.Scan(&username, &hashedPwd)
		if err != nil {
			l.Logged = false
			l.Error = "Le pseudo ou le mot de passe est incorrect."
			err := l.Template.Execute(w, l)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Password comparison and verification
		err = bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(mdp))
		if err != nil {
			l.Logged = false
			l.Error = "Le pseudo ou le mot de passe est incorrect."
			err := l.Template.Execute(w, l)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Creation of a new user session
		session, err := store.Get(r, "user-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["username"] = username
		session.Save(r, w)

		l.Logged = true
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

// function thatt gets the user by his username
func GetUserByUsername(username string) (*Login, error) {
	db, err := sql.Open("sqlite3", "./db.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT username, password FROM users WHERE username=?", username)
	var user Login
	err = row.Scan(&user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
