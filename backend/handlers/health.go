package handlers

import (
	"net/http"
	"time"

	"dineware-backend/models"
)

// Health returns the system health status.
// @Summary      Health Check
// @Description  Returns 200 OK if service is up and running.
// @Tags         system
// @Produce      json
// @Success      200  {object}  models.HealthResponse
// @Router       /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != "" {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	WriteJSON(w, http.StatusOK, models.HealthResponse{
		Status:    "OK",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
