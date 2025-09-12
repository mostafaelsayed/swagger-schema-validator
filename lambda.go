package main

import (
	"context"
	"encoding/json"
	"log"
	"swagger_validator"
	"time"
	"github.com/aws/aws-lambda-go/lambda"
)

func validateApi(body map[string]any) ([]string, error) {
	schema_name := body["schema"].(string)
	swagger := body["swagger"].(string)
	data := body["data"].(string)
	errors := swagger_validator.Validate(data, swagger, schema_name)
	return errors, nil
}

func handleRequest(ctx context.Context, event json.RawMessage) ([]string, error) {
	var request map[string]any
	if err := json.Unmarshal(event, &request); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return nil, err
	}
	var body map[string]any
	err := json.Unmarshal([]byte(request["body"].(string)), &body)
	if err != nil {
		log.Printf("Error getting req body : %v", err)
		return nil, err
	}

	return validateApi(body)
}

func main() {
	log.SetPrefix("Server Log [" + time.Now().Format("2006-01-02T15:04:05.000Z") + "] ")
	log.SetFlags(0)
	lambda.Start(handleRequest)
}