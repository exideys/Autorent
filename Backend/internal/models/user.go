package models

import (
	"strings"
	"time"
)

const (
	UserRoleUser  = "user"
	UserRoleAdmin = "admin"
)

type User struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Rating      float64   `json:"rating"`
	RatingCount int       `json:"rating_count"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserWithPassword struct {
	User
	PasswordHash string
}

type GoogleUserInput struct {
	GoogleSub string
	FirstName string
	LastName  string
	Email     string
}

type RegisterInput struct {
	Name      string `json:"name" binding:"omitempty,max=100"`
	FirstName string `json:"first_name" binding:"omitempty,max=50"`
	LastName  string `json:"last_name" binding:"omitempty,max=50"`
	Email     string `json:"email" binding:"required,email,max=100"`
	Password  string `json:"password" binding:"required,min=8,max=72"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required"`
}

type GoogleLoginInput struct {
	Credential string `json:"credential" binding:"required"`
}

type UpdateCurrentUserInput struct {
	FirstName string `json:"first_name" binding:"required,max=50"`
	LastName  string `json:"last_name" binding:"required,max=50"`
	Email     string `json:"email" binding:"required,email,max=100"`
}

type RateUserInput struct {
	Rating float64 `json:"rating" binding:"required,gte=1,lte=5"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

func (input RegisterInput) HasName() bool {
	return strings.TrimSpace(input.Name) != "" ||
		strings.TrimSpace(input.FirstName) != "" ||
		strings.TrimSpace(input.LastName) != ""
}

func (input RegisterInput) NameParts() (string, string) {
	firstName := strings.TrimSpace(input.FirstName)
	lastName := strings.TrimSpace(input.LastName)

	if firstName == "" && lastName == "" {
		parts := strings.Fields(input.Name)
		if len(parts) > 0 {
			firstName = parts[0]
		}
		if len(parts) > 1 {
			lastName = strings.Join(parts[1:], " ")
		}
	}

	return firstName, lastName
}

func (input RegisterInput) DisplayName() string {
	firstName, lastName := input.NameParts()
	return strings.TrimSpace(strings.Join([]string{firstName, lastName}, " "))
}
