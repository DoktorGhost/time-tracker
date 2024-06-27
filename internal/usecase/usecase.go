package usecase

import "time-tracker/internal/storage"

type UseCaseStorage struct {
	storage storage.RepositoryDB
}

func NewUseCaseStorage(storage storage.RepositoryDB) *UseCaseStorage {
	return &UseCaseStorage{storage: storage}
}

func (uc *UseCaseStorage) UseCaseCreate(name string) error {
	return uc.storage.Create(name)
}

func (uc *UseCaseStorage) UseCaseRead(name string) (string, error) {
	return uc.storage.Read(name)
}

func (uc *UseCaseStorage) UseCaseUpdate(name string) error {
	return uc.storage.Update(name)
}

func (uc *UseCaseStorage) UseCaseDelete(name string) error {
	return uc.storage.Delete(name)
}
