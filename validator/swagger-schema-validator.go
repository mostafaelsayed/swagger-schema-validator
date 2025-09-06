package swagger_validator

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strings"
	"time"
	"gopkg.in/yaml.v3"
)

func Validate(payload string, swagger_content string, schema_name string) []string {
	log.SetPrefix("Swagger Validator Log [" + time.Now().Format("2006-01-02T15:04:05.000Z") + "] ")
	log.SetFlags(0)
	if schema_name == "" {
		return []string{"Please specify a schema name"}
	}
	var data map[string]any
	data_err := json.Unmarshal([]byte(payload), &data)

	if data_err != nil {
		log.Printf("error unmarshal payload: %v", data_err.Error())
		return []string{data_err.Error()}
	}

	var swagger map[string]map[string]map[string]any
	swagger_err := yaml.Unmarshal([]byte(swagger_content), &swagger)

	if swagger_err != nil {
		log.Printf("error unmarshal swagger: %v", swagger_err.Error())
		return []string{swagger_err.Error()}
	}

	components := swagger["components"]
	schemas := components["schemas"]
	schema := schemas[schema_name]
	if schema == nil {
		return []string{"Unknown schema " + schema_name}
	}
	schema_map := schema.(map[string]any)
	
	return validateSchema(data, schema_map, schemas, swagger, schema_name, schema_name)
}

func validateVal(data any, main_schema map[string]any, schema map[string]any, schemas map[string]any, swagger map[string]map[string]map[string]any, schema_name string, schema_path string) []string {
	var errors []string
	schema_type := schema["type"]

	if schema_type == "object" {
		errors = ValidateObject(schemas, main_schema, schema, schema_name, data, swagger, schema_path)
	} else if schema_type == "string" {
		error_msg := validateString(data, schema_path)
		if error_msg != "" {
			errors = append(errors, error_msg)
		}
	} else if schema_type == "integer" {
		error_msg := validateInteger(data, schema_path)
		if error_msg != "" {
			errors = append(errors, error_msg)
		}
	} else if schema_type == "number" {
		error_msg := validateNumber(data, schema_path)
		if error_msg != "" {
			errors = append(errors, error_msg)
		}
	} else if schema_type == "boolean" {
		error_msg := validateBoolean(data, schema_path)
		if error_msg != "" {
			errors = append(errors, error_msg)
		}
	} else if schema["$ref"] != nil {
		ref := schema["$ref"].(string)
		ref_splitted := strings.Split(ref, "/")
		new_schema_name := ref_splitted[len(ref_splitted) - 1]
		errors = validateSchema(data, main_schema, schemas, swagger, new_schema_name, schema_path)
	} else if schema["type"] == "array" {
		errors = validateArray(data, schemas, main_schema, schema, swagger, schema_name, schema_path)
	}

	return errors
}

func validateSchema(data any, main_schema map[string]any, schemas map[string]any, swagger map[string]map[string]map[string]any, schema_name string, schema_path string) []string {
	var errors []string
	schema := schemas[schema_name].(map[string]any)
	new_errors := validateVal(data, main_schema, schema, schemas, swagger, schema_name, schema_path)

	if len(new_errors) > 0 {
		errors = slices.Concat(errors, new_errors)
	}

	return errors
}

func validateArray(data any, schemas map[string]any, main_schema map[string]any, schema map[string]any, swagger map[string]map[string]map[string]any, schema_name string, schema_path string) []string {
	var errors []string
	arr, ok := data.([]any)
	if data == nil {
		return errors
	}
	if !ok {
		errors = append(errors, schema_path + ": expected array but found " + getType(data))
		return errors
	}

	items := schema["items"].(map[string]any)
	errors = validateArrayItems(schemas, main_schema, items, arr, swagger, schema_name, schema_path + "[]")

	return errors
}

func validateArrayItems(schemas map[string]any, main_schema map[string]any, items map[string]any, data []any, swagger map[string]map[string]map[string]any, schema_name string, schema_path string) []string {
	var errors []string
	for ind, val := range(data) {
		new_errors := validateVal(val, main_schema, items, schemas, swagger, schema_name, schema_path + "[" + fmt.Sprint(ind) + "]")
		if len(new_errors) > 0 {
			errors = slices.Concat(errors, new_errors)
		}
	}
	return errors
}

func ValidateObject(schemas map[string]any, main_schema map[string]any, schema map[string]any, schema_name string, data any, swagger map[string]map[string]map[string]any, schema_path string) []string {
	var errors []string
	if data == nil {
		return errors
	}
	data_map, ok := data.(map[string]any)

	if !ok {
		errors = append(errors, schema_path + ": type expected is object but found " + getType(data))
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
		errors = slices.Concat(errors, validateProp(prop, val, schemas, main_schema, schema, schema_name, swagger, data.(map[string]any)[prop], new_schema_path))
	}

	return errors
}

func validateProp(prop string, val any, schemas map[string]any, main_schema map[string]any, schema map[string]any, schema_name string, swagger map[string]map[string]map[string]any, data any, schema_path string) []string {
	var errors []string
	if data == nil && schema["required"] != nil {
		error_msg := validatePropRequired(prop, schema["required"].([]any), schema_path)
		if error_msg != "" {
			errors = append(errors, error_msg)
		}
	} else {
		new_val := val.(map[string]any)
		new_errors := validateVal(data, main_schema, new_val, schemas, swagger, schema_name, schema_path)
		if len(new_errors) > 0 {
			errors = slices.Concat(errors, new_errors)
		}
	}

	return errors
}

func validatePropRequired(prop string, required []any, schema_path string) string {
	var error_msg string
	if required != nil {
		is_required := false
		for _, val1 := range(required) {
			if val1 == prop {
				is_required = true
				break
			}
		}
		if (is_required) {
			error_msg =  schema_path + ": prop " + prop + " is missing but required"
		}
	}
	return error_msg
}

func validateString(data any, schema_path string) string {
	if data == nil {
		return ""
	}
	var error_msg string
	_, ok := data.(string)
	if !ok {
		error_msg = schema_path + ": expected type string but found " + getType(data)
	}

	return error_msg
}

func validateInteger(data any, schema_path string) string {
	if data == nil {
		return ""
	}
	var error_msg string
	_, ok := data.(float64)
	if !ok {
		error_msg = schema_path + ": expected type integer but found " + getType(data)
	} else {
		str_val := fmt.Sprint(data)
		if strings.Contains(str_val, ".") {
			error_msg = schema_path + ": expected type integer but found " + getType(data)
		}
	}

	return error_msg
}

func getType(data any) string {
	if data == nil {
		return "nil"
	}
	return reflect.TypeOf(data).Name()
}

func validateNumber(data any, schema_path string) string {
	if data == nil {
		return ""
	}
	var error_msg string
	_, ok := data.(float64)
	if !ok {
		error_msg = schema_path + ": expected type number but found " + getType(data)
	}

	return error_msg
}

func validateBoolean(data any, schema_path string) string {
	if data == nil {
		return ""
	}
	var error_msg string
	_, ok := data.(bool)
	if !ok {
		error_msg = schema_path + ": expected type boolean but found " + getType(data)
	}

	return error_msg
}