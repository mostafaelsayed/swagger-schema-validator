package main

import "swagger_validator"

func main() {
	swagger_validator.Validate("samples/data-1.json", "swaggers/user.yaml", "User")
}