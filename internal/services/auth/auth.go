package auth

import (
	"auth/internal/domain/models"
	jwtutils "auth/internal/lib/jwt-utils"
	"auth/internal/lib/logger/sl"
	"auth/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	GetUser(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	GetApp(ctx context.Context, appID int) (models.App, error)
}

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokentTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokentTTL,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const operation = "Auth.RegisterNewUser"

	log := a.log.With(slog.String("operation", operation), slog.String("email", email))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return id, nil
}

func (a *Auth) LoginUser(ctx context.Context, email string, password string, appID int) (string, error) {
	const operation = "Auth.Login"

	log := a.log.With(slog.String("operation", operation), slog.String("username", email))
	log.Info("attempting to login user")

	user, err := a.userProvider.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", operation, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", operation, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))
		return "", fmt.Errorf("%s: %w", operation, ErrInvalidCredentials)
	}

	app, err := a.appProvider.GetApp(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", operation, err)
	}

	log.Info("user logged in successfully")

	token, err := jwtutils.CreateNewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", operation, err)
	}

	return token, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const operation = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("operation", operation),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
