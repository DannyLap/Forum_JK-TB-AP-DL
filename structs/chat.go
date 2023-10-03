package structs

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/mattn/go-sqlite3"
)

type Chat struct {
	Template     *template.Template
	DB           *sql.DB
	Topics       []Topic
	Topic        string
	Posts        []Post
	Username     string
	Like         Like
	Logged       bool
	ShowLiked    string
	ShowDisliked string
	ShowCreated  string
}

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

func (c Chat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	err := c.ChargeTopic()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Username, c.Logged = session.Values["username"].(string)

	post := r.FormValue("content")
	if post != "" {
		topic := r.FormValue("topic")
		if c.Logged {
			c.CreatePost(post, topic)
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
	}

	answer := r.FormValue("answer")
	if answer != "" {
		id, _ := strconv.Atoi(r.FormValue("idAnswer"))
		if c.Logged {
			c.CreateAnswer(answer, int64(id))
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
	}

	allPosts, err := c.AllPosts()
	c.Posts = allPosts
	if err != nil {
		log.Fatal(err)
	}

	h := new(Home)
	h.DB = c.DB
	allTopics, err := h.AllTopics()
	c.Topics = allTopics

	if c.Logged {
		c.AllDataAboutLike()
	}

	err = c.Template.Execute(w, c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if r.Method == "POST" && r.URL.Path == "/delete-answer" {
		answerID, _ := strconv.Atoi(r.FormValue("answerID"))
		c.DeleteAnswer(int64(answerID))
	}
}

func (c *Chat) MigrateChat() error {
	query := `
    CREATE TABLE IF NOT EXISTS chat(
        topic TEXT NOT NULL,
		showLiked TEXT NOT NULL,
		showDisliked TEXT NOT NULL,
		showCreated TEXT NOT NULL
    );`

	_, err := c.DB.Exec(query)
	return err
}

func (c *Chat) MigrateUser() error {
	query := `
        CREATE TABLE IF NOT EXISTS user (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL UNIQUE,
            mail TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        );`

	_, err := c.DB.Exec(query)
	return err
}

func (c *Chat) UpdateTopic() error {
	_, err := c.DB.Exec("DELETE FROM chat")
	if err != nil {
		return err
	}

	_, err1 := c.DB.Exec("INSERT INTO chat(topic, showLiked, showDisliked, showCreated) values(?, ?, ?, ?)", c.Topic, c.ShowLiked, c.ShowDisliked, c.ShowCreated)
	if err1 != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return ErrDuplicate
			}
		}
		return err1
	}

	return nil
}

func UpdateTopicPostHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("sqlite3", "data/db.db")
	if err != nil {
		log.Fatal(err)
	}
	chat := new(Chat)
	chat.DB = db

	chat.Topic = r.FormValue("topicChoose")
	chat.ShowLiked = r.FormValue("likedCheckbox")
	chat.ShowDisliked = r.FormValue("dislikedCheckbox")
	chat.ShowCreated = r.FormValue("createdCheckbox")

	chat.UpdateTopic()

	http.Redirect(w, r, "/chat", http.StatusFound)
}

func (c *Chat) ChargeTopic() error {
	row := c.DB.QueryRow("SELECT * FROM chat")

	if err := row.Scan(&c.Topic, &c.ShowLiked, &c.ShowDisliked, &c.ShowCreated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotExists
		}
		return err
	}
	return nil
}
