package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Arrange
	apiToken := "test-secret-token"

	router := gin.New()
	router.Use(AuthMiddleware(apiToken))
	router.GET("/api/v1/messages", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	req.Header.Set("Authorization", "Bearer test-secret-token")

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	// Arrange
	apiToken := "test-secret-token"
	middleware := AuthMiddleware(apiToken)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)

	// Act
	middleware(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Arrange
	apiToken := "test-secret-token"
	middleware := AuthMiddleware(apiToken)

	testCases := []struct {
		name          string
		authHeader    string
		expectedError string
	}{
		{
			name:          "missing Bearer prefix",
			authHeader:    "test-secret-token",
			expectedError: "invalid authorization format",
		},
		{
			name:          "wrong prefix",
			authHeader:    "Basic test-secret-token",
			expectedError: "invalid authorization format",
		},
		{
			name:          "empty token",
			authHeader:    "Bearer ",
			expectedError: "invalid token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
			c.Request.Header.Set("Authorization", tc.authHeader)

			// Act
			middleware(c)

			// Assert
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedError)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Arrange
	apiToken := "test-secret-token"
	middleware := AuthMiddleware(apiToken)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	c.Request.Header.Set("Authorization", "Bearer wrong-token")

	// Act
	middleware(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

func TestAuthMiddleware_SkipHealthEndpoints(t *testing.T) {
	// Arrange
	apiToken := "test-secret-token"

	testCases := []struct {
		name string
		path string
	}{
		{name: "health endpoint", path: "/health"},
		{name: "ready endpoint", path: "/ready"},
		{name: "live endpoint", path: "/live"},
		{name: "swagger endpoint", path: "/swagger/index.html"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware(apiToken))
			router.GET(tc.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			// No auth header - should still pass for public endpoints

			// Act
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "ok")
		})
	}
}

func TestAuthMiddleware_RequireAuthForProtectedEndpoints(t *testing.T) {
	// Arrange
	apiToken := "test-secret-token"
	middleware := AuthMiddleware(apiToken)

	testCases := []string{
		"/api/v1/messages",
		"/api/v1/messages/123",
		"/api/v1/scheduler/start",
		"/api/v1/scheduler/stop",
	}

	for _, path := range testCases {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, path, nil)

			// Act - no auth header
			middleware(c)

			// Assert
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}
