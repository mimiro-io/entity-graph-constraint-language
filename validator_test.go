package egcl

import (
	egdm "github.com/mimiro-io/entity-graph-data-model"
	"os"
	"testing"
)

func TestBuildValidator(t *testing.T) {

	// load entities
	nsmanager := egdm.NewNamespaceContext()
	parser := egdm.NewEntityParser(nsmanager)
	parser.WithExpandURIs()

	// open file with reader
	file, err := os.Open("test_data/test_entities.json")
	if err != nil {
		t.Error(err)
	}

	ec, err := parser.LoadEntityCollection(file)
	if err != nil {
		t.Error(err)
	}

	// load schema
	schemaFile, err := os.ReadFile("test_data/egcl-sample.yaml")
	if err != nil {
		t.Error(err)
	}

	schema, err := parseYaml(schemaFile)
	if err != nil {
		t.Error(err)
	}

	// new validator with no remote validation
	v := NewValidator().WithSettings(&ValidatorSettings{})

	ok, violations, err := v.ValidateEntityCollection(schema, ec)
	if err != nil {
		t.Error(err)
	}

	if ok {
		t.Error("expected validation to fail")
	}

	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(violations))
	}
}
