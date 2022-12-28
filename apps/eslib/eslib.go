package eslib

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

func GetESClient() *elasticsearch.Client {
	url := os.Getenv("ES_HOST_URL")
	user := os.Getenv("ES_USERNAME")
	esPasswdFilename := os.Getenv("ES_PASSWD_FILENAME")
	espass, err := os.ReadFile(esPasswdFilename)
	if err != nil {
		log.Fatalf("reading es password from file %s: %s", esPasswdFilename, err)
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

func GetSearchResult(queryBuf *bytes.Buffer) map[string]interface{} {
	es := GetESClient()
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
