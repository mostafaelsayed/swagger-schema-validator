package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"swagger_validator"
)

type Page struct {
    Title string
    PayloadBody []byte
	SwaggerBody []byte
	Errors []string
}

func loadPage(title string) *Page {
    return &Page{Title: title}
}

func loadPageWithErrors(title string, errors []string) *Page {
    return &Page{Title: title, Errors: errors}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    t, _ := template.ParseFiles(tmpl + ".html")
    t.Execute(w, p)
}

func viewSwaggerValidatorForm(w http.ResponseWriter, r *http.Request) {
	p := loadPage("Swagger Validator")
   	renderTemplate(w, "templates/validator-form", p)
}

func validateSwagger(w http.ResponseWriter, r *http.Request) {
	title := "Validation Results"
    body := r.FormValue("payload-body")
	swagger := r.FormValue("swagger-body")
	errors := swagger_validator.Validate(body, swagger, "User")
	log.Println(errors)
	p := loadPageWithErrors(title, errors)
   	renderTemplate(w, "templates/validation-results", p)
}

func validateApi(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatalf("error reading request data: %v", err)
	} else {
		swagger := body["swagger"].(string)
		data := body["data"].(string)
		errors := swagger_validator.Validate(data, swagger, "User")
		json.NewEncoder(w).Encode(errors)
	}
}

func main() {
	http.HandleFunc("/", viewSwaggerValidatorForm)
	http.HandleFunc("/validation-results", validateSwagger)
	http.HandleFunc("/api/validate", validateApi)
    log.Fatal(http.ListenAndServe(":8081", nil))
}