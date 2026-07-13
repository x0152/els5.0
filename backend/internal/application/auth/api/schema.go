package api

import (
	"time"

	authx "github.com/els/backend/internal/utils/auth"
)

type LoginStartInput struct {
	Body struct {
		Email    string `json:"email" format:"email"`
		Password string `json:"password" minLength:"1" maxLength:"128"`
	}
}

type LoginConfirmInput struct {
	Body struct {
		Token string `json:"token" minLength:"1"`
	}
}

type SetPasswordInput struct {
	Body struct {
		Token           string `json:"token" minLength:"1"`
		Password        string `json:"password" minLength:"1" maxLength:"128"`
		PasswordConfirm string `json:"password_confirm" minLength:"1" maxLength:"128"`
	}
}

type ResendInviteInput struct {
	Body struct {
		Email string `json:"email" format:"email"`
	}
}

type ResendInviteOutput struct {
	Body struct {
		SentTo string `json:"sent_to"`
	}
}

type ForgotPasswordInput struct {
	Body struct {
		Email string `json:"email" format:"email"`
	}
}

type ForgotPasswordOutput struct {
	Body struct {
		SentTo string `json:"sent_to"`
	}
}

type ResetPasswordInput struct {
	Body struct {
		Token           string `json:"token" minLength:"1"`
		Password        string `json:"password" minLength:"1" maxLength:"128"`
		PasswordConfirm string `json:"password_confirm" minLength:"1" maxLength:"128"`
	}
}

type LogoutInput struct {
	authx.BearerInput
}

type ImpersonateInput struct {
	authx.BearerInput
	Body struct {
		AccountID string `json:"account_id" minLength:"1"`
	}
}

type SessionOutput struct {
	AccountID string    `json:"account_id"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PasswordChangedOutput struct {
	Email string `json:"email"`
}

type EmptyOutput struct{}
