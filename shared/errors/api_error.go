package errors

import (
	"fmt"
)

func InvalidAccessToken() error {
	return fmt.Errorf("invalid access token")
}

func NewBadRequestError(message string) *AppError {
	return ErrBadRequest.WithMessage(message)
}

func NewNotFoundError(resource string) *AppError {
	return ErrNotFound.WithMessage(fmt.Sprintf("%s not found", resource))
}

func NewConflictError(message string) *AppError {
	return ErrConflict.WithMessage(message)
}

func NewInternalError(err error) *AppError {
	return ErrInternal.WithInternal(err)
}

func NewServiceUnavailableError(service string) *AppError {
	return ErrServiceUnavailable.WithMessage(fmt.Sprintf("%s is temporarily unavailable", service))
}

func NewForbiddenError(message string) *AppError {
	return ErrForbidden.WithMessage(message)
}
