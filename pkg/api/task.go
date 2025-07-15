package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go1f/pkg/db"
)

type tasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

type addTaskResp struct {
	ID string `json:"id"`
}
type doneResp struct {
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

func TasksHandler(w http.ResponseWriter, req *http.Request) {

	var tasks []*db.Task
	var err error

	searchStr := req.URL.Query().Get("search")

	if len(searchStr) > 0 {

		dateRegex := `^\d{2}.\d{2}.\d{4}$`
		match, _ := regexp.MatchString(dateRegex, searchStr)
		if match {
			dateMap := strings.Split(searchStr, ".")
			dateStr := fmt.Sprintf("%s%s%s", dateMap[2], dateMap[1], dateMap[0])

			tasks, err = db.TaskList("", dateStr, 50)
			if err != nil {
				writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
				return
			}
		} else {
			tasks, err = db.TaskList(searchStr, "", 50)
			if err != nil {
				writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
				return
			}
		}

	} else {

		tasks, err = db.TaskList("", "", 50)
		if err != nil {
			writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
			return
		}
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
		writeJson(w, errResp{Error: "empty title"}, http.StatusBadRequest)
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
		writeJson(w, errResp{Error: "empty title"}, http.StatusBadRequest)
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
