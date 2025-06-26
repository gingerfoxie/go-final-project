package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	DATE_FORMAT = "20060102"
)

type errResp struct {
	Error string `json:"error"`
}

var Pass string

func Init() {
	http.HandleFunc("/api/nextdate", NextStartDateHandler)
	http.HandleFunc("/api/task", auth(TaskHandler))
	http.HandleFunc("/api/tasks", auth(TasksHandler))
	http.HandleFunc("/api/task/done", auth(TaskDoneHandler))
	http.HandleFunc("/api/signin", SigninHandler)

	Pass = os.Getenv("TODO_PASSWORD")
}

func writeJson(w http.ResponseWriter, data any, status int) {

	d, err := json.Marshal(data)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusInternalServerError)
		fmt.Println("err")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_, err = w.Write(d)

	if err != nil {
		log.Printf("Write response failed: %v", err)
	}
}
