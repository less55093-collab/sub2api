package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type playgroundAPIKeyListerStub struct {
	keys    []service.APIKey
	userID  int64
	filters service.APIKeyListFilters
}

func (s *playgroundAPIKeyListerStub) List(ctx context.Context, userID int64, params pagination.PaginationParams, filters service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error) {
	s.userID = userID
	s.filters = filters
	return s.keys, &pagination.PaginationResult{Total: int64(len(s.keys)), Page: params.Page, PageSize: params.PageSize}, nil
}

func TestPlaygroundAPIKeyInjector_SelectsUserKeyAndRequestedGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(20)
	lister := &playgroundAPIKeyListerStub{
		keys: []service.APIKey{
			{Key: "disabled", Status: service.StatusAPIKeyDisabled, GroupID: &groupID},
			{ID: 2, Key: "playground-key", Status: service.StatusAPIKeyActive, GroupIDs: []int64{10, groupID}},
		},
	}

	router := gin.New()
	router.POST("/pg/chat/completions", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
		c.Next()
	}, playgroundAPIKeyInjector(lister), func(c *gin.Context) {
		require.Equal(t, "Bearer playground-key", c.Request.Header.Get("Authorization"))
		require.Equal(t, "20", c.Request.Header.Get(servermiddleware.HeaderAPIKeyGroupID))
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/pg/chat/completions", strings.NewReader(`{"model":"gpt-5","group_id":20,"messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(playgroundSelectedAPIKeyIDHeader, "2")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, int64(42), lister.userID)
	require.Equal(t, service.StatusAPIKeyActive, lister.filters.Status)
}

func TestPlaygroundAPIKeyInjector_HonorsSelectedUserKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(20)
	lister := &playgroundAPIKeyListerStub{
		keys: []service.APIKey{
			{ID: 1, Key: "first-key", Status: service.StatusAPIKeyActive, GroupIDs: []int64{groupID}},
			{ID: 2, Key: "selected-key", Status: service.StatusAPIKeyActive, GroupIDs: []int64{groupID}},
		},
	}

	router := gin.New()
	router.POST("/pg/chat/completions", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
		c.Next()
	}, playgroundAPIKeyInjector(lister), func(c *gin.Context) {
		require.Equal(t, "Bearer selected-key", c.Request.Header.Get("Authorization"))
		require.Equal(t, "20", c.Request.Header.Get(servermiddleware.HeaderAPIKeyGroupID))
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/pg/chat/completions", strings.NewReader(`{"model":"gpt-5","group_id":20,"messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(playgroundSelectedAPIKeyIDHeader, "2")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, int64(42), lister.userID)
	require.Equal(t, service.StatusAPIKeyActive, lister.filters.Status)
}

func TestPlaygroundAPIKeyInjector_RejectsMalformedSelectedKeyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lister := &playgroundAPIKeyListerStub{}

	router := gin.New()
	router.POST("/pg/chat/completions", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
		c.Next()
	}, playgroundAPIKeyInjector(lister))

	req := httptest.NewRequest(http.MethodPost, "/pg/chat/completions", strings.NewReader(`{"model":"gpt-5","messages":[]}`))
	req.Header.Set(playgroundSelectedAPIKeyIDHeader, "abc")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "INVALID_API_KEY")
	require.Zero(t, lister.userID)
}

func TestPlaygroundAPIKeyInjector_RejectsMissingSelectedKeyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lister := &playgroundAPIKeyListerStub{}

	router := gin.New()
	router.POST("/pg/chat/completions", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
		c.Next()
	}, playgroundAPIKeyInjector(lister))

	req := httptest.NewRequest(http.MethodPost, "/pg/chat/completions", strings.NewReader(`{"model":"gpt-5","messages":[]}`))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "PLAYGROUND_API_KEY_REQUIRED")
	require.Zero(t, lister.userID)
}

func TestPlaygroundAPIKeyInjector_RejectsUnavailableSelectedKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(20)
	lister := &playgroundAPIKeyListerStub{
		keys: []service.APIKey{
			{ID: 1, Key: "first-key", Status: service.StatusAPIKeyActive, GroupIDs: []int64{groupID}},
			{ID: 2, Key: "disabled-key", Status: service.StatusAPIKeyDisabled, GroupIDs: []int64{groupID}},
		},
	}

	router := gin.New()
	router.POST("/pg/chat/completions", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
		c.Next()
	}, playgroundAPIKeyInjector(lister))

	req := httptest.NewRequest(http.MethodPost, "/pg/chat/completions", strings.NewReader(`{"model":"gpt-5","group_id":20,"messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(playgroundSelectedAPIKeyIDHeader, "2")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "PLAYGROUND_API_KEY_UNAVAILABLE")
}

func TestPlaygroundAPIKeyInjector_RejectsSelectedKeyWithoutRequestedGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lister := &playgroundAPIKeyListerStub{
		keys: []service.APIKey{
			{ID: 2, Key: "selected-key", Status: service.StatusAPIKeyActive, GroupIDs: []int64{10}},
		},
	}

	router := gin.New()
	router.POST("/pg/chat/completions", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
		c.Next()
	}, playgroundAPIKeyInjector(lister))

	req := httptest.NewRequest(http.MethodPost, "/pg/chat/completions", strings.NewReader(`{"model":"gpt-5","group_id":20,"messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(playgroundSelectedAPIKeyIDHeader, "2")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "PLAYGROUND_API_KEY_UNAVAILABLE")
}

func TestPlaygroundAPIKeyInjector_ReturnsForbiddenWhenNoUsableKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lister := &playgroundAPIKeyListerStub{
		keys: []service.APIKey{{ID: 1, Key: "disabled", Status: service.StatusAPIKeyDisabled}},
	}

	router := gin.New()
	router.POST("/pg/chat/completions", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
		c.Next()
	}, playgroundAPIKeyInjector(lister))

	req := httptest.NewRequest(http.MethodPost, "/pg/chat/completions", strings.NewReader(`{"model":"gpt-5","messages":[]}`))
	req.Header.Set(playgroundSelectedAPIKeyIDHeader, "1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "PLAYGROUND_API_KEY_UNAVAILABLE")
}
