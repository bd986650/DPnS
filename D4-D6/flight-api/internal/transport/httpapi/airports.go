package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"flightapi/internal/ctxlang"
	"flightapi/internal/model"
)

func parseOptionalRFC3339(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func scheduleQueryFromRequest(c *gin.Context) (model.ScheduleQuery, error) {
	var sq model.ScheduleQuery
	if v := c.Query("from"); v != "" {
		t, err := parseOptionalRFC3339(v)
		if err != nil {
			return sq, err
		}
		sq.From = t
	}
	if v := c.Query("to"); v != "" {
		t, err := parseOptionalRFC3339(v)
		if err != nil {
			return sq, err
		}
		sq.To = t
	}
	sq.Timezone = c.Query("timezone")
	sq.Status = c.Query("status")
	sq.Airline = c.Query("airline")
	sq.FlightNumber = c.Query("flightNumber")
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 100 {
			sq.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			sq.Offset = n
		}
	}
	return sq, nil
}

// GET /airports
func (h *Handler) handleListAirports(c *gin.Context) {
	ctx := ctxlang.With(c.Request.Context(), ctxlang.ParseAcceptLanguage(c.GetHeader("Accept-Language")))
	role := c.Query("role")
	airports, err := h.repo.ListAirports(ctx, role)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, airports)
}

// GET /airports/{airportCode}/schedule/inbound
func (h *Handler) handleInboundSchedule(c *gin.Context) {
	code := c.Param("airportCode")
	if code == "" {
		writeError(c, http.StatusBadRequest, "airportCode is required")
		return
	}
	sq, err := scheduleQueryFromRequest(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid from/to datetime")
		return
	}
	ctx := ctxlang.With(c.Request.Context(), ctxlang.ParseAcceptLanguage(c.GetHeader("Accept-Language")))
	entries, err := h.repo.GetInboundSchedule(ctx, code, sq)
	if err == model.ErrAirportNotFound {
		writeError(c, http.StatusNotFound, "airport not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, entries)
}

// GET /airports/{airportCode}/schedule/outbound
func (h *Handler) handleOutboundSchedule(c *gin.Context) {
	code := c.Param("airportCode")
	if code == "" {
		writeError(c, http.StatusBadRequest, "airportCode is required")
		return
	}
	sq, err := scheduleQueryFromRequest(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid from/to datetime")
		return
	}
	ctx := ctxlang.With(c.Request.Context(), ctxlang.ParseAcceptLanguage(c.GetHeader("Accept-Language")))
	entries, err := h.repo.GetOutboundSchedule(ctx, code, sq)
	if err == model.ErrAirportNotFound {
		writeError(c, http.StatusNotFound, "airport not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, entries)
}
