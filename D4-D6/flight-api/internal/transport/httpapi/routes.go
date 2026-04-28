package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"flightapi/internal/ctxlang"
	"flightapi/internal/helper"
	"flightapi/internal/model"
)

// GET /routes
func (h *Handler) handleListRoutes(c *gin.Context) {
	originType := c.Query("originType")
	origin := c.Query("origin")
	destType := c.Query("destinationType")
	dest := c.Query("destination")
	departureDateStr := c.Query("departureDate")
	bookingClass := c.Query("bookingClass")
	maxConnectionsStr := c.Query("maxConnections")

	if originType == "" || origin == "" || destType == "" || dest == "" || departureDateStr == "" {
		writeError(c, http.StatusBadRequest, "missing required parameters")
		return
	}

	if originType != "airport" && originType != "city" {
		writeError(c, http.StatusBadRequest, "originType must be 'airport' or 'city'")
		return
	}
	if destType != "airport" && destType != "city" {
		writeError(c, http.StatusBadRequest, "destinationType must be 'airport' or 'city'")
		return
	}

	departureDate, err := time.Parse("2006-01-02", departureDateStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid departureDate, expected YYYY-MM-DD")
		return
	}

	if bookingClass != "" && !helper.ContainsString([]string{"Economy", "Comfort", "Business"}, bookingClass) {
		writeError(c, http.StatusBadRequest, "invalid bookingClass")
		return
	}

	maxConnections := -1
	if maxConnectionsStr != "" {
		val, err := strconv.Atoi(maxConnectionsStr)
		if err != nil || val < 0 || val > 3 {
			writeError(c, http.StatusBadRequest, "maxConnections must be one of 0,1,2,3")
			return
		}
		maxConnections = val
	}

	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 100 {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	tz := c.Query("timezone")

	ctx := ctxlang.With(c.Request.Context(), ctxlang.ParseAcceptLanguage(c.GetHeader("Accept-Language")))
	originAirports, err := h.resolvePointToAirports(ctx, originType, origin)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid origin")
		return
	}
	destAirports, err := h.resolvePointToAirports(ctx, destType, dest)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid destination")
		return
	}

	routes, err := h.repo.FindRoutes(ctx, originAirports, destAirports, model.RouteSearchParams{
		DepartureDate:  departureDate,
		BookingClass:   bookingClass,
		MaxConnections: maxConnections,
		Timezone:       tz,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, routes)
}

func (h *Handler) resolvePointToAirports(ctx context.Context, pointType, value string) ([]string, error) {
	if pointType == "airport" {
		return []string{strings.ToUpper(value)}, nil
	}

	airports, err := h.repo.ListAirportsByCity(ctx, value)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(airports))
	for _, a := range airports {
		res = append(res, a.Code)
	}
	return res, nil
}
