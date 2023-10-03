package main

import (
	"database/sql"
	"forum/structs"

	"fmt"
	"log"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

// creation of the structure of the forum
type Forum struct {
	Home     structs.Home
	Chat     structs.Chat
	Login    structs.Login
	Register structs.Register
}

// creation of the templates, handler of the forum and the migration to the database
var (
	homeTemplate     = template.Must(template.ParseFiles("www/home.html"))
	loginTemplate    = template.Must(template.ParseFiles("www/login.html"))
	registerTemplate = template.Must(template.ParseFiles("www/register.html"))
	chatTemplate     = template.Must(template.ParseFiles("www/chat.html"))
	forum            = Forum{
		Home:     structs.Home{Template: homeTemplate},
		Chat:     structs.Chat{Template: chatTemplate},
		Login:    structs.Login{Template: loginTemplate},
		Register: structs.Register{Template: registerTemplate}}
	funcs = []func() error{
		forum.Home.MigrateHome,
		forum.Chat.MigrateChat,
		forum.Chat.MigratePost,
		forum.Chat.MigrateAnswer,
		forum.Chat.MigrateUser,
	}
)

func main() {
	db, err := sql.Open("sqlite3", "data/db.db")
	if err != nil {
		log.Fatal(err)
	}
	// creation of the database
	forum.Chat.DB = db
	forum.Home.DB = db
	forum.Register.DB = db

	for _, fn := range funcs {
		if err := fn(); err != nil {
			log.Fatal(err)
		}
	}
	// access to the home page of the forum
	fmt.Println("http://localhost:8080/home")

	http.Handle("/home", forum.Home)
	http.Handle("/chat", forum.Chat)
	http.Handle("/login", forum.Login)
	http.Handle("/register", forum.Register)
	// handler of the css files
	http.Handle("/style/", http.StripPrefix("/style/", http.FileServer(http.Dir("./www/static"))))
	// handler of the struct in the related files
	http.HandleFunc("/deleteAnswer", structs.AnswerDeletePostHandler)
	http.HandleFunc("/deletePost", structs.PostDeletePostHandler)
	http.HandleFunc("/changeTopic", structs.UpdateTopicPostHandler)
	http.HandleFunc("/likeDislike", structs.LikeDislikePostHandler)
	http.HandleFunc("/logout", structs.LogoutHandler)

	err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
