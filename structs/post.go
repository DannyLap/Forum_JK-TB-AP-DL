package structs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/mattn/go-sqlite3"
)

type Post struct {
	ID      int64
	Writer  string
	Content string
	Like    int64
	Dislike int64
	Topic   string
	Answers []Answer
}

// MigratePost creates the "posts" table if it doesn't exist
func (c *Chat) MigratePost() error {
	query := `
    CREATE TABLE IF NOT EXISTS posts(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		writer TEXT NOT NULL,
        content TEXT NOT NULL,
		topic TEXT NOT NULL,
		like INTEGER,
		dislike INTEGER
    );`

	_, err := c.DB.Exec(query)
	return err
}

// CreatePost creates a new post with the given content and topic
func (c *Chat) CreatePost(content string, topic string) error {
	_, err := c.DB.Exec("INSERT INTO posts(writer, content, topic, like, dislike) values(?, ?, ?, ?, ?)", c.Username, content, topic, 0, 0)
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

// AllPosts retrieves all posts from the "posts" table
func (c *Chat) AllPosts() ([]Post, error) {
	rows, err := c.DB.Query("SELECT * FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Post
	// Get all posts from the "posts" table
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Writer, &post.Content, &post.Topic, &post.Like, &post.Dislike); err != nil {
			return nil, err
		}
		allAnswers, err := c.AllAnswers(post.ID)
		post.Answers = allAnswers
		if err != nil {
			log.Fatal(err)
		}
		all = append(all, post)
	}
	return all, nil
}

// DeletePost deletes the post with the given ID from the "posts" table
func (c *Chat) DeletePost(id int64) error {
	res, err := c.DB.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}
	//Delete the post in the "posts" table
	res, err = c.DB.Exec("DELETE FROM "+c.Username+"WHERE idOfPost = ? AND typeOfPost = ?", id, "posts")
	if err != nil {
		return err
	}

	return err
}

// PostDeletePostHandler handles the deletion of a post
func PostDeletePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDstr := r.FormValue("id")
	postID, _ := strconv.Atoi(postIDstr)
	fmt.Println(postID)
	
	c := new(Chat)
	db, err := sql.Open("sqlite3", "data/db.db")
	if err != nil {
		log.Fatal(err)
	}
	c.DB = db
	c.DeletePost(int64(postID))

	http.Redirect(w, r, "/chat", http.StatusFound)
}
