package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/masterofless/eslib"
	"github.com/masterofless/rmqlib"
)

func main() {
	server := echo.New()
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	server.GET("/", healthcheck)
	server.GET("/redeem/:token", redeem)

	server.Logger.Fatal(server.Start(":8080")) // Start server
}

func healthcheck(httpContext echo.Context) error {
	return httpContext.String(http.StatusOK, "I'm okay; how are you?")
}

type Response struct {
	Answer string
}

type WristBand struct {
	Token string
	URL   string
	Used  string
}

func redeem(httpContext echo.Context) error {
	wristband := readWristBand(httpContext) // read from elasticsearch
	enqueueRedemption(wristband)            // write to rabbitmq
	// respond
	json, _ := json.Marshal(wristband)
	return httpContext.String(http.StatusOK, string(json))
}

func getQuery(token string) *bytes.Buffer {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"token": token,
			},
		},
	}
	var queryBuf bytes.Buffer
	if err := json.NewEncoder(&queryBuf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	return &queryBuf
}

func readWristBand(httpContext echo.Context) *WristBand {
	token := httpContext.Param("token")
	results := eslib.GetSearchResult(getQuery(token))
	// for _, hit := range results["hits"].(map[string]interface{})["hits"].([]interface{}) {
	// 	log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	// }
	hits := int(results["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	if hits == 0 {
		return &WristBand{Token: token, URL: "Sorry, that token was not found"}
	}
	firstHit := results["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})
	if firstHit["used"].(string) == "false" {
		return &WristBand{Token: firstHit["token"].(string), Used: firstHit["used"].(string), URL: firstHit["url"].(string)}
	}
	return &WristBand{Token: token, URL: "Sorry, that token was already used"}
}

func enqueueRedemption(wristband *WristBand) {
	message := fmt.Sprintf("Wristband token %s used", wristband.Token)
	rmqlib.EnqueueMessage([]byte(message))
}
