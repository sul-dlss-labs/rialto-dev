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
	neptuneClient := sparql.NewNeptuneClient(getEnv("RIALTO_SPARQL_ENDPOINT", ""))

	snsClient := message.NewClient(getEnv("RIALTO_SNS_ENDPOINT", ""),
		getEnv("RIALTO_TOPIC_ARN", ""),
		getEnv("AWS_REGION", ""))

	registry := runtime.NewRegistry(neptuneClient, snsClient)

	handler := runtime.NewHandler(registry)

	sparqlHandler := func(w http.ResponseWriter, req *http.Request) {
		body := reqToBody(req)
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
			events.APIGatewayProxyRequest{Headers: reqToHeaders(req),
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

	address := fmt.Sprintf("%s:%s", getEnv("HOST", ""), getEnv("PORT", "8080"))
	log.Printf("Starting server on %s", address)

	http.ListenAndServe(address, nil)
}

func reqToBody(req *http.Request) string {
	bodyBuf := new(bytes.Buffer)
	bodyBuf.ReadFrom(req.Body)
	return bodyBuf.String()
}

func reqToHeaders(req *http.Request) map[string]string {
	header := make(map[string]string)
	for key, value := range req.Header {
		header[key] = value[0]
	}
	return header
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}
