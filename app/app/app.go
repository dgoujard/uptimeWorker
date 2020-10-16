package app

import (
	"github.com/go-playground/validator/v10"

	"github.com/dgoujard/uptimeWorker/util/logger"
)

const (
	appErrDataCreationFailure    = "data creation failure"
	appErrDataAccessFailure      = "data access failure"
	appErrDataUpdateFailure      = "data update failure"
	appErrJsonCreationFailure    = "json creation failure"
	appErrFormDecodingFailure    = "form decoding failure"
	appErrFormErrResponseFailure = "form error response failure"
)

type App struct {
	logger    *logger.Logger
	validator *validator.Validate
	zones []string
	username string
	password string
}

func New(
	logger *logger.Logger,
	validator *validator.Validate,
) *App {
	return &App{
		logger:    logger,
		validator: validator,
	}
}

func (app *App) Logger() *logger.Logger {
	return app.logger
}