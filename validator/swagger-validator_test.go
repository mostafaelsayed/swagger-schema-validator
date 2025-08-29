package swagger_validator

import (
	"log"
	"os"
	"testing"
)

func dataFromFiles(payload_file_name string, swagger_file_name string) []string {

	content, err := os.ReadFile(payload_file_name)

	if err != nil {
		log.Fatal("error reading data: " + err.Error())
	}

	swagger_content, yaml_err := os.ReadFile(swagger_file_name)

	if yaml_err != nil {
		log.Fatal("error reading swagger file: " + yaml_err.Error())
	}

	return []string{string(content[:]), string(swagger_content[:])}
}

func TestInvalidUserData1(t *testing.T) {
	errors_count := 6
    var want []string = make([]string, errors_count)
	want[0] = "User.username: prop username is missing but required"
	want[1] = "User.password: expected type string but found float64"
	want[2] = "User.job.location.id: expected type string but found float64"
	want[3] = "User.job.location.country.label: prop label is missing but required"
	want[4] = "User.job.location.map.test: unexpected prop pod"
	want[5] = "User.job.location.map.test.a.b[][0]: expected type integer but found string"

	data := dataFromFiles("../samples/invalid-data-1.json", "../swaggers/user.yaml")
    errors := Validate(data[0], data[1], "User")

    if len(errors) != errors_count {
        t.Errorf("errors length not equal to %v", errors_count)
    }
	not_found := make([]string, 0)
	for _, want_msg := range(want) {
		found := false
		for _, error_msg := range(errors) {
			if error_msg == want_msg {
				found = true
				break
			}
		}
		if !found {
			not_found = append(not_found, want_msg)
		}
	}

	if len(not_found) > 0 {
		t.Errorf("errors expected but not found: %v", not_found)
	}
}

func TestValidUserData1(t *testing.T) {
	errors_count := 0
	data := dataFromFiles("../samples/valid-data-1.json", "../swaggers/user.yaml")
    errors := Validate(data[0], data[1], "User")
    if len(errors) > errors_count {
        t.Errorf("errors length not equal to %v", errors_count)
    }
}