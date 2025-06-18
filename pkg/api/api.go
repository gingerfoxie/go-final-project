package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gingerfoxie/go-final-project/pkg/db"
)

const (
	DATE_FORMAT = "20060102"
)

type errResp struct {
	Error string `json:"error"`
}

type tasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

type addTaskResp struct {
	ID string `json:"id"`
}
type doneResp struct {
}

func Init() {
	http.HandleFunc("/api/nextdate", NextStartDateHandler)
	http.HandleFunc("/api/task", TaskHandler)
	http.HandleFunc("/api/tasks", TasksHandler)
	http.HandleFunc("/api/task/done", TaskDoneHandler)
}

func TaskHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		GetTaskHandler(w, req)
	case http.MethodPost:
		AddTaskHandler(w, req)
	case http.MethodPut:
		UpdateTaskHandler(w, req)
	case http.MethodDelete:
		DeleteTaskHandler(w, req)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func NextStartDateHandler(w http.ResponseWriter, req *http.Request) {

	now := time.Now()
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	nowParam := req.FormValue("now")

	if nowParam != "" {
		nowDateParam, err := time.Parse(DATE_FORMAT, nowParam)
		if err != nil {
			writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
			return
		}
		nowDate = nowDateParam
	}

	dstart := req.FormValue("date")
	repeat := req.FormValue("repeat")

	nextdate, err := NextStartDate(nowDate, dstart, repeat)

	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
	} else {
		w.Write([]byte(nextdate))
	}
}

func TasksHandler(w http.ResponseWriter, req *http.Request) {

	tasks, err := db.TaskList(50)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJson(w, tasksResp{
		Tasks: tasks,
	}, http.StatusOK)

}

func AddTaskHandler(w http.ResponseWriter, req *http.Request) {

	var task db.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	// проверка корректности данных
	if !CheckTitle(&task) {
		writeJson(w, errResp{Error: "заголовок задачи не может быть пустым"}, http.StatusBadRequest)
		return
	}

	res, err := CheckDate(&task)
	if !res {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	_, err = NextStartDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)

	if err == nil {
		writeJson(w, addTaskResp{ID: fmt.Sprintf("%d", id)}, http.StatusOK)
	} else {
		writeJson(w, errResp{Error: err.Error()}, http.StatusInternalServerError)
	}

}

func UpdateTaskHandler(w http.ResponseWriter, req *http.Request) {

	var task db.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	// проверка корректности данных
	if !CheckTitle(&task) {
		writeJson(w, errResp{Error: "заголовок задачи не может быть пустым"}, http.StatusBadRequest)
		return
	}

	res, err := CheckDate(&task)
	if !res {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	_, err = NextStartDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	err = db.UpdateTask(&task)

	if err == nil {
		writeJson(w, doneResp{}, http.StatusOK)
	} else {
		writeJson(w, errResp{Error: err.Error()}, http.StatusInternalServerError)
	}

}

func DeleteTaskHandler(w http.ResponseWriter, req *http.Request) {

	idParam := req.URL.Query().Get("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	_, err = db.GetTask(id)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	err = db.DeleteTask(id)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJson(w, doneResp{}, http.StatusOK)

}

func GetTaskHandler(w http.ResponseWriter, req *http.Request) {

	idParam := req.URL.Query().Get("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJson(w, task, http.StatusOK)

}

func TaskDoneHandler(w http.ResponseWriter, req *http.Request) {

	idParam := req.URL.Query().Get("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
			return
		}
	} else {
		now := time.Now()
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		nextDate, err := NextStartDate(nowDate, task.Date, task.Repeat)
		if err != nil {
			writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
			return
		}
		task.Date = nextDate
		err = db.UpdateDate(task)
		if err != nil {
			writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
			return
		}
	}

	writeJson(w, doneResp{}, http.StatusOK)

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

func NextStartDate(now time.Time, dstart string, repeat string) (string, error) {

	if repeat == "" {
		return "", nil
	}

	date, err := time.Parse(DATE_FORMAT, dstart)
	if err != nil {
		return "", errors.New("неверный формат даты отсчета" + dstart)
	}

	rules := strings.Split(repeat, " ")

	if len(rules) == 0 {
		return "", errors.New("неверно задано правило повторения" + repeat)
	}

	switch rules[0] {

	case "d":
		if len(rules) == 1 {
			return "", errors.New("не задано количество дней")
		}
		interval, err := strconv.Atoi(rules[1])
		if err != nil {
			return "", errors.New("указанное некорректно количество дней")
		}
		if interval > 400 {
			return "", errors.New("максимально допустимое число дней - 400")
		}

		for {
			date = date.AddDate(0, 0, interval)
			if AfterNow(now, date) {
				return date.Format(DATE_FORMAT), nil
			}
		}

	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if AfterNow(now, date) {
				return date.Format(DATE_FORMAT), nil
			}
		}

	default:
		return "", errors.New("указанное правило повторения не поддерживается")
	}

}
