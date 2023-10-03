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

type Answer struct {
	ID         int64
	IDanswer   int64
	Writer     string
	Content    string
	Like       int64
	Dislike    int64
	UsersLiked []string
}

// function that migrates the table answers to the db
func (c *Chat) MigrateAnswer() error {
	query := `
    CREATE TABLE IF NOT EXISTS answers(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		idAnswer INTEGER,
		writer TEXT NOT NULL,
        content TEXT NOT NULL,
		like INTEGER,
		dislike INTEGER
    );`

	_, err := c.DB.Exec(query)
	return err
}

// function that creates an answer by writting it with all the needed information
func (c *Chat) CreateAnswer(content string, id int64) error {
	_, err := c.DB.Exec("INSERT INTO answers(idAnswer, writer, content, like, dislike) values(?, ?, ?, ?, ?)", id, c.Username, content, 0, 0)
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

// answer to a post
func (c *Chat) AllAnswers(idAnswer int64) ([]Answer, error) {
	rows, err := c.DB.Query("SELECT * FROM answers WHERE idAnswer = ?", idAnswer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Answer
	for rows.Next() {
		var answer Answer
		if err := rows.Scan(&answer.ID, &answer.IDanswer, &answer.Writer, &answer.Content, &answer.Like, &answer.Dislike); err != nil {
			return nil, err
		}
		all = append(all, answer)
	}
	return all, nil
}

// function that deletes a post and all the likes and dislikes associated with it
func (c Chat) DeleteAnswer(id int64) error {
	res, err := c.DB.Exec("DELETE FROM answers WHERE id = ?", id)
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
	res, err = c.DB.Exec("DELETE FROM "+c.Username+"WHERE idOfPost = ? AND typeOfPost = ?", id, "posts")
	if err != nil {
		return err
	}
	return err
}

// funtcion that deletes an answer to a post and all the likes and dislikes associated with it in the db
func AnswerDeletePostHandler(w http.ResponseWriter, r *http.Request) {
	answerIDstr := r.FormValue("id")
	answerID, _ := strconv.Atoi(answerIDstr)
	fmt.Println(answerID)

	c := new(Chat)
	db, err := sql.Open("sqlite3", "data/db.db")
	if err != nil {
		log.Fatal(err)
	}
	c.DB = db
	c.DeleteAnswer(int64(answerID))

	http.Redirect(w, r, "/chat", http.StatusFound)
}
