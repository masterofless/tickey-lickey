package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	echoer := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	context := echoer.NewContext(req, rec)

	if assert.NoError(t, healthcheck(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)      // test status
		assert.Contains(t, rec.Body.String(), "okay") // test response body
	}
}
