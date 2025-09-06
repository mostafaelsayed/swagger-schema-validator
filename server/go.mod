module main

go 1.25.0

replace swagger_validator => ../validator

require swagger_validator v1.0.0

require (
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
