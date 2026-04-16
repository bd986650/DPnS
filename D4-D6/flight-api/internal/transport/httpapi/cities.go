package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"flightapi/internal/ctxlang"
	"flightapi/internal/model"
)

// GET /cities
func (h *Handler) handleListCities(c *gin.Context) {
	ctx := ctxlang.With(c.Request.Context(), ctxlang.ParseAcceptLanguage(c.GetHeader("Accept-Language")))
	role := c.Query("role")
	cities, err := h.repo.ListCities(ctx, role)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, cities)
}

// GET /cities/{cityId}/airports
func (h *Handler) handleListCityAirports(c *gin.Context) {
	cityID := c.Param("cityId")
	if cityID == "" {
		writeError(c, http.StatusBadRequest, "cityId is required")
		return
	}
	ctx := ctxlang.With(c.Request.Context(), ctxlang.ParseAcceptLanguage(c.GetHeader("Accept-Language")))
	airports, err := h.repo.ListAirportsByCity(ctx, cityID)
	if err == model.ErrCityNotFound {
		writeError(c, http.StatusNotFound, "city not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, airports)
}
