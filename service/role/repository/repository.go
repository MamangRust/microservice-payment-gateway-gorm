package repository

import "gorm.io/gorm"

// Repositories is a struct containing role command and query repositories.
type Repositories interface {
	RoleQueryRepository
	RoleCommandRepository
}

type repositories struct {
	RoleQueryRepository
	RoleCommandRepository
}

// NewRepositories creates a new instance of Repositories with the provided
// GORM database client.
func NewRepositories(db *gorm.DB) Repositories {
	return &repositories{
		RoleQueryRepository:   NewRoleQueryRepository(db),
		RoleCommandRepository: NewRoleCommandRepository(db),
	}
}
