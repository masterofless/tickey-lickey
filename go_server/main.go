package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func redeem(httpContext echo.Context) error {
	wristband := readWristBand(httpContext)
	json, _ := json.Marshal(wristband)
	return httpContext.String(http.StatusOK, string(json))
}

type WristBand struct {
	Token string
	URL   string
	Used  string
}

func getESClient() *elasticsearch.Client {
	filename := "/mnt/secrets-store/password"
	url := os.Getenv("ES_HOST_URL")
	user := os.Getenv("ES_USERNAME")
	espass, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("reading es password from file %s: %s", filename, err)
	}
	es, err := elasticsearch.NewClient(
		elasticsearch.Config{
			Addresses: []string{url},
			Username:  string(user),
			Password:  string(espass),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	info, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	if info.IsError() {
		log.Fatalf("Error: %s", info.String())
	}
	return es
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

func getSearchResult(queryBuf *bytes.Buffer) map[string]interface{} {
	es := getESClient()
	searchRes, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("tickets"),
		es.Search.WithBody(queryBuf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	if searchRes.IsError() {
		var errBody map[string]interface{}
		if err := json.NewDecoder(searchRes.Body).Decode(&errBody); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				searchRes.Status(),
				errBody["error"].(map[string]interface{})["type"],
				errBody["error"].(map[string]interface{})["reason"],
			)
		}
	}
	defer searchRes.Body.Close()

	var results map[string]interface{}
	if err := json.NewDecoder(searchRes.Body).Decode(&results); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	return results
}

func readWristBand(httpContext echo.Context) *WristBand {
	token := httpContext.Param("token")
	results := getSearchResult(getQuery(token))
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
