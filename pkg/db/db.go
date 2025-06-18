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

var DB *sql.DB

func Init() error {

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Init path error: %s", err.Error())
	}

	rootPath := filepath.Dir(exe)
	dbPath := filepath.Join(rootPath, "scheduler.db")

	_, err = os.Stat(dbPath)

	var install bool
	if err != nil {
		install = true
	}

	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX
	if install {
		_, err := DB.Query(schema)
		if err != nil {
			return err
		}
	}

	return nil

}

func AddTask(task *Task) (int64, error) {

	var id int64

	if DB == nil {
		Init()
		fmt.Println("Database init add")
	}

	res, err := DB.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err

}

func TaskList(limit int16) ([]*Task, error) {

	tasks := []*Task{}

	if DB == nil {
		Init()
		fmt.Println("Database init list")
	}

	rows, err := DB.Query(fmt.Sprintf("SELECT * FROM scheduler ORDER BY date limit %d ", limit))
	if err != nil {
		log.Println(err)
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}

		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Println(err)
			return tasks, err
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return tasks, err
	}

	return tasks, err

}

func GetTask(id int) (*Task, error) {

	var task Task

	if DB == nil {
		Init()
		fmt.Println("Database init get task")
	}

	row := DB.QueryRow("SELECT * FROM scheduler where id = ?", id)
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		log.Println(err)
	}

	return &task, err

}

func UpdateTask(task *Task) error {

	if DB == nil {
		Init()
		fmt.Println("Database init update")
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
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

	if DB == nil {
		Init()
		fmt.Println("Database init update date")
	}

	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.ID)
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

	if DB == nil {
		Init()
		fmt.Println("Database init delete")
	}

	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(query, id)
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
