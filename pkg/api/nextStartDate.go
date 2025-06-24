package api

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

		return nextDateDayly(rules, date, now)

	case "y":

		return nextDateYearly(date, now)

	case "w":

		return nextDateWeekly(rules, date, now, repeat)

	case "m":

		return newDateMonthly(rules, date, now, repeat)

	default:
		return "", errors.New("the specified repeat rule is not supported")
	}

}

func nextDateDayly(rules []string, date time.Time, now time.Time) (string, error) {

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

}

func nextDateWeekly(rules []string, date time.Time, now time.Time, repeat string) (string, error) {

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

}

func newDateMonthly(rules []string, date time.Time, now time.Time, repeat string) (string, error) {

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

}

func nextDateYearly(date time.Time, now time.Time) (string, error) {

	for {
		date = date.AddDate(1, 0, 0)
		if AfterNow(now, date) {
			return date.Format(DATE_FORMAT), nil
		}
	}

}
