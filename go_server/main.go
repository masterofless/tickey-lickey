package main

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	shouter := echo.New()
	shouter.Use(middleware.Logger())
	shouter.Use(middleware.Recover())

	shouter.GET("/", healthcheck)
	shouter.GET("/redeem", redeem)

	shouter.Logger.Fatal(shouter.Start(":8080")) // Start server
}

func healthcheck(context echo.Context) error {
	return context.String(http.StatusOK, "I'm okay; how are you?")
}

type RedemptionResponse struct {
	ValidationStatus string
	URL              string `json:"url"`
}

func redeem(context echo.Context) error {
	answer := &RedemptionResponse{}
	json, _ := json.Marshal(answer)
	return context.String(http.StatusOK, string(json))
}
