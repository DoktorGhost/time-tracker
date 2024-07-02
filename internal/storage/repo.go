package storage

import "time-tracker/internal/models"

// Repository представляет интерфейс для работы с хранилищем данных.
type RepositoryDB interface {
	Create(userData models.UserData) (int, error)
	Read(userID int) (models.UserData, error)
	Update(userID int, userData models.UserData) error
	Delete(userID int) error
	GetUsers(dataFilter models.UserData, page, limit int) ([]models.UserData, error)
	CreateTask(userID int, nameTask string) (int, error)
	ReadTask(taskID int) (models.TaskData, error)
	AddStartTime(taskID int) error
	AddEndTime(taskID int) error
	GetTasksUser(userID int, timeTask models.TaskTime) ([]models.Tasks, error)
}
