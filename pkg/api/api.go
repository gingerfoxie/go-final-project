package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go1f/pkg/db"
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
		return "", errors.New("wrong start date format" + dstart)
	}

	rules := strings.Split(repeat, " ")

	if len(rules) == 0 {
		return "", errors.New("wrong repeat rule format" + repeat)
	}

	switch rules[0] {

	case "d":
		if len(rules) == 1 {
			return "", errors.New("the number of days is not set")
		}
		interval, err := strconv.Atoi(rules[1])
		if err != nil {
			return "", errors.New("the number of days is incorrect")
		}
		if interval > 400 {
			return "", errors.New("the number of days more than 400")
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

	case "w":
		if len(rules) < 2 {
			return "", fmt.Errorf("wrong repeat rule format (%s)", repeat)
		}

		weekDays := [7]bool{}
		weekRules := strings.Split(rules[1], ",")

		for _, value := range weekRules {
			num, err := strconv.Atoi(value)
			if err != nil {
				return "", fmt.Errorf("week day num is incorrect (%s)", value)
			}
			if num < 1 || num > 7 {
				return "", fmt.Errorf("week day num is incorrect (%s)", value)
			}
			if num == 7 {
				weekDays[0] = true
			} else {
				weekDays[num] = true
			}
		}

		for {
			date = date.AddDate(0, 0, 1)
			if weekDays[date.Weekday()] {
				if AfterNow(now, date) {
					return date.Format(DATE_FORMAT), nil
				}
			}

		}
		//return "", fmt.Errorf("wrong repeat rule format (%s)", repeat)

	case "m":
		if len(rules) < 2 {
			return "", fmt.Errorf("wrong repeat rule format (%s)", repeat)
		}

		months := []int{}
		if len(rules) == 2 {
			// every month
			for i := 1; i < 13; i++ {
				months = append(months, i)
			}
		} else {
			mRules := strings.Split(rules[2], ",")
			for _, value := range mRules {
				num, err := strconv.Atoi(value)
				if err != nil {
					return "", fmt.Errorf("month num is incorrect (%s)", value)
				}
				if num < 1 || num > 12 {
					return "", fmt.Errorf("month num is incorrect (%s)", value)
				}

				months = append(months, num)
			}

		}

		days := []int{}
		dRules := strings.Split(rules[1], ",")

		for _, value := range dRules {
			num, err := strconv.Atoi(value)
			if err != nil {
				return "", fmt.Errorf("day num is incorrect (%s)", value)
			}
			if num < -2 || num > 31 {
				return "", fmt.Errorf("day num is incorrect (%s)", value)
			}

			days = append(days, num)
		}

		dates := []string{}

		for year := now.Year(); year <= now.Year()+1; year++ {
			for _, month := range months {
				begMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, date.UTC().Location())
				endMonth := begMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)
				var newDate time.Time
				for _, day := range days {
					if day > 0 {
						newDate = begMonth.AddDate(0, 0, day-1)
					} else {
						newDate = endMonth.AddDate(0, 0, day+1)
					}
					if time.Month(month) == newDate.Month() && newDate.Format(DATE_FORMAT) >= date.Format(DATE_FORMAT) {
						dates = append(dates, newDate.Format(DATE_FORMAT))
					}
				}
			}

		}

		sort.Strings(dates)

		for _, CheckDate := range dates {
			dateParam, err := time.Parse(DATE_FORMAT, CheckDate)
			if err == nil {
				if AfterNow(now, dateParam) {
					return CheckDate, nil
				}
			}
		}
		return "", fmt.Errorf("wrong repeat rule format (%s)", repeat)

	default:
		return "", errors.New("the specified repeat rule is not supported")
	}

}
