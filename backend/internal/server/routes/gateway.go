package routes

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// RegisterGatewayRoutes 注册 API 网关路由（Claude/OpenAI/Gemini 兼容）
func RegisterGatewayRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	apiKeyAuth middleware.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
) {
	bodyLimit := middleware.RequestBodyLimit(cfg.Gateway.MaxBodySize)
	clientRequestID := middleware.ClientRequestID()
	opsErrorLogger := handler.OpsErrorLoggerMiddleware(opsService)
	endpointNorm := handler.InboundEndpointMiddleware()

	// 未分组 Key 拦截中间件（按协议格式区分错误响应）
	requireGroupAnthropic := middleware.RequireGroupAssignment(settingService, middleware.AnthropicErrorWriter)
	requireGroupGoogle := middleware.RequireGroupAssignment(settingService, middleware.GoogleErrorWriter)
	playgroundKeyAuth := playgroundAPIKeyInjector(apiKeyService)

	dispatchMessages := func(c *gin.Context) {
		platform := routePlatformForMessagesEndpoint(c)
		middleware.SetRoutePlatformIntent(c, platform)
		if isOpenAICompatibleRoutePlatform(platform) {
			h.OpenAIGateway.Messages(c)
			return
		}
		h.Gateway.Messages(c)
	}
	dispatchOpenAICompatible := func(openAIHandler gin.HandlerFunc, anthropicHandler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			platform := routePlatformForOpenAICompatibleEndpoint(c)
			middleware.SetRoutePlatformIntent(c, platform)
			if isOpenAICompatibleRoutePlatform(platform) {
				openAIHandler(c)
				return
			}
			anthropicHandler(c)
		}
	}
	countTokensHandler := func(c *gin.Context) {
		platform := routePlatformForCountTokensEndpoint(c)
		middleware.SetRoutePlatformIntent(c, platform)
		if platform == service.PlatformOpenAI {
			h.OpenAIGateway.CountTokens(c)
			return
		}
		if platform == service.PlatformGrok {
			service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
			c.JSON(http.StatusNotFound, gin.H{
				"type": "error",
				"error": gin.H{
					"type":    "not_found_error",
					"message": "Token counting is not supported for this platform",
				},
			})
			return
		}
		h.Gateway.CountTokens(c)
	}
	imagesHandler := func(c *gin.Context) {
		platform := routePlatformForOpenAIMediaEndpoint(c)
		middleware.SetRoutePlatformIntent(c, platform)
		if platform == service.PlatformGrok {
			h.OpenAIGateway.GrokImages(c)
			return
		}
		if platform == service.PlatformOpenAI {
			h.OpenAIGateway.Images(c)
			return
		}
		service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"type":    "not_found_error",
				"message": "Images API is not supported for this platform",
			},
		})
	}
	videoGenerationHandler := func(c *gin.Context) {
		if hasAPIKeyPlatform(c, service.PlatformGrok) {
			middleware.SetRoutePlatformIntent(c, service.PlatformGrok)
			h.OpenAIGateway.GrokVideoGeneration(c)
			return
		}
		service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"type":    "not_found_error",
				"message": "Videos API is not supported for this platform",
			},
		})
	}
	videoStatusHandler := func(c *gin.Context) {
		if hasAPIKeyPlatform(c, service.PlatformGrok) {
			middleware.SetRoutePlatformIntent(c, service.PlatformGrok)
			h.OpenAIGateway.GrokVideoStatus(c)
			return
		}
		service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"type":    "not_found_error",
				"message": "Videos API is not supported for this platform",
			},
		})
	}
	// API网关（Claude API兼容）
	gateway := r.Group("/v1")
	gateway.Use(bodyLimit)
	gateway.Use(clientRequestID)
	gateway.Use(opsErrorLogger)
	gateway.Use(endpointNorm)
	gateway.Use(gin.HandlerFunc(apiKeyAuth))
	gateway.Use(requireGroupAnthropic)
	{
		// /v1/messages: auto-route based on group platform
		gateway.POST("/messages", dispatchMessages)
		// /v1/messages/count_tokens: OpenAI uses Anthropic-compat bridge; other
		// OpenAI-compatible platforms keep the prior unsupported response.
		gateway.POST("/messages/count_tokens", countTokensHandler)
		gateway.GET("/models", h.Gateway.Models)
		gateway.GET("/usage", h.Gateway.Usage)
		// OpenAI Responses API: auto-route based on group platform
		gateway.POST("/responses", dispatchOpenAICompatible(h.OpenAIGateway.Responses, h.Gateway.Responses))
		gateway.POST("/responses/*subpath", dispatchOpenAICompatible(h.OpenAIGateway.Responses, h.Gateway.Responses))
		gateway.GET("/responses", func(c *gin.Context) {
			middleware.SetRoutePlatformIntent(c, routePlatformForOpenAICompatibleEndpoint(c))
			h.OpenAIGateway.ResponsesWebSocket(c)
		})
		// OpenAI Chat Completions API: auto-route based on group platform
		gateway.POST("/chat/completions", dispatchOpenAICompatible(h.OpenAIGateway.ChatCompletions, h.Gateway.ChatCompletions))
		gateway.POST("/embeddings", func(c *gin.Context) {
			if !hasAPIKeyPlatform(c, service.PlatformOpenAI) {
				service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Embeddings API is not supported for this platform",
					},
				})
				return
			}
			h.OpenAIGateway.Embeddings(c)
		})
		gateway.POST("/images/generations", imagesHandler)
		gateway.POST("/images/edits", imagesHandler)
		gateway.POST("/videos/generations", videoGenerationHandler)
		gateway.GET("/videos/:request_id", videoStatusHandler)
	}

	// Gemini 原生 API 兼容层（Gemini SDK/CLI 直连）
	gemini := r.Group("/v1beta")
	gemini.Use(bodyLimit)
	gemini.Use(clientRequestID)
	gemini.Use(opsErrorLogger)
	gemini.Use(endpointNorm)
	gemini.Use(middleware.APIKeyAuthWithSubscriptionGoogle(apiKeyService, subscriptionService, cfg))
	gemini.Use(requireGroupGoogle)
	{
		gemini.GET("/models", h.Gateway.GeminiV1BetaListModels)
		gemini.GET("/models/:model", h.Gateway.GeminiV1BetaGetModel)
		// Gin treats ":" as a param marker, but Gemini uses "{model}:{action}" in the same segment.
		gemini.POST("/models/*modelAction", h.Gateway.GeminiV1BetaModels)
	}

	// OpenAI Responses API（不带v1前缀的别名）— auto-route based on group platform
	responsesHandler := dispatchOpenAICompatible(h.OpenAIGateway.Responses, h.Gateway.Responses)
	r.POST("/responses", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, responsesHandler)
	r.POST("/responses/*subpath", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, responsesHandler)
	r.GET("/responses", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, func(c *gin.Context) {
		middleware.SetRoutePlatformIntent(c, routePlatformForOpenAICompatibleEndpoint(c))
		h.OpenAIGateway.ResponsesWebSocket(c)
	})
	codexDirect := r.Group("/backend-api/codex")
	codexDirect.Use(bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic)
	{
		codexDirect.POST("/responses", responsesHandler)
		codexDirect.POST("/responses/*subpath", responsesHandler)
		codexDirect.GET("/responses", func(c *gin.Context) {
			middleware.SetRoutePlatformIntent(c, routePlatformForOpenAICompatibleEndpoint(c))
			h.OpenAIGateway.ResponsesWebSocket(c)
		})
	}
	// OpenAI Chat Completions API（不带v1前缀的别名）— auto-route based on group platform
	r.POST("/chat/completions", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, dispatchOpenAICompatible(h.OpenAIGateway.ChatCompletions, h.Gateway.ChatCompletions))
	r.POST("/embeddings", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, func(c *gin.Context) {
		if !hasAPIKeyPlatform(c, service.PlatformOpenAI) {
			service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"type":    "not_found_error",
					"message": "Embeddings API is not supported for this platform",
				},
			})
			return
		}
		h.OpenAIGateway.Embeddings(c)
	})
	r.POST("/images/generations", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, imagesHandler)
	r.POST("/images/edits", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, imagesHandler)
	r.POST("/videos/generations", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, videoGenerationHandler)
	r.GET("/videos/:request_id", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, videoStatusHandler)

	playground := r.Group("/pg")
	playground.Use(bodyLimit)
	playground.Use(clientRequestID)
	playground.Use(opsErrorLogger)
	playground.Use(endpointNorm)
	playground.Use(gin.HandlerFunc(jwtAuth))
	playground.Use(middleware.BackendModeUserGuard(settingService))
	playground.Use(playgroundKeyAuth)
	playground.Use(gin.HandlerFunc(apiKeyAuth))
	playground.Use(requireGroupAnthropic)
	{
		playground.POST("/chat/completions", dispatchOpenAICompatible(h.OpenAIGateway.ChatCompletions, h.Gateway.ChatCompletions))
	}

	playgroundAPI := r.Group("/api/v1/playground")
	playgroundAPI.Use(clientRequestID)
	playgroundAPI.Use(opsErrorLogger)
	playgroundAPI.Use(endpointNorm)
	playgroundAPI.Use(gin.HandlerFunc(jwtAuth))
	playgroundAPI.Use(middleware.BackendModeUserGuard(settingService))
	playgroundAPI.Use(playgroundKeyAuth)
	playgroundAPI.Use(gin.HandlerFunc(apiKeyAuth))
	playgroundAPI.Use(requireGroupAnthropic)
	{
		playgroundAPI.GET("/models", h.Gateway.Models)
	}

	// Antigravity 模型列表
	r.GET("/antigravity/models", gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, h.Gateway.AntigravityModels)

	// Antigravity 专用路由（仅使用 antigravity 账户，不混合调度）
	antigravityV1 := r.Group("/antigravity/v1")
	antigravityV1.Use(bodyLimit)
	antigravityV1.Use(clientRequestID)
	antigravityV1.Use(opsErrorLogger)
	antigravityV1.Use(endpointNorm)
	antigravityV1.Use(middleware.ForcePlatform(service.PlatformAntigravity))
	antigravityV1.Use(gin.HandlerFunc(apiKeyAuth))
	antigravityV1.Use(requireGroupAnthropic)
	{
		antigravityV1.POST("/messages", h.Gateway.Messages)
		antigravityV1.POST("/messages/count_tokens", h.Gateway.CountTokens)
		antigravityV1.GET("/models", h.Gateway.AntigravityModels)
		antigravityV1.GET("/usage", h.Gateway.Usage)
	}

	antigravityV1Beta := r.Group("/antigravity/v1beta")
	antigravityV1Beta.Use(bodyLimit)
	antigravityV1Beta.Use(clientRequestID)
	antigravityV1Beta.Use(opsErrorLogger)
	antigravityV1Beta.Use(endpointNorm)
	antigravityV1Beta.Use(middleware.ForcePlatform(service.PlatformAntigravity))
	antigravityV1Beta.Use(middleware.APIKeyAuthWithSubscriptionGoogle(apiKeyService, subscriptionService, cfg))
	antigravityV1Beta.Use(requireGroupGoogle)
	{
		antigravityV1Beta.GET("/models", h.Gateway.GeminiV1BetaListModels)
		antigravityV1Beta.GET("/models/:model", h.Gateway.GeminiV1BetaGetModel)
		antigravityV1Beta.POST("/models/*modelAction", h.Gateway.GeminiV1BetaModels)
	}

}

type playgroundAPIKeyLister interface {
	List(ctx context.Context, userID int64, params pagination.PaginationParams, filters service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error)
}

type playgroundGroupRequest struct {
	id   *int64
	name string
}

const playgroundSelectedAPIKeyIDHeader = "X-FluxRouter-API-Key-ID"

func playgroundAPIKeyInjector(lister playgroundAPIKeyLister) gin.HandlerFunc {
	return func(c *gin.Context) {
		if lister == nil {
			middleware.AbortWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Playground API key service is not configured")
			return
		}
		subject, ok := middleware.GetAuthSubjectFromContext(c)
		if !ok {
			middleware.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
			return
		}
		groupReq, ok := playgroundGroupRequestFromContext(c)
		if !ok {
			middleware.AbortWithError(c, http.StatusBadRequest, "INVALID_GROUP", "Invalid group_id")
			return
		}
		selectedKeyID, ok := playgroundSelectedAPIKeyIDFromContext(c)
		if !ok {
			middleware.AbortWithError(c, http.StatusBadRequest, "INVALID_API_KEY", "Invalid API key selection")
			return
		}
		if selectedKeyID == nil {
			middleware.AbortWithError(c, http.StatusForbidden, "PLAYGROUND_API_KEY_REQUIRED", "Select an API key for playground")
			return
		}

		keys, _, err := lister.List(c.Request.Context(), subject.UserID, pagination.PaginationParams{
			Page:      1,
			PageSize:  1000,
			SortBy:    "created_at",
			SortOrder: pagination.SortOrderAsc,
		}, service.APIKeyListFilters{Status: service.StatusAPIKeyActive})
		if err != nil {
			middleware.AbortWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list user API keys")
			return
		}

		key, groupID := selectPlaygroundAPIKey(keys, groupReq, selectedKeyID)
		if key == nil {
			middleware.AbortWithError(c, http.StatusForbidden, "PLAYGROUND_API_KEY_UNAVAILABLE", "Selected API key is not available for playground")
			return
		}
		c.Request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(key.Key))
		if groupID != nil {
			c.Request.Header.Set(middleware.HeaderAPIKeyGroupID, strconv.FormatInt(*groupID, 10))
		}
		c.Next()
	}
}

func selectPlaygroundAPIKey(keys []service.APIKey, groupReq playgroundGroupRequest, selectedKeyID *int64) (*service.APIKey, *int64) {
	for i := range keys {
		if selectedKeyID != nil && keys[i].ID != *selectedKeyID {
			continue
		}
		if !playgroundAPIKeyUsable(keys[i]) {
			continue
		}
		if groupReq.id == nil && groupReq.name == "" {
			return &keys[i], nil
		}
		if groupID, ok := playgroundAPIKeyGroupID(keys[i], groupReq); ok {
			return &keys[i], &groupID
		}
	}
	return nil, nil
}

func playgroundSelectedAPIKeyIDFromContext(c *gin.Context) (*int64, bool) {
	raw := strings.TrimSpace(c.GetHeader(playgroundSelectedAPIKeyIDHeader))
	if raw == "" {
		return nil, true
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return nil, false
	}
	return &id, true
}

func playgroundAPIKeyUsable(key service.APIKey) bool {
	if strings.TrimSpace(key.Key) == "" {
		return false
	}
	return key.Status == "" || key.IsActive()
}

func playgroundAPIKeyGroupID(key service.APIKey, groupReq playgroundGroupRequest) (int64, bool) {
	if groupReq.id != nil {
		groupID := *groupReq.id
		if key.GroupID != nil && *key.GroupID == groupID {
			return groupID, true
		}
		for _, id := range key.GroupIDs {
			if id == groupID {
				return groupID, true
			}
		}
		for _, group := range key.Groups {
			if group.ID == groupID {
				return groupID, true
			}
		}
		if key.Group != nil && key.Group.ID == groupID {
			return groupID, true
		}
		return 0, false
	}
	if groupReq.name == "" {
		return 0, false
	}
	if key.Group != nil && strings.EqualFold(strings.TrimSpace(key.Group.Name), groupReq.name) && key.Group.ID > 0 {
		return key.Group.ID, true
	}
	for _, group := range key.Groups {
		if strings.EqualFold(strings.TrimSpace(group.Name), groupReq.name) && group.ID > 0 {
			return group.ID, true
		}
	}
	return 0, false
}

func playgroundGroupRequestFromContext(c *gin.Context) (playgroundGroupRequest, bool) {
	if raw := strings.TrimSpace(c.Query("group_id")); raw != "" {
		return playgroundGroupRequestFromID(raw)
	}
	if raw := strings.TrimSpace(c.Query("group")); raw != "" {
		return playgroundGroupRequestFromFlexible(raw), true
	}
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return playgroundGroupRequest{}, true
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return playgroundGroupRequest{}, true
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return playgroundGroupRequest{}, true
	}
	if value := gjson.GetBytes(body, "group_id"); value.Exists() {
		return playgroundGroupRequestFromID(value.String())
	}
	if value := gjson.GetBytes(body, "group"); value.Exists() {
		return playgroundGroupRequestFromFlexible(value.String()), true
	}
	return playgroundGroupRequest{}, true
}

func playgroundGroupRequestFromID(raw string) (playgroundGroupRequest, bool) {
	groupID, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || groupID <= 0 {
		return playgroundGroupRequest{}, false
	}
	return playgroundGroupRequest{id: &groupID}, true
}

func playgroundGroupRequestFromFlexible(raw string) playgroundGroupRequest {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return playgroundGroupRequest{}
	}
	if groupID, err := strconv.ParseInt(raw, 10, 64); err == nil && groupID > 0 {
		return playgroundGroupRequest{id: &groupID}
	}
	return playgroundGroupRequest{name: raw}
}

// getGroupPlatform extracts the group platform from the API Key stored in context.
func getGroupPlatform(c *gin.Context) string {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey.Group == nil {
		return ""
	}
	return apiKey.Group.Platform
}

func hasAPIKeyPlatform(c *gin.Context, platform string) bool {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok {
		return false
	}
	return service.APIKeyHasCandidateGroup(apiKey, platform)
}

func routePlatformForMessagesEndpoint(c *gin.Context) string {
	model := normalizedRequestModel(c)
	if requestModelLooksGrok(model) && hasAPIKeyPlatform(c, service.PlatformGrok) {
		return service.PlatformGrok
	}
	if requestModelLooksOpenAI(model) && hasAPIKeyPlatform(c, service.PlatformOpenAI) {
		return service.PlatformOpenAI
	}
	if hasAPIKeyPlatform(c, service.PlatformAnthropic) {
		return service.PlatformAnthropic
	}
	if hasAPIKeyPlatform(c, service.PlatformOpenAI) {
		return service.PlatformOpenAI
	}
	if hasAPIKeyPlatform(c, service.PlatformGrok) {
		return service.PlatformGrok
	}
	return getGroupPlatform(c)
}

func routePlatformForOpenAICompatibleEndpoint(c *gin.Context) string {
	model := normalizedRequestModel(c)
	if requestModelLooksAnthropic(model) && hasAPIKeyPlatform(c, service.PlatformAnthropic) {
		return service.PlatformAnthropic
	}
	if requestModelLooksGrok(model) && hasAPIKeyPlatform(c, service.PlatformGrok) {
		return service.PlatformGrok
	}
	if requestModelLooksOpenAI(model) && hasAPIKeyPlatform(c, service.PlatformOpenAI) {
		return service.PlatformOpenAI
	}
	if hasAPIKeyPlatform(c, service.PlatformOpenAI) {
		return service.PlatformOpenAI
	}
	if hasAPIKeyPlatform(c, service.PlatformGrok) {
		return service.PlatformGrok
	}
	if hasAPIKeyPlatform(c, service.PlatformAnthropic) {
		return service.PlatformAnthropic
	}
	return getGroupPlatform(c)
}

func routePlatformForCountTokensEndpoint(c *gin.Context) string {
	model := normalizedRequestModel(c)
	if requestModelLooksAnthropic(model) && hasAPIKeyPlatform(c, service.PlatformAnthropic) {
		return service.PlatformAnthropic
	}
	if requestModelLooksOpenAI(model) && hasAPIKeyPlatform(c, service.PlatformOpenAI) {
		return service.PlatformOpenAI
	}
	if requestModelLooksGrok(model) && hasAPIKeyPlatform(c, service.PlatformGrok) {
		return service.PlatformGrok
	}
	if hasAPIKeyPlatform(c, service.PlatformOpenAI) {
		return service.PlatformOpenAI
	}
	if hasAPIKeyPlatform(c, service.PlatformAnthropic) {
		return service.PlatformAnthropic
	}
	if hasAPIKeyPlatform(c, service.PlatformGrok) {
		return service.PlatformGrok
	}
	return getGroupPlatform(c)
}

func routePlatformForOpenAIMediaEndpoint(c *gin.Context) string {
	model := normalizedRequestModel(c)
	if requestModelLooksGrok(model) && hasAPIKeyPlatform(c, service.PlatformGrok) {
		return service.PlatformGrok
	}
	if hasAPIKeyPlatform(c, service.PlatformOpenAI) {
		return service.PlatformOpenAI
	}
	if hasAPIKeyPlatform(c, service.PlatformGrok) {
		return service.PlatformGrok
	}
	return getGroupPlatform(c)
}

func isOpenAICompatibleRoutePlatform(platform string) bool {
	return platform == service.PlatformOpenAI || platform == service.PlatformGrok
}

func requestModelLooksOpenAICompatible(c *gin.Context) bool {
	model := normalizedRequestModel(c)
	return requestModelLooksOpenAI(model) || requestModelLooksGrok(model)
}

func normalizedRequestModel(c *gin.Context) string {
	return strings.ToLower(strings.TrimSpace(peekJSONRequestModel(c)))
}

func requestModelLooksOpenAI(model string) bool {
	return strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "chatgpt-") ||
		strings.HasPrefix(model, "codex") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4") ||
		strings.HasPrefix(model, "openai/")
}

func requestModelLooksGrok(model string) bool {
	return strings.HasPrefix(model, "grok") ||
		strings.HasPrefix(model, "xai/")
}

func requestModelLooksAnthropic(model string) bool {
	return strings.HasPrefix(model, "claude") ||
		strings.HasPrefix(model, "anthropic/claude") ||
		strings.HasPrefix(model, "sonnet") ||
		strings.HasPrefix(model, "opus") ||
		strings.HasPrefix(model, "haiku")
}

func peekJSONRequestModel(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return ""
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return ""
	}
	return gjson.GetBytes(body, "model").String()
}
