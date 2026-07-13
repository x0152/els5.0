package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

type EchoInput struct {
	authx.BearerInput
	Body EchoBody
}

type EchoBody struct {
	Message string `json:"message" minLength:"1" maxLength:"500" doc:"Arbitrary text echoed back after normalization"`
}

type EchoOutput struct {
	Message string `json:"message"`
}
