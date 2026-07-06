package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestResolveAPIKeyGroupFromHeader_BindsCandidateGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/pg/chat/completions", nil)
	c.Request.Header.Set(HeaderAPIKeyGroupID, "20")

	primaryID := int64(10)
	apiKey := &service.APIKey{
		Key:      "sk-test",
		GroupID:  &primaryID,
		GroupIDs: []int64{primaryID, 20},
		Group:    &service.Group{ID: primaryID, Name: "Default", Platform: service.PlatformAnthropic, Status: service.StatusActive},
		Groups: []service.Group{
			{ID: primaryID, Name: "Default", Platform: service.PlatformAnthropic, Status: service.StatusActive},
			{ID: 20, Name: "OpenAI", Platform: service.PlatformOpenAI, Status: service.StatusActive},
		},
	}

	resolved, ok := resolveAPIKeyGroupFromHeader(c, apiKey)
	require.True(t, ok)
	require.NotNil(t, resolved.GroupID)
	require.Equal(t, int64(20), *resolved.GroupID)
	require.NotNil(t, resolved.Group)
	require.Equal(t, service.PlatformOpenAI, resolved.Group.Platform)
}

func TestResolveAPIKeyGroupFromHeader_AcceptsLegacyHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/pg/chat/completions", nil)
	c.Request.Header.Set(LegacyHeaderAPIKeyGroupID, "20")

	primaryID := int64(10)
	apiKey := &service.APIKey{
		Key:      "sk-test",
		GroupID:  &primaryID,
		GroupIDs: []int64{primaryID, 20},
		Group:    &service.Group{ID: primaryID, Name: "Default", Platform: service.PlatformAnthropic, Status: service.StatusActive},
		Groups: []service.Group{
			{ID: primaryID, Name: "Default", Platform: service.PlatformAnthropic, Status: service.StatusActive},
			{ID: 20, Name: "OpenAI", Platform: service.PlatformOpenAI, Status: service.StatusActive},
		},
	}

	resolved, ok := resolveAPIKeyGroupFromHeader(c, apiKey)
	require.True(t, ok)
	require.NotNil(t, resolved.GroupID)
	require.Equal(t, int64(20), *resolved.GroupID)
	require.NotNil(t, resolved.Group)
	require.Equal(t, service.PlatformOpenAI, resolved.Group.Platform)
}
