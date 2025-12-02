package page_builder

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/owner/go-cms/pkg/errors"
	"gorm.io/gorm"
)

// WrapError wraps a database error into an AppError
func WrapError(err error, defaultMsg string) error {
	if err == nil {
		return nil
	}

	if err == gorm.ErrRecordNotFound {
		return errors.Wrap(err, errors.ErrCodeNotFound, "Resource not found", http.StatusNotFound)
	}

	return errors.Wrap(err, errors.ErrCodeInternal, defaultMsg, http.StatusInternalServerError)
}

// ParseUUID parses a UUID string and returns an error if invalid
func ParseUUID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(err, errors.ErrCodeBadRequest, "Invalid UUID format", http.StatusBadRequest)
	}
	return parsed, nil
}
