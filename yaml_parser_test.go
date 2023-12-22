package egcl

import (
	"os"
	"testing"
)

func TestParseYAML(t *testing.T) {
	// load yaml file as bytes
	yamlFile, err := os.ReadFile("egcl-sample.yaml")
	if err != nil {
		t.Error(err)
	}

	schema, err := parseYaml(yamlFile)

	if err != nil {
		t.Error(err)
	}

	if schema == nil {
		t.Error("schema is nil")
	}

	if schema.EntityClasses == nil {
		t.Error("schema entity classes is nil")
	}

	// check there are 3 entity classes
	if len(schema.EntityClasses) != 3 {
		t.Errorf("expected schema to have 3 entity classes, got %d", len(schema.EntityClasses))
	}

	// check there are 6 constraints
	if len(schema.Constraints) != 6 {
		t.Errorf("expected schema to have 3 constraints, got %d", len(schema.Constraints))
	}
}
