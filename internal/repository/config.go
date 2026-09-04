package repository

import "os"

func NewConfiguredRepository() Repository {
	if os.Getenv("REPOSITORY_DRIVER") == "postgres" {
		if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
			if repo, err := OpenPostgresRepository(dsn); err == nil {
				return repo
			}
		}
	}
	return NewInMemoryRepository()
}
