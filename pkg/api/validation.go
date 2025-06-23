package api

import (
	"errors"
	"time"

	"go1f/pkg/db"
)

func AfterNow(now time.Time, date time.Time) bool {
	nowStr := now.Format(DATE_FORMAT)
	dateStr := date.Format(DATE_FORMAT)
	return (dateStr > nowStr)
}

func CheckTitle(task *db.Task) bool {
	return !(task.Title == "")
}

func CheckDate(task *db.Task) (bool, error) {

	now := time.Now().Format(DATE_FORMAT)
	nowDate, _ := time.Parse(DATE_FORMAT, now)

	if task.Date == "" {
		task.Date = now
	}

	t, err := time.Parse(DATE_FORMAT, task.Date)
	if err != nil {
		return false, err
	}

	if !AfterNow(nowDate, t) && (now != task.Date) {
		if len(task.Repeat) == 0 {
			// если правила повторения нет, то берём сегодняшнее число
			task.Date = now
		} else {
			// в противном случае, берём вычисленную ранее следующую дату
			task.Date, err = NextStartDate(nowDate, task.Date, task.Repeat)
			if err != nil {
				return false, errors.New("ошибка расчета даты задачи " + err.Error())
			}
		}
	}

	return true, nil
}

func CheckRepeatRule(repeat string) (bool, error) {

	if repeat == "" {
		return true, nil
	}
	return true, nil
}
