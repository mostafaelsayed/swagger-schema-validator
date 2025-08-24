package main

import (
	"fmt"
	"log"
	"swagger_validator"
)

func main() {
	errors := swagger_validator.Validate("samples/data-1.json", "swaggers/user.yaml", "User")

	if len(errors) > 0 {
		log.Println("swagger errors found:")
		for index, error := range(errors) {
			log.Println(fmt.Sprint(index + 1) + ": " + error)
		}
	}
}