package service

import (
	"context"
	"fmt"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"time"
)

type Service interface {
	Create(ctx context.Context, dto user.CreateUserDTO) (string, error)
	GetByEmailAndPassword(ctx context.Context, email string, password string) (user.User, error)
	GetById(ctx context.Context, id string) (user.User, error)
}

type service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) Service {
	return service{
		storage: storage,
	}
}

func emailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s service) Create(ctx context.Context, dto user.CreateUserDTO) (string, error) {
	if !emailValid(dto.Email) {
		return "", fmt.Errorf("invalid email address")
	}
	hash, err := hashPassword(dto.Password)
	if err != nil {
		return "", fmt.Errorf("could not hash password: %w", err)
	}
	u := user.User{
		Email:        dto.Email,
		Password:     hash,
		RegisteredAt: time.Now(),
	}
	id, err := s.storage.Create(ctx, u)
	if err != nil {
		return "", fmt.Errorf("could not create user: %w", err)
	}
	return id, nil
}

func verifyPassword(hashed string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

func (s service) GetByEmailAndPassword(ctx context.Context, email string, password string) (u user.User, err error) {
	u, err = s.storage.FindByEmail(ctx, email)
	if err != nil {
		return
	}
	if !verifyPassword(u.Password, password) {
		err = fmt.Errorf("wrong password")
		return
	}
	return
}

func (s service) GetById(ctx context.Context, id string) (u user.User, err error) {
	u, err = s.storage.FindById(ctx, id)
	return
}
