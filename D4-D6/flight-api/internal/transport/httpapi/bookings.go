package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"flightapi/internal/helper"
	"flightapi/internal/model"
)

// POST /bookings
func (h *Handler) handleCreateBooking(c *gin.Context) {
	var req model.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.RouteID == "" || req.BookingClass == "" {
		writeError(c, http.StatusBadRequest, "routeId and bookingClass are required")
		return
	}
	if !helper.ContainsString([]string{"Economy", "Comfort", "Business"}, req.BookingClass) {
		writeError(c, http.StatusBadRequest, "invalid bookingClass")
		return
	}
	if req.Passenger.FirstName == "" || req.Passenger.LastName == "" || req.Passenger.DateOfBirth == "" || req.Passenger.DocumentNumber == "" {
		writeError(c, http.StatusBadRequest, "passenger fields are incomplete")
		return
	}

	booking, err := h.repo.CreateBooking(c.Request.Context(), req)
	if err == model.ErrRouteNotFound {
		writeError(c, http.StatusBadRequest, "route not found")
		return
	}
	if err == model.ErrNoSeatsAvailable {
		writeError(c, http.StatusConflict, "no seats available")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal error")
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// POST /bookings/{bookingId}/checkin
func (h *Handler) handleCheckIn(c *gin.Context) {
	bookingID := c.Param("bookingId")
	if bookingID == "" {
		writeError(c, http.StatusBadRequest, "bookingId is required")
		return
	}
	var req model.CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.FlightNo == "" {
		writeError(c, http.StatusBadRequest, "flightNo is required")
		return
	}

	bp, err := h.repo.CheckIn(c.Request.Context(), bookingID, req)
	if err == model.ErrBookingNotFound {
		writeError(c, http.StatusNotFound, "booking not found")
		return
	}
	if err == model.ErrCheckInNotAvailable {
		writeError(c, http.StatusBadRequest, "check-in not available for this flight")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal error")
		return
	}

	c.JSON(http.StatusOK, bp)
}
