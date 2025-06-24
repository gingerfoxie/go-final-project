package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	DATE_FORMAT = "20060102"
)

type errResp struct {
	Error string `json:"error"`
}

func Init() {
	http.HandleFunc("/api/nextdate", NextStartDateHandler)
	http.HandleFunc("/api/task", TaskHandler)
	http.HandleFunc("/api/tasks", TasksHandler)
	http.HandleFunc("/api/task/done", TaskDoneHandler)
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
	w.Write(d)

}
