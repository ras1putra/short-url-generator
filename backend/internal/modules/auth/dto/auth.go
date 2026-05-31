package dto

import (
	"time"

	"urlshortener/internal/repository"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

func MapUserToResponse(user repository.User) UserResponse {
	return UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}
}

func NewAuthResponse(user repository.User, accessToken, refreshToken string) *AuthResponse {
	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         MapUserToResponse(user),
	}
}

func NewAccessTokenResponse(user repository.User, accessToken string) *AuthResponse {
	return &AuthResponse{
		AccessToken: accessToken,
		User:        MapUserToResponse(user),
	}
}
