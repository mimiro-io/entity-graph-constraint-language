package main

import (
	"flag"
)

func main() {
	var closedWorldFlag bool
	var clientSecret string
	var clientKey string
	var server string
	var privateKey string
	var validateRelated bool
	var schema string
	var dataset string
	var authType string

	flag.StringVar(&server, "server", "", "Server")
	flag.StringVar(&authType, "authentication type", "none", "One of: client, key")
	flag.StringVar(&clientSecret, "clientSecret", "", "Client secret")
	flag.StringVar(&clientKey, "clientKey", "", "Client key")
	flag.StringVar(&privateKey, "privateKey", "", "Private key")
	flag.BoolVar(&closedWorldFlag, "closedWorld", false, "Closed world assumption. Only allow what is defined in the model.")
	flag.BoolVar(&validateRelated, "validateRelated", false, "validate related entities, check to see they exist and are of the correct type")
	flag.StringVar(&schema, "schema", "", "Schema file location or remote location")
	flag.StringVar(&dataset, "dataset", "", "Dataset file or URL")
	flag.Parse()
}

func Validate() {
}
