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
	server.GET("/redeem", redeem)

	server.Logger.Fatal(server.Start(":8080")) // Start server
}

func healthcheck(httpContext echo.Context) error {
	return httpContext.String(http.StatusOK, "I'm okay; how are you?")
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
	log.Printf("ES url: %s, user: %s, password: %s", url, user, espass)
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

func readWristBand(httpContext echo.Context) *WristBand {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"token": "youwonaprize",
			},
		},
	}
	var queryBuf bytes.Buffer
	if err := json.NewEncoder(&queryBuf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	es := getESClient()
	searchRes, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("tickets"),
		es.Search.WithBody(&queryBuf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer searchRes.Body.Close()

	if searchRes.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(searchRes.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				searchRes.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
	var results map[string]interface{}
	if err := json.NewDecoder(searchRes.Body).Decode(&results); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	log.Printf(
		"[%s] %d hits; took: %dms",
		searchRes.Status(),
		int(results["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(results["took"].(float64)),
	)
	for _, hit := range results["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}
	firstHit := results["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})
	log.Printf("%s", firstHit)
	wb := WristBand{Token: firstHit["token"].(string), Used: firstHit["used"].(string), URL: firstHit["url"].(string)}
	return &wb
}
