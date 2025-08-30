package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"swagger_validator"
	"path/filepath"
	"crypto/rand"
	"encoding/hex"
)

type Page struct {
	Title       string
	PayloadBody []byte
	SwaggerBody []byte
	Errors      []string
}

func generate_id() (string, error) {
	b := make([]byte, 16) // 128-bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil // 32 hex chars
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
		log.Printf("error reading request data: %v", err)
	} else {
		swagger := body["swagger"].(string)
		data := body["data"].(string)
		errors := swagger_validator.Validate(data, swagger, "User")
		json.NewEncoder(w).Encode(errors)
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request, file_name string) (*multipart.FileHeader, error) {
	log.Printf("uploading file with name: %v", file_name)
	var error_msg error
	r.ParseMultipartForm(32 >> 20)
	file, handler, err := r.FormFile(file_name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, fmt.Errorf("error reading file request: %v", err.Error())
	}
	defer file.Close()

	// Create a new file in the server's directory
	newpath := filepath.Join(".", "uploaded")
	os.MkdirAll(newpath, os.ModePerm)
	id, err := generate_id()

	if err != nil {
		log.Printf("error generating id: %v", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, fmt.Errorf("error generating id: %v", err.Error())
	}

	dst, err := os.Create("./uploaded/" + id + "_" + handler.Filename)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, fmt.Errorf("error uploading file with name %v: %v", handler.Filename, err.Error()) 
	}

	defer dst.Close()

	// Copy the uploaded file's content to the new file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		error_msg = fmt.Errorf("error coying file content with name %v: %v", handler.Filename, err.Error())
	}

	handler.Filename = id + "_" + handler.Filename

	return handler, error_msg
}

func validateApiWithFiles(w http.ResponseWriter, r *http.Request) {
	swagger_handler, swagger_err := uploadFile(w, r, "swagger.yaml")

	if swagger_err != nil {
		log.Print(swagger_err.Error())
		return
	}

	data_handler, data_error := uploadFile(w, r, "data.json")

	if data_error != nil {
		log.Print(data_error.Error())
		return
	}

	log.Printf("data file name after upload: %v", data_handler.Filename)
	data_content, err := os.ReadFile("./uploaded/" + data_handler.Filename)

	if err != nil {
		log.Printf("error reading data: %v", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("swagger file name after upload: %v", swagger_handler.Filename)
	swagger_content, yaml_err := os.ReadFile("./uploaded/" + swagger_handler.Filename)

	if yaml_err != nil {
		log.Printf("error reading swagger file: %v", yaml_err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	errors := swagger_validator.Validate(string(data_content[:]), string(swagger_content[:]), "User")
	json.NewEncoder(w).Encode(errors)
}

func main() {
	http.HandleFunc("/", viewSwaggerValidatorForm)
	http.HandleFunc("/validation-results", validateSwagger)
	http.HandleFunc("/api/validate", validateApi)
	http.HandleFunc("/api/validate/files", validateApiWithFiles)
	log.Fatal(http.ListenAndServe("localhost:8081", nil))
}
