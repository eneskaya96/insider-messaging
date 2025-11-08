package persistence

import (
	"errors"

	apperrors "github.com/eneskaya/insider-messaging/pkg/errors"
	"gorm.io/gorm"
)

func mapGormError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return apperrors.NewNotFoundError("record not found")

	case errors.Is(err, gorm.ErrInvalidTransaction):
		return apperrors.NewDatabaseError(err)

	case errors.Is(err, gorm.ErrInvalidField):
		return apperrors.NewValidationError("invalid field in database operation")

	case errors.Is(err, gorm.ErrDuplicatedKey):
		return apperrors.New(apperrors.ErrorCodeAlreadyExists, "duplicate record")

	case errors.Is(err, gorm.ErrInvalidData):
		return apperrors.NewValidationError("invalid data")

	default:
		return apperrors.NewDatabaseError(err)
	}
}

func checkRowsAffected(db *gorm.DB, expectedMin int64) error {
	if db.Error != nil {
		return mapGormError(db.Error)
	}

	if db.RowsAffected < expectedMin {
		return apperrors.NewNotFoundError("no rows affected, record may not exist or version mismatch")
	}

	return nil
}
