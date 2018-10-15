package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sul-dlss-labs/sparql-loader/message"
	"github.com/sul-dlss-labs/sparql-loader/runtime"
	"github.com/sul-dlss-labs/sparql-loader/sparql"
)

func main() {
	// Establish the clients and register the Lambda handler
	neptuneClient := sparql.NewNeptuneClient(os.Getenv("RIALTO_SPARQL_ENDPOINT"))

	snsClient := message.NewClient(os.Getenv("RIALTO_SNS_ENDPOINT"),
		os.Getenv("RIALTO_TOPIC_ARN"),
		os.Getenv("AWS_REGION"))

	registry := runtime.NewRegistry(neptuneClient, snsClient)

	handler := runtime.NewHandler(registry)

	sparqlHandler := func(w http.ResponseWriter, req *http.Request) {
		body := req_to_body(req)
		// Catch panics
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				log.Printf("Caught %v for %s", err, body)
				http.Error(w, err.Error(), 500)
			}
		}()

		// Will be invoking Lamda request handler.
		// Need to map from http.Request to events.APIGatewayProxyRequest
		resp, err := handler.RequestHandler(context.TODO(),
			events.APIGatewayProxyRequest{Headers: req_to_headers(req),
				Body: body})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		//Write the response to http.ResponseWriter
		w.WriteHeader(resp.StatusCode)
		io.WriteString(w, resp.Body)
	}

	http.HandleFunc("/sparql", sparqlHandler)

	address := fmt.Sprintf("%s:%s", get_env("HOST", ""), get_env("PORT", "8080"))
	log.Printf("Starting server on %s", address)

	http.ListenAndServe(address, nil)
}

func req_to_body(req *http.Request) string {
	body_buf := new(bytes.Buffer)
	body_buf.ReadFrom(req.Body)
	return body_buf.String()
}

func req_to_headers(req *http.Request) map[string]string {
	header := make(map[string]string)
	for key, value := range req.Header {
		header[key] = value[0]
	}
	return header
}

func get_env(key string, default_value string) string {
	value := os.Getenv(key)
	if value == "" {
		value = default_value
	}
	return value
}
