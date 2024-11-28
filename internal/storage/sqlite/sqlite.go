package sqlite

import (
	"auth/internal/domain/models"
	"auth/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const operation = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const operation = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare(createUserQuery)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr *sqlite.Error

		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
			return 0, fmt.Errorf("%s: %w", operation, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	return id, nil
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) {
	const operation = "storage.sqlite.GetUser"

	stmt, err := s.db.Prepare(getUserQuery)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", operation, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", operation, err)
	}

	return user, nil
}

func (s *Storage) GetApp(ctx context.Context, id int) (models.App, error) {
	const operation = "storage.sqlite.GetApp"

	stmt, err := s.db.Prepare(getAppQuery)
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, id)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", operation, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", operation, err)
	}

	return app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const operation = "storage.sqlite.IsAdmin"

	stmt, err := s.db.Prepare(isAdminQuery)
	if err != nil {
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool

	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", operation, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", operation, err)
	}

	return isAdmin, nil
}
