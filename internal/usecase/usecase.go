package usecase

import (
	"time-tracker/internal/models"
	"time-tracker/internal/storage"
)

type useCaseStorage struct {
	storage storage.RepositoryDB
}

func NewUseCaseStorage(storage storage.RepositoryDB) UseCaseStorage {
	return &useCaseStorage{storage: storage}
}
func (uc *useCaseStorage) UseCaseCreate(userData models.UserData) (int, error) {
	return uc.storage.Create(userData)
}

func (uc *useCaseStorage) UseCaseRead(userID int) (models.UserData, error) {
	return uc.storage.Read(userID)
}

func (uc *useCaseStorage) UseCaseUpdate(userID int, userData models.UserData) error {
	return uc.storage.Update(userID, userData)
}

func (uc *useCaseStorage) UseCaseDelete(userID int) error {
	return uc.storage.Delete(userID)
}

func (uc *useCaseStorage) UseCaseGetUsers(dataUser models.UserData, page, limit int) ([]models.UserData, error) {
	return uc.storage.GetUsers(dataUser, page, limit)
}

func (uc *useCaseStorage) UseCaseCreateTask(userID int, nameTask string) (int, error) {
	return uc.storage.CreateTask(userID, nameTask)
}

func (uc *useCaseStorage) UseCaseReadTask(taskID int) (models.TaskData, error) {
	return uc.storage.ReadTask(taskID)
}
func (uc *useCaseStorage) UseCaseAddStartTime(taskID int) error {
	return uc.storage.AddStartTime(taskID)
}

func (uc *useCaseStorage) UseCaseAddEndTime(taskID int) error {
	return uc.storage.AddEndTime(taskID)
}

func (uc *useCaseStorage) UseCaseGetTasksUser(userID int, timeTask models.TaskTime) ([]models.Tasks, error) {
	return uc.storage.GetTasksUser(userID, timeTask)
}
