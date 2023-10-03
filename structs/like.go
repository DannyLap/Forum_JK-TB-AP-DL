package structs

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

type Like struct {
	PostsLiked      []int64
	AnswersLiked    []int64
	PostsDisliked   []int64
	AnswersDisliked []int64
}

// funtcion that handles the like and dislike of a post
func LikeDislikePostHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.Values["username"] == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		c := new(Chat)
		c.Username = session.Values["username"].(string)
		db, err := sql.Open("sqlite3", "data/db.db")
		if err != nil {
			log.Fatal(err)
		}
		c.DB = db

		typeOfPost := r.FormValue("typeOfPost")
		typeOfLike := r.FormValue("typeOfLike")
		like, _ := strconv.Atoi(r.FormValue("like"))
		update, _ := strconv.Atoi(r.FormValue("update"))
		id, _ := strconv.Atoi(r.FormValue("id"))
		c.UpdateLike(int64(id), int64(like), int64(update), typeOfPost, typeOfLike)
		http.Redirect(w, r, "/chat", http.StatusFound)
	}
}

// migration of the like table to the db
func (h *Home) MigrateLikes(username string) error {
	query := `
    CREATE TABLE IF NOT EXISTS ` + username + ` ( 
		idOfPost INTEGER,
		typeOfPost TEXT NOT NULL,
		typeOfLike TEXT NOT NULL
    );`

	_, err := h.DB.Exec(query)
	return err
}

// update like and dislike of a post
func (c *Chat) UpdateLike(id int64, like int64, update int64, typeOfPost string, typeOfLike string) error {
	if id == 0 {
		return errors.New("invalid updated ID")
	}
	res, err := c.DB.Exec("UPDATE "+typeOfPost+" SET "+typeOfLike+" = ? WHERE id = ?", like+update, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUpdateFailed
	}
	if update == 1 {
		_, err := c.DB.Exec("INSERT INTO "+c.Username+"(idOfPost, typeOfLike, typeOfPost) values(?, ?, ?)", id, typeOfLike, typeOfPost)
		if err != nil {
			return err
		}
		return nil
	} else {
		res, err := c.DB.Exec("DELETE FROM "+c.Username+" WHERE typeOfLike = ? AND typeOfPost = ? AND idOfPost = ?", typeOfLike, typeOfPost, id)
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
	}
	return nil
}

// if you are not connected to the chat you can't like or dislike a post
func (c *Chat) AllAboutLike(typeOfLike string, typeOfPost string) ([]int64, error) {
	rows, err := c.DB.Query("SELECT idOfPost FROM "+c.Username+" WHERE typeOfLike = ? AND typeOfPost = ?", typeOfLike, typeOfPost)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var all []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		all = append(all, id)
	}
	return all, nil
}

// function to show all the likes and dislikes of a post
func (c *Chat) AllDataAboutLike() {
	for _, typeOfPost := range []string{"posts", "answers"} {
		for _, typeOfLike := range []string{"like", "dislike"} {
			all, err := c.AllAboutLike(typeOfLike, typeOfPost)

			if err != nil {
				log.Fatal(err)
			}

			switch typeOfPost {
			case "posts":
				switch typeOfLike {
				case "like":
					c.Like.PostsLiked = all
				case "dislike":
					c.Like.PostsDisliked = all
				}
			case "answers":
				switch typeOfLike {
				case "like":
					c.Like.AnswersLiked = all
				case "dislike":
					c.Like.AnswersDisliked = all
				}
			}
		}
	}
}
