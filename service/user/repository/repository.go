package repository

import "gorm.io/gorm"

type Repositories interface {
	UserQuery() UserQueryRepository
	UserCommand() UserCommandRepository
	Role() RoleRepository
}

type repositories struct {
	userQuery   UserQueryRepository
	userCommand UserCommandRepository
	role        RoleRepository
}

func NewRepositories(db *gorm.DB) Repositories {
	return &repositories{
		userCommand: NewUserCommandRepository(db),
		userQuery:   NewUserQueryRepository(db),
		role:        NewRoleRepository(db),
	}
}

func (r *repositories) UserQuery() UserQueryRepository {
	return r.userQuery
}
func (r *repositories) UserCommand() UserCommandRepository {
	return r.userCommand
}
func (r *repositories) Role() RoleRepository {
	return r.role
}
