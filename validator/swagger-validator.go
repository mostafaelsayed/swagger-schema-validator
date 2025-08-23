package swagger_validator

import (
	"fmt"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"slices"
	"strings"
	"gopkg.in/yaml.v3"
)

func Validate(payload_file_name string, swagger_file_name string, schema_name string) {
	log.SetPrefix("Swagger_Validator: ")
	log.SetFlags(0)

	content, err := os.ReadFile(payload_file_name)

	if err != nil {
		log.Fatal("error reading data: " + err.Error())
	}

	var data map[string]any
	err = json.Unmarshal(content, &data)

	if err != nil {
		log.Fatal("error unmarshal data: " + err.Error())
	}

	swagger_content, yaml_err := os.ReadFile(swagger_file_name)

	if yaml_err != nil {
		log.Fatal("error reading swagger file: " + err.Error())
	}

	var swagger map[string]map[string]map[string]any
	swagger_err := yaml.Unmarshal(swagger_content, &swagger)

	if err != nil {
		log.Fatal("error unmarshal: " + err.Error())
	}

	if swagger_err != nil {
		log.Fatal("error unmarshal swagger: " + swagger_err.Error())
	}

	errors := validateSchema(data, swagger, schema_name, schema_name)

	if len(errors) > 0 {
		log.Println("swagger errors found:")
		for index, error := range(errors) {
			log.Println(fmt.Sprint(index + 1) + ": " + error)
		}
		os.Exit(1)
	}

	log.Println("data is valid!")
}

func validateSchema(data any, swagger map[string]map[string]map[string]any, schema_name string, schema_path string) []string {
	components := swagger["components"]
	schemas := components["schemas"]
	schema := schemas[schema_name].(map[string]any)
	var errors []string

	switch schema["type"] {
	case "object":
		errors = ValidateObject(schema, schema_name, data, swagger, schema_path)
	case "string":
		errors = validateString(data, schema_path)
	case "integer":
		errors = validateInteger(data, schema_path)
	case "number":
		errors = validateNumber(data, schema_path)
	case "array":
		errors = validateArray(data, schema, swagger, schema_path)
	}

	return errors
}

func validateArray(data any, schema map[string]any, swagger map[string]map[string]map[string]any, schema_path string) []string {
	var errors []string
	arr, ok := data.([]any)
	if !ok {
		errors = append(errors, schema_path + ": expected array but found " + reflect.TypeOf(data).String())
		return errors
	}
	items := schema["items"].(map[string]any)
	errors = validateArrayItems(items, arr, swagger, schema_path)

	return errors
}

func validateArrayItems(items map[string]any, data []any, swagger map[string]map[string]map[string]any, schema_path string) []string {
	var errors []string
	for _, val := range(data) {
		if items["type"] == "string" {
			errors = slices.Concat(errors, validateString(val, schema_path))
		} else if items["type"] == "integer" {
			errors = slices.Concat(errors, validateInteger(val, schema_path))
		} else if items["type"] == "number" {
			errors = slices.Concat(errors, validateNumber(val, schema_path))
		} else if items["$ref"] != nil {
			ref := items["$ref"].(string)
			ref_splitted := strings.Split(ref, "/")
			new_schema_name := ref_splitted[len(ref_splitted) - 1]
			errors = slices.Concat(validateSchema(val.(map[string]any), swagger, new_schema_name, schema_path + "[]"))
		}
	}
	return errors
}

func ValidateObject(schema map[string]any, schema_name string, data any, swagger map[string]map[string]map[string]any, schema_path string) []string {
	var errors []string
	data_map, ok := data.(map[string]any)

	if !ok {
		errors = append(errors, schema_path + ": type expected is object but found " + reflect.TypeOf(data).String())
		return errors
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		return errors
	}
	for prop1 := range(data_map) {
		_, exists := props[prop1]
		if !exists {
			additional_exists := schema["additionalProperties"]
			if additional_exists == nil || !additional_exists.(bool) {
				errors = append(errors, schema_path + ": unexpected prop " + prop1)
			}
		}
	}
	for prop, val := range(props) {
		new_schema_path := schema_path + "." + prop
		errors = slices.Concat(errors, validateProp(prop, val, schema, schema_name, swagger, data.(map[string]any)[prop], new_schema_path))
	}

	return errors
}

func validateProp(prop string, val any, schema map[string]any, schema_name string, swagger map[string]map[string]map[string]any, data any, schema_path string) []string {
	var errors []string
	if data == nil {
		errors = checkObjectPropRequired(prop, schema["required"].([]any), schema_path)
	} else {
		new_val := val.(map[string]any)
		if new_val["type"] == "string" {
			errors = validateString(data, schema_path)
		} else if new_val["type"] == "integer" {
			errors = validateInteger(data, schema_path)
		} else if new_val["type"] == "number" {
			errors = validateNumber(data, schema_path)
		} else if new_val["$ref"] != nil {
			ref := new_val["$ref"].(string)
			ref_splitted := strings.Split(ref, "/")
			new_schema_name := ref_splitted[len(ref_splitted) - 1]
			errors = validateSchema(data, swagger, new_schema_name, schema_path)
		} else if new_val["type"] == "object" {
			errors = ValidateObject(schema["properties"].(map[string]any)[prop].(map[string]any), schema_name, data, swagger, schema_path)
		}
	}

	return errors
}

func checkObjectPropRequired(prop string, required []any, schema_path string) []string {
	var errors []string
	if required != nil {
		not_exists := false
		for _, val1 := range(required) {
			if val1 == prop {
				not_exists = true
				break
			}
		}
		if (not_exists) {
			errors = append(errors, schema_path + ": prop " + prop + " is missing but required")
		}
	}
	return errors
}

func validateString(data any, schema_path string) []string {
	var errors []string
	_, ok := data.(string)
	if !ok {
		errors = append(errors, schema_path + ": expected type string but found " + reflect.TypeOf(data).String())
	}

	return errors
}

func validateInteger(data any, schema_path string) []string {
	var errors []string
	_, ok := data.(float64)
	if !ok {
		errors = append(errors, schema_path + ": expected type integer but found " + reflect.TypeOf(data).String())
	} else {
		str_val := fmt.Sprint(data)
		if strings.Contains(str_val, ".") {
			errors = append(errors, schema_path + ": expected type integer but found " + reflect.TypeOf(data).String())
		}
	}

	return errors
}

func validateNumber(data any, schema_path string) []string {
	var errors []string
	_, ok := data.(float64)
	if !ok {
		errors = append(errors, schema_path + ": expected type number but found " + reflect.TypeOf(data).String())
	}

	return errors
}