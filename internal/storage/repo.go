package storage

// Repository представляет интерфейс для работы с хранилищем данных.
type RepositoryDB interface {
	Create(name string) error
	Read(name string) (string, error)
	Update(name string) error
	Delete(name string) error
}
