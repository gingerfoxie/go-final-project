package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
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

var DBFilePath string

func initPath() error {

	if DBFilePath != "" {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Init path error: %s", err.Error())
	}

	rootPath := filepath.Dir(exe)
	DBFilePath = filepath.Join(rootPath, "scheduler.db")
	return nil

}
func openDB() (*sqlx.DB, error) {

	// dbfile := DBFilePath
	// envFile := os.Getenv("TODO_DBFILE")
	// if len(envFile) > 0 {
	// 	dbfile = envFile
	// }
	initPath()
	//fmt.Println(DBFilePath)
	db, err := sqlx.Connect("sqlite", DBFilePath)
	return db, err
}

func Init() error {

	err := initPath()

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	_, err = os.Stat(DBFilePath)

	var install bool
	if err != nil {
		install = true
	}
	db, err := openDB()
	if err != nil {
		return err
	}

	defer db.Close()

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

func AddTask(task *Task) (int64, error) {

	var id int64

	db, err := openDB()
	if err != nil {
		fmt.Printf("Open database error: %s", err.Error())
		return 0, err
	}

	defer db.Close()

	res, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err

}

func TaskList(limit int16) ([]*Task, error) {

	tasks := []*Task{}

	db, err := openDB()
	if err != nil {
		fmt.Printf("Open database error: %s", err.Error())
		return tasks, err
	}

	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM scheduler ORDER BY date limit %d ", limit))
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

	db, err := openDB()
	if err != nil {
		fmt.Printf("Open database error: %s", err.Error())
		return &task, err
	}

	defer db.Close()

	row := db.QueryRow("SELECT * FROM scheduler where id = ?", id)
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		log.Println(err)
	}

	return &task, err

}

func UpdateTask(task *Task) error {

	db, err := openDB()
	if err != nil {
		fmt.Printf("Open database error: %s", err.Error())
		return err
	}

	defer db.Close()

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	// метод RowsAffected() возвращает количество записей к которым
	// был применена SQL команда
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

	db, err := openDB()
	if err != nil {
		fmt.Printf("Open database error: %s", err.Error())
		return err
	}

	defer db.Close()

	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.ID)
	if err != nil {
		return err
	}
	// метод RowsAffected() возвращает количество записей к которым
	// был применена SQL команда
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

	db, err := openDB()
	if err != nil {
		fmt.Printf("Open database error: %s", err.Error())
		return err
	}

	defer db.Close()

	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	// метод RowsAffected() возвращает количество записей к которым
	// был применена SQL команда
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}
