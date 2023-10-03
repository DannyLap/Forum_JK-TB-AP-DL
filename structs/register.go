package structs

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Register struct {
	Template *template.Template
	DB       *sql.DB
	Error    bool
	ErrorMsg string
}


// ServeHTTP handles HTTP requests for registration.
func (h Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "POST" {
		h.RegisterPostHandler(w, r)
		return
	}
	err := h.Template.Execute(w, &h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RegisterPostHandler handles the registration form submission.
func (h *Register) RegisterPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		pseudo, email, pwd, pwdv := r.FormValue("pseudo"), r.FormValue("mail"), r.FormValue("password"), r.FormValue("password_verif")

		//her check if the password and the password verification is the same
		if pwd != pwdv {
			h.Error = true
			h.ErrorMsg = "Le mot de passe et la vérification du mot de passe ne sont pas les mêmes."
			err := h.Template.Execute(w, h)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		//her we hash the password and entered it into the database
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			fmt.Fprintf(w, "Erreur lors du hachage du mot de passe : %s", err.Error())
			return
		}

		db, err := sql.Open("sqlite3", "./data/db.db")
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		defer db.Close()//close the database in the end of code

		rows, err := db.Query("SELECT username, password, mail FROM user WHERE username = ? OR mail = ?", pseudo, email)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		defer rows.Close()//close results line in the end of code

		//variable for verifying if the password and mail are used
		foundPseudo, foundEmail := false, false 

		for rows.Next() {
			var (
				username string
				password string
				mail     string
			)

			err := rows.Scan(&username, &password, &mail)
			if err != nil {
				fmt.Fprintf(w, "Error: %s", err.Error())
				return
			}
			//verify if username and mail are used
			if pseudo == username {
				foundPseudo = true
			}
			if email == mail {
				foundEmail = true
			}
		}

		if err := rows.Err(); err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		//this line verifies if the password and mail are used and returns error 
		if foundPseudo {
			h.Error = true
			h.ErrorMsg = "Le pseudo est déjà utilisé."
		} else if foundEmail {
			h.Error = true
			h.ErrorMsg = "L'e-mail est déjà utilisé."
		} else {
			_, err := db.Exec("INSERT INTO user (username, mail, password, password_verif) VALUES (?, ?, ?, ?)", pseudo, email, hashedPwd, pwdv)
			if err != nil {
				fmt.Fprintf(w, "Erreur : %s", err.Error())
				return
			}
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		err = h.Template.Execute(w, h)//execute template of error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
