package usecase

import "time-tracker/internal/models"

type UseCaseStorage interface {
	UseCaseCreate(userData models.UserData) (int, error)
	UseCaseRead(userID int) (models.UserData, error)
	UseCaseUpdate(userID int, userData models.UserData) error
	UseCaseDelete(userID int) error
	UseCaseGetUsers(dataUser models.UserData, page, limit int) ([]models.UserData, error)
	UseCaseCreateTask(userID int, nameTask string) (int, error)
	UseCaseReadTask(taskID int) (models.TaskData, error)
	UseCaseAddStartTime(taskID int) error
	UseCaseAddEndTime(taskID int) error
	UseCaseGetTasksUser(userID int, timeTask models.TaskTime) ([]models.Tasks, error)
}
