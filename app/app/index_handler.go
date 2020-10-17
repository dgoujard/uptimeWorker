package app

import (
	"net/http"
)

func (app *App) HandleIndex(w http.ResponseWriter, _ *http.Request)  {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write([]byte("World !"))

	w.WriteHeader(http.StatusOK)
}
