package models

import (
	"database/sql"
)

type UserData struct {
	UserID         string `json:"id"`
	PassportSeries string `json:"passport_series"`
	PassportNumber string `json:"passport_number"`
	Surname        string `json:"surname"`
	Name           string `json:"name"`
	Patronymic     string `json:"patronymic"`
	Address        string `json:"address"`
}

type PassportRequest struct {
	PassportNumber string `json:"passportNumber"`
}

type TaskName struct {
	Name string `json:"task_name"`
}

type TaskData struct {
	TaskID    string       `json:"id"`
	UserID    string       `json:"user_id"`
	NameTask  string       `json:"name_task"`
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
	AllTime   sql.NullTime `json:"all_time"`
}

type TaskTime struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Tasks struct {
	Name    string `json:"task_name"`
	AllTime string `json:"all_time"`
}
