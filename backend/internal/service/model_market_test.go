//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type modelMarketSettingRepoStub struct {
	value string
}

func (s *modelMarketSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *modelMarketSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if key != SettingKeyModelMarketConfig || s.value == "" {
		return "", ErrSettingNotFound
	}
	return s.value, nil
}

func (s *modelMarketSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *modelMarketSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *modelMarketSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *modelMarketSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *modelMarketSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_DefaultModelMarketConfigIncludesGPTImage2(t *testing.T) {
	svc := NewSettingService(nil, nil)

	cfg, err := svc.GetModelMarketConfig(context.Background())
	require.NoError(t, err)

	model := requireModelMarketModel(t, cfg.Models, "gpt-image-2")
	require.True(t, model.Enabled)
	require.Equal(t, "openai", model.Provider)
	require.Equal(t, PlatformOpenAI, model.Platform)
	require.Equal(t, "image", model.Category)
	require.ElementsMatch(t, []string{"openai-images-generations", "openai-images-edits"}, model.EndpointTypes)
	require.ElementsMatch(t, []string{"文本", "图片"}, model.InputModalities)
	require.ElementsMatch(t, []string{"图片"}, model.OutputModalities)

	prices := make(map[string]ModelMarketPrice, len(model.Prices))
	for _, price := range model.Prices {
		prices[price.Key] = price
	}
	require.Equal(t, 5.0, prices["input_text"].BasePrice)
	require.Equal(t, 8.0, prices["input_image"].BasePrice)
	require.Equal(t, 10.0, prices["output_text"].BasePrice)
	require.Equal(t, 30.0, prices["output_image"].BasePrice)
	require.Equal(t, 1.25, prices["cache_read_text"].BasePrice)
	require.Equal(t, 2.0, prices["cache_read_image"].BasePrice)
}

func TestSettingService_PublicModelMarketIncludesGPTImage2(t *testing.T) {
	svc := NewSettingService(nil, nil)

	market, err := svc.GetPublicModelMarket(context.Background())
	require.NoError(t, err)

	model := requireModelMarketModel(t, market.Models, "gpt-image-2")
	require.Equal(t, "image", model.Category)
	require.Contains(t, model.EndpointTypes, "openai-images-generations")
	require.Contains(t, model.EndpointTypes, "openai-images-edits")
}

func TestSettingService_ModelMarketConfigSeedsGPTImage2ForExistingChatOnlyConfig(t *testing.T) {
	raw, err := json.Marshal(ModelMarketConfig{
		Enabled: true,
		Models: []ModelMarketModel{
			{
				ID:            "legacy-chat",
				DisplayName:   "legacy-chat",
				Provider:      "openai",
				Platform:      PlatformOpenAI,
				Category:      "chat",
				EndpointTypes: []string{"openai-chat"},
				Enabled:       true,
				SortOrder:     10,
			},
		},
	})
	require.NoError(t, err)
	svc := NewSettingService(&modelMarketSettingRepoStub{value: string(raw)}, nil)

	cfg, err := svc.GetModelMarketConfig(context.Background())
	require.NoError(t, err)

	requireModelMarketModel(t, cfg.Models, "legacy-chat")
	model := requireModelMarketModel(t, cfg.Models, "gpt-image-2")
	require.Equal(t, "image", model.Category)
	require.Contains(t, model.EndpointTypes, "openai-images-generations")
}

func TestSettingService_ModelMarketConfigSeedsGPTImage2AlongsideExistingImageModel(t *testing.T) {
	raw, err := json.Marshal(ModelMarketConfig{
		Enabled: true,
		Models: []ModelMarketModel{
			{
				ID:            "custom-image",
				DisplayName:   "custom-image",
				Provider:      "openai",
				Platform:      PlatformOpenAI,
				Category:      "image",
				EndpointTypes: []string{"openai-images-generations"},
				Enabled:       true,
				SortOrder:     10,
			},
		},
	})
	require.NoError(t, err)
	svc := NewSettingService(&modelMarketSettingRepoStub{value: string(raw)}, nil)

	cfg, err := svc.GetModelMarketConfig(context.Background())
	require.NoError(t, err)

	requireModelMarketModel(t, cfg.Models, "custom-image")
	requireModelMarketModel(t, cfg.Models, "gpt-image-2")
}

func TestSettingService_ModelMarketConfigRespectsExistingGPTImage2(t *testing.T) {
	raw, err := json.Marshal(ModelMarketConfig{
		Enabled: true,
		Models: []ModelMarketModel{
			{
				ID:            "gpt-image-2",
				DisplayName:   "Hidden image model",
				Provider:      "openai",
				Platform:      PlatformOpenAI,
				Category:      "image",
				EndpointTypes: []string{"openai-images-generations"},
				Enabled:       false,
				SortOrder:     99,
			},
		},
	})
	require.NoError(t, err)
	svc := NewSettingService(&modelMarketSettingRepoStub{value: string(raw)}, nil)

	cfg, err := svc.GetModelMarketConfig(context.Background())
	require.NoError(t, err)

	model := requireModelMarketModel(t, cfg.Models, "gpt-image-2")
	require.Equal(t, "Hidden image model", model.DisplayName)
	require.False(t, model.Enabled)
}

func requireModelMarketModel(t *testing.T, models []ModelMarketModel, id string) ModelMarketModel {
	t.Helper()
	for _, model := range models {
		if model.ID == id {
			return model
		}
	}
	require.Failf(t, "model not found", "model %q not found", id)
	return ModelMarketModel{}
}
