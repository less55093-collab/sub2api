package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ModelMarketHandler struct {
	settingService *service.SettingService
}

func NewModelMarketHandler(settingService *service.SettingService) *ModelMarketHandler {
	return &ModelMarketHandler{settingService: settingService}
}

// Public returns the public, channel-independent model market.
// GET /api/v1/model-market
func (h *ModelMarketHandler) Public(c *gin.Context) {
	data, err := h.settingService.GetPublicModelMarket(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

// AdminConfig returns the full editable config plus active groups.
// GET /api/v1/admin/model-market/config
func (h *ModelMarketHandler) AdminConfig(c *gin.Context) {
	data, err := h.settingService.GetAdminModelMarket(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

// UpdateAdminConfig replaces the model market config.
// PUT /api/v1/admin/model-market/config
func (h *ModelMarketHandler) UpdateAdminConfig(c *gin.Context) {
	var cfg service.ModelMarketConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	saved, err := h.settingService.UpdateModelMarketConfig(c.Request.Context(), cfg)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, saved)
}
