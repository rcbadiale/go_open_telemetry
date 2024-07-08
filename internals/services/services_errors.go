package services

import "errors"

var (
	ErrInvalidCEP  = errors.New("invalid CEP provided")
	ErrCEPNotFound = errors.New("CEP not found")
)
