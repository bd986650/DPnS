package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"flightapi/internal/model"
)

type Handler struct {
	repo model.FlightRepository
}

func NewHandler(repo model.FlightRepository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes регистрирует все маршруты API.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")

	// Cities
	v1.GET("/cities", h.handleListCities)
	v1.GET("/cities/:cityId/airports", h.handleListCityAirports)

	// Airports
	v1.GET("/airports", h.handleListAirports)
	v1.GET("/airports/:airportCode/schedule/inbound", h.handleInboundSchedule)
	v1.GET("/airports/:airportCode/schedule/outbound", h.handleOutboundSchedule)

	// Routes
	v1.GET("/routes", h.handleListRoutes)

	// Bookings & check-in
	v1.POST("/bookings", h.handleCreateBooking)
	v1.POST("/bookings/:bookingId/checkin", h.handleCheckIn)
}

func writeError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func writeMethodNotAllowed(c *gin.Context) {
	writeError(c, http.StatusMethodNotAllowed, "method not allowed")
}
