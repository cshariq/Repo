package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type HealthApiHandler struct {
	db *gorm.DB
}

func NewHealthApiHandler(db *gorm.DB) *HealthApiHandler {
	return &HealthApiHandler{db: db}
}

func (h *HealthApiHandler) RegisterRoutes(router chi.Router) {
	router.Get("/health", h.Get)
}

// @Summary Check the application's health status
// @ID get-health
// @Tags misc
// @Produce plain
// @Success 200 {string} string
// @Router /health [get]
func (h *HealthApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	var dbStatus int
	if sqlDb, err := h.db.DB(); err == nil {
		if err := sqlDb.Ping(); err == nil {
			dbStatus = 1
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	if dbStatus == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(fmt.Sprintf("app=1\ndb=%d", dbStatus)))
}
