package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

// ActiveGroupReader is the small group view required by the public model market.
type ActiveGroupReader interface {
	ListActive(ctx context.Context) ([]Group, error)
}

type ModelMarketConfig struct {
	Enabled         bool               `json:"enabled"`
	VisibleGroupIDs []int64            `json:"visible_group_ids"`
	Models          []ModelMarketModel `json:"models"`
}

type ModelMarketModel struct {
	ID               string             `json:"id"`
	DisplayName      string             `json:"display_name"`
	Provider         string             `json:"provider"`
	Platform         string             `json:"platform"`
	Category         string             `json:"category"`
	BillingMode      string             `json:"billing_mode"`
	Description      string             `json:"description"`
	EndpointTypes    []string           `json:"endpoint_types"`
	InputModalities  []string           `json:"input_modalities"`
	OutputModalities []string           `json:"output_modalities"`
	Tags             []string           `json:"tags"`
	Prices           []ModelMarketPrice `json:"prices"`
	DocsURL          string             `json:"docs_url"`
	ContextWindow    string             `json:"context_window"`
	MaxOutput        string             `json:"max_output"`
	Enabled          bool               `json:"enabled"`
	SortOrder        int                `json:"sort_order"`
}

type ModelMarketPrice struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Unit      string  `json:"unit"`
	BasePrice float64 `json:"base_price"`
}

type ModelMarketPublicGroup struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Platform         string  `json:"platform"`
	RateMultiplier   float64 `json:"rate_multiplier"`
	SubscriptionType string  `json:"subscription_type"`
	SortOrder        int     `json:"sort_order"`
}

type ModelMarketPublicResponse struct {
	Enabled bool                     `json:"enabled"`
	Groups  []ModelMarketPublicGroup `json:"groups"`
	Models  []ModelMarketModel       `json:"models"`
}

type ModelMarketAdminResponse struct {
	Config ModelMarketConfig        `json:"config"`
	Groups []ModelMarketPublicGroup `json:"groups"`
}

func (s *SettingService) SetModelMarketGroupReader(reader ActiveGroupReader) {
	s.modelMarketGroupReader = reader
}

func (s *SettingService) GetModelMarketConfig(ctx context.Context) (ModelMarketConfig, error) {
	cfg := defaultModelMarketConfig()
	if s == nil || s.settingRepo == nil {
		return cfg, nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyModelMarketConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return cfg, nil
		}
		return ModelMarketConfig{}, fmt.Errorf("get model market config: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return ModelMarketConfig{}, fmt.Errorf("parse model market config: %w", err)
	}
	return normalizeModelMarketConfig(cfg), nil
}

func (s *SettingService) UpdateModelMarketConfig(ctx context.Context, cfg ModelMarketConfig) (ModelMarketConfig, error) {
	if s == nil || s.settingRepo == nil {
		return ModelMarketConfig{}, fmt.Errorf("setting repository is not configured")
	}
	normalized := normalizeModelMarketConfig(cfg)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return ModelMarketConfig{}, fmt.Errorf("marshal model market config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyModelMarketConfig, string(payload)); err != nil {
		return ModelMarketConfig{}, fmt.Errorf("save model market config: %w", err)
	}
	return normalized, nil
}

func (s *SettingService) GetPublicModelMarket(ctx context.Context) (ModelMarketPublicResponse, error) {
	cfg, err := s.GetModelMarketConfig(ctx)
	if err != nil {
		return ModelMarketPublicResponse{}, err
	}
	groups, err := s.modelMarketPublicGroups(ctx, cfg.VisibleGroupIDs)
	if err != nil {
		return ModelMarketPublicResponse{}, err
	}
	return ModelMarketPublicResponse{
		Enabled: cfg.Enabled,
		Groups:  groups,
		Models:  publicModelMarketModels(cfg.Models),
	}, nil
}

func (s *SettingService) GetAdminModelMarket(ctx context.Context) (ModelMarketAdminResponse, error) {
	cfg, err := s.GetModelMarketConfig(ctx)
	if err != nil {
		return ModelMarketAdminResponse{}, err
	}
	groups, err := s.modelMarketAllGroups(ctx)
	if err != nil {
		return ModelMarketAdminResponse{}, err
	}
	return ModelMarketAdminResponse{Config: cfg, Groups: groups}, nil
}

func (s *SettingService) modelMarketAllGroups(ctx context.Context) ([]ModelMarketPublicGroup, error) {
	if s == nil || s.modelMarketGroupReader == nil {
		return []ModelMarketPublicGroup{defaultModelMarketGroup()}, nil
	}
	groups, err := s.modelMarketGroupReader.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	out := make([]ModelMarketPublicGroup, 0, len(groups)+1)
	out = append(out, defaultModelMarketGroup())
	for _, g := range groups {
		out = append(out, toModelMarketGroup(g))
	}
	sortModelMarketGroups(out)
	return out, nil
}

func (s *SettingService) modelMarketPublicGroups(ctx context.Context, allowlist []int64) ([]ModelMarketPublicGroup, error) {
	if s == nil || s.modelMarketGroupReader == nil {
		return []ModelMarketPublicGroup{defaultModelMarketGroup()}, nil
	}
	groups, err := s.modelMarketGroupReader.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	allowed := make(map[int64]struct{}, len(allowlist))
	for _, id := range allowlist {
		if id > 0 {
			allowed[id] = struct{}{}
		}
	}
	out := []ModelMarketPublicGroup{defaultModelMarketGroup()}
	for _, g := range groups {
		if g.IsExclusive || !g.IsActive() {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[g.ID]; !ok {
				continue
			}
		}
		out = append(out, toModelMarketGroup(g))
	}
	sortModelMarketGroups(out)
	return out, nil
}

func defaultModelMarketGroup() ModelMarketPublicGroup {
	return ModelMarketPublicGroup{
		ID:               0,
		Name:             "官方基准价",
		Platform:         "",
		RateMultiplier:   1,
		SubscriptionType: SubscriptionTypeStandard,
		SortOrder:        -1,
	}
}

func toModelMarketGroup(g Group) ModelMarketPublicGroup {
	multiplier := g.RateMultiplier
	if multiplier <= 0 || math.IsNaN(multiplier) || math.IsInf(multiplier, 0) {
		multiplier = 1
	}
	return ModelMarketPublicGroup{
		ID:               g.ID,
		Name:             g.Name,
		Platform:         strings.TrimSpace(g.Platform),
		RateMultiplier:   multiplier,
		SubscriptionType: firstNonEmpty(g.SubscriptionType, SubscriptionTypeStandard),
		SortOrder:        g.SortOrder,
	}
}

func sortModelMarketGroups(groups []ModelMarketPublicGroup) {
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].ID == 0 {
			return true
		}
		if groups[j].ID == 0 {
			return false
		}
		if groups[i].SortOrder != groups[j].SortOrder {
			return groups[i].SortOrder < groups[j].SortOrder
		}
		return groups[i].ID < groups[j].ID
	})
}

func publicModelMarketModels(models []ModelMarketModel) []ModelMarketModel {
	out := make([]ModelMarketModel, 0, len(models))
	for _, model := range models {
		if model.Enabled {
			out = append(out, model)
		}
	}
	sortModelMarketModels(out)
	return out
}

func normalizeModelMarketConfig(cfg ModelMarketConfig) ModelMarketConfig {
	defaults := defaultModelMarketConfig()
	if cfg.Models == nil {
		cfg.Models = defaults.Models
	}
	cleanGroups := make([]int64, 0, len(cfg.VisibleGroupIDs))
	seenGroups := make(map[int64]struct{}, len(cfg.VisibleGroupIDs))
	for _, id := range cfg.VisibleGroupIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seenGroups[id]; ok {
			continue
		}
		seenGroups[id] = struct{}{}
		cleanGroups = append(cleanGroups, id)
	}
	cfg.VisibleGroupIDs = cleanGroups

	models := make([]ModelMarketModel, 0, len(cfg.Models))
	for i, model := range cfg.Models {
		model.ID = strings.TrimSpace(model.ID)
		if model.ID == "" {
			continue
		}
		model.DisplayName = firstNonEmpty(strings.TrimSpace(model.DisplayName), model.ID)
		model.Provider = normalizeModelMarketProvider(model.Provider, model.Platform)
		model.Platform = normalizeModelMarketPlatform(model.Platform, model.Provider)
		model.Category = firstNonEmpty(strings.TrimSpace(model.Category), "chat")
		model.BillingMode = firstNonEmpty(strings.TrimSpace(model.BillingMode), "按量计费")
		model.EndpointTypes = uniqueNonEmpty(model.EndpointTypes)
		model.InputModalities = uniqueNonEmpty(model.InputModalities)
		model.OutputModalities = uniqueNonEmpty(model.OutputModalities)
		model.Tags = uniqueNonEmpty(model.Tags)
		model.Prices = normalizeModelMarketPrices(model.Prices)
		if model.SortOrder == 0 {
			model.SortOrder = i + 1
		}
		models = append(models, model)
	}
	cfg.Models = ensureModelMarketRequiredDefaults(models)
	sortModelMarketModels(cfg.Models)
	return cfg
}

func ensureModelMarketRequiredDefaults(models []ModelMarketModel) []ModelMarketModel {
	hasGPTImage2 := false
	for _, model := range models {
		if strings.EqualFold(model.ID, "gpt-image-2") {
			hasGPTImage2 = true
		}
	}
	if hasGPTImage2 {
		return models
	}
	return append(models, defaultGPTImage2Model())
}

func normalizeModelMarketProvider(provider, platform string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "" {
		return provider
	}
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case PlatformOpenAI:
		return "openai"
	case PlatformAnthropic:
		return "anthropic"
	default:
		return strings.ToLower(strings.TrimSpace(platform))
	}
}

func normalizeModelMarketPlatform(platform, provider string) string {
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform != "" {
		return platform
	}
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		return PlatformOpenAI
	case "anthropic":
		return PlatformAnthropic
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}

func normalizeModelMarketPrices(prices []ModelMarketPrice) []ModelMarketPrice {
	out := make([]ModelMarketPrice, 0, len(prices))
	for _, price := range prices {
		price.Key = strings.TrimSpace(price.Key)
		if price.Key == "" {
			continue
		}
		price.Label = firstNonEmpty(strings.TrimSpace(price.Label), price.Key)
		price.Unit = firstNonEmpty(strings.TrimSpace(price.Unit), "/ 1M")
		if price.BasePrice < 0 || math.IsNaN(price.BasePrice) || math.IsInf(price.BasePrice, 0) {
			price.BasePrice = 0
		}
		out = append(out, price)
	}
	return out
}

func sortModelMarketModels(models []ModelMarketModel) {
	sort.SliceStable(models, func(i, j int) bool {
		if models[i].SortOrder != models[j].SortOrder {
			return models[i].SortOrder < models[j].SortOrder
		}
		return models[i].ID < models[j].ID
	})
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

func defaultModelMarketConfig() ModelMarketConfig {
	return ModelMarketConfig{
		Enabled: true,
		Models: []ModelMarketModel{
			{
				ID:               "gpt-5.5",
				DisplayName:      "gpt-5.5",
				Provider:         "openai",
				Platform:         PlatformOpenAI,
				Category:         "chat",
				BillingMode:      "按量计费",
				Description:      "OpenAI 最新旗舰推理模型，适合复杂专业任务、代码与长上下文工作。",
				EndpointTypes:    []string{"openai-response", "openai-chat"},
				InputModalities:  []string{"文本", "图片"},
				OutputModalities: []string{"文本"},
				Tags:             []string{"OpenAI", "对话", "按量"},
				ContextWindow:    "1.05M tokens",
				MaxOutput:        "128K tokens",
				DocsURL:          "https://developers.openai.com/api/docs/models/gpt-5.5",
				Enabled:          true,
				SortOrder:        10,
				Prices: []ModelMarketPrice{
					{Key: "input", Label: "输入", Unit: "/ 1M", BasePrice: 5},
					{Key: "output", Label: "输出", Unit: "/ 1M", BasePrice: 30},
					{Key: "cache_read", Label: "缓存", Unit: "/ 1M", BasePrice: 0.5},
				},
			},
			{
				ID:               "gpt-5.4",
				DisplayName:      "gpt-5.4",
				Provider:         "openai",
				Platform:         PlatformOpenAI,
				Category:         "chat",
				BillingMode:      "按量计费",
				Description:      "OpenAI 高性价比前沿模型，适合代码、专业写作和多模态输入。",
				EndpointTypes:    []string{"openai-response", "openai-chat"},
				InputModalities:  []string{"文本", "图片"},
				OutputModalities: []string{"文本"},
				Tags:             []string{"OpenAI", "对话", "按量"},
				ContextWindow:    "1.05M tokens",
				MaxOutput:        "128K tokens",
				DocsURL:          "https://developers.openai.com/api/docs/models/gpt-5.4",
				Enabled:          true,
				SortOrder:        20,
				Prices: []ModelMarketPrice{
					{Key: "input", Label: "输入", Unit: "/ 1M", BasePrice: 2.5},
					{Key: "output", Label: "输出", Unit: "/ 1M", BasePrice: 15},
					{Key: "cache_read", Label: "缓存", Unit: "/ 1M", BasePrice: 0.25},
				},
			},
			defaultGPTImage2Model(),
			defaultClaudeModel("claude-fable-5", "Claude Fable 5", "下一代长任务智能模型，适合最高能力需求和复杂代理工作。", 10, 12.5, 20, 1, 50, 30),
			defaultClaudeModel("claude-opus-4-8", "Claude Opus 4.8", "面向复杂代理编码和企业工作的高能力模型。", 5, 6.25, 10, 0.5, 25, 40),
			defaultClaudeModel("claude-sonnet-5", "Claude Sonnet 5", "速度与智能平衡的模型。默认使用 2026-08-31 前官方介绍价。", 2, 2.5, 4, 0.2, 10, 50),
			defaultClaudeModel("claude-haiku-4-5", "Claude Haiku 4.5", "速度最快、接近前沿能力的 Claude 模型。", 1, 1.25, 2, 0.1, 5, 60),
		},
	}
}

func defaultGPTImage2Model() ModelMarketModel {
	return ModelMarketModel{
		ID:               "gpt-image-2",
		DisplayName:      "gpt-image-2",
		Provider:         "openai",
		Platform:         PlatformOpenAI,
		Category:         "image",
		BillingMode:      "按量计费",
		Description:      "OpenAI 最新生图模型，支持文生图和图生图编辑，适合高质量图片生成工作流。",
		EndpointTypes:    []string{"openai-images-generations", "openai-images-edits"},
		InputModalities:  []string{"文本", "图片"},
		OutputModalities: []string{"图片"},
		Tags:             []string{"OpenAI", "生图", "图生图", "按量"},
		ContextWindow:    "-",
		MaxOutput:        "-",
		DocsURL:          "https://platform.openai.com/docs/api-reference/images",
		Enabled:          true,
		SortOrder:        25,
		Prices: []ModelMarketPrice{
			{Key: "input_text", Label: "文本输入", Unit: "/ 1M tokens", BasePrice: 5},
			{Key: "input_image", Label: "图片输入", Unit: "/ 1M image tokens", BasePrice: 8},
			{Key: "output_text", Label: "文本输出", Unit: "/ 1M tokens", BasePrice: 10},
			{Key: "output_image", Label: "图片输出", Unit: "/ 1M image tokens", BasePrice: 30},
			{Key: "cache_read_text", Label: "文本缓存命中", Unit: "/ 1M tokens", BasePrice: 1.25},
			{Key: "cache_read_image", Label: "图片缓存命中", Unit: "/ 1M image tokens", BasePrice: 2},
		},
	}
}

func defaultClaudeModel(id, name, description string, input, cache5m, cache1h, cacheRead, output float64, sortOrder int) ModelMarketModel {
	return ModelMarketModel{
		ID:               id,
		DisplayName:      name,
		Provider:         "anthropic",
		Platform:         PlatformAnthropic,
		Category:         "chat",
		BillingMode:      "按量计费",
		Description:      description,
		EndpointTypes:    []string{"anthropic-messages"},
		InputModalities:  []string{"文本", "图片"},
		OutputModalities: []string{"文本"},
		Tags:             []string{"Anthropic", "对话", "按量"},
		ContextWindow:    "1M tokens",
		MaxOutput:        "128K tokens",
		DocsURL:          "https://platform.claude.com/docs/en/about-claude/models/overview",
		Enabled:          true,
		SortOrder:        sortOrder,
		Prices: []ModelMarketPrice{
			{Key: "input", Label: "输入", Unit: "/ 1M", BasePrice: input},
			{Key: "output", Label: "输出", Unit: "/ 1M", BasePrice: output},
			{Key: "cache_5m", Label: "5m 缓存写入", Unit: "/ 1M", BasePrice: cache5m},
			{Key: "cache_1h", Label: "1h 缓存写入", Unit: "/ 1M", BasePrice: cache1h},
			{Key: "cache_read", Label: "缓存命中", Unit: "/ 1M", BasePrice: cacheRead},
		},
	}
}
