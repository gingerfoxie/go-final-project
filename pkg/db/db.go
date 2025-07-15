package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	schema = `CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT """",
    title VARCHAR NOT NULL DEFAULT """",
	comment TEXT DEFAULT NULL,
	repeat VARCHAR NOT NULL DEFAULT """");
	CREATE INDEX date ON scheduler (date)`
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

var db *sql.DB

func Init() error {

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Init path error: %s", err.Error())
	}

	rootPath := filepath.Dir(exe)
	dbPath := filepath.Join(rootPath, "scheduler.db")

	envDBPath := os.Getenv("TODO_DBFILE")

	if len(envDBPath) > 0 {
		dbPath = envDBPath
	}

	log.Printf("DB path: %s\n", dbPath)

	_, err = os.Stat(dbPath)

	var install bool
	if err != nil {
		install = true
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX
	if install {
		_, err := db.Query(schema)
		if err != nil {
			return err
		}
	}

	return nil

}

func Close() {

	if db != nil {
		db.Close()
	}
}

func AddTask(task *Task) (int64, error) {

	var id int64

	if db == nil {
		err := Init()
		log.Println("Database init (add task)")
		if err != nil {
			log.Printf("Init database error: %s", err.Error())
			return 0, err
		}
	}

	res, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err

}

func TaskList(text string, date string, limit int16) ([]*Task, error) {

	tasks := []*Task{}

	if db == nil {
		err := Init()
		log.Println("Database init (task list)")
		if err != nil {
			log.Printf("Init database error: %s", err.Error())
			return nil, err
		}
	}

	queryStr := fmt.Sprintf("SELECT * FROM scheduler ORDER BY date limit %d ", limit)
	if text != "" {
		queryStr = fmt.Sprintf("SELECT * FROM scheduler WHERE title like '%s' OR comment like '%s' ORDER BY date limit %d", "%"+text+"%", "%"+text+"%", limit)
	} else if date != "" {
		queryStr = fmt.Sprintf("SELECT * FROM scheduler WHERE date = %s ORDER BY date limit %d", date, limit)
	}

	rows, err := db.Query(queryStr)
	if err != nil {
		log.Printf("Retrieving task list error: %s\n", err)
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}

		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Printf("Retrieving task list error: %s\n", err)
			return tasks, err
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Retrieving task list error: %s\n", err)
		return tasks, err
	}

	return tasks, err

}

func GetTask(id int) (*Task, error) {

	var task Task

	if db == nil {
		err := Init()
		log.Println("Database init (get task)")
		if err != nil {
			log.Printf("Init database error: %s", err.Error())
			return nil, err
		}
	}

	row := db.QueryRow("SELECT * FROM scheduler where id = ?", id)
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		log.Println(err)
	}

	return &task, err

}

func UpdateTask(task *Task) error {

	if db == nil {
		err := Init()
		log.Println("Database init (update task)")
		if err != nil {
			log.Printf("Init database error: %s", err.Error())
			return err
		}
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func UpdateDate(task *Task) error {

	if db == nil {
		err := Init()
		log.Println("Database init (update date)")
		if err != nil {
			log.Printf("Init database error: %s", err.Error())
			return err
		}
	}

	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func DeleteTask(id int) error {

	if db == nil {
		err := Init()
		log.Println("Database init (delete task)")
		if err != nil {
			log.Printf("Init database error: %s", err.Error())
			return err
		}
	}

	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil

}
