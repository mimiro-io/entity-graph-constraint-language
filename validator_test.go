package egcl

import (
	"context"
	dh "github.com/mimiro-io/datahub"
	"github.com/mimiro-io/datahub-client-sdk-go"
	egdm "github.com/mimiro-io/entity-graph-data-model"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "datahub-egcl-test-")
	if err != nil {
		panic(err)
	}

	// create store and security folders
	os.MkdirAll(tmpDir+"/store", 0777)
	os.MkdirAll(tmpDir+"/security", 0777)

	// startup data hub instance
	dhi, err := StartTestDatahub(tmpDir, "10778")
	if err != nil {
		panic(err)
	}

	// load datasets
	client, err := datahub.NewClient("http://localhost:10778")
	if err != nil {
		panic(err)
	}

	// load entities
	nsmanager := egdm.NewNamespaceContext()
	parser := egdm.NewEntityParser(nsmanager)
	parser.WithExpandURIs()

	// open file with reader
	file, err := os.Open("test_data/test_entities.json")
	if err != nil {
		panic(err)
	}

	ec, err := parser.LoadEntityCollection(file)
	if err != nil {
		panic(err)
	}

	// create data set and add entities
	err = client.AddDataset("test", nil)
	if err != nil {
		panic(err)
	}

	err = client.StoreEntities("test", ec)

	code := m.Run()

	dhi.Stop(context.Background())
	os.RemoveAll(tmpDir)

	os.Exit(code)
}

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

	if len(violations) != 7 {
		t.Errorf("expected 7 violations, got %d", len(violations))
	}
}

func TestValidationOfRemoteDataset(t *testing.T) {
	// add some data to data hub instance

	// create sdk client
	client, err := datahub.NewClient("http://localhost:10778")

	provider, err := NewRemoteDataProvider(client)
	if err != nil {
		t.Error(err)
	}

	// setup validator
	schemaFile, err := os.ReadFile("test_data/egcl-sample.yaml")
	if err != nil {
		t.Error(err)
	}

	schema, err := parseYaml(schemaFile)
	if err != nil {
		t.Error(err)
	}

	// new validator with no remote validation
	v := NewValidator().WithSettings(&ValidatorSettings{}).WithDataProvider(provider)

	ok, violations, err := v.ValidateDataset(schema, "test")
	if err != nil {
		t.Error(err)
	}

	if ok {
		t.Error("expected validation errors")
	}

	if len(violations) != 7 {
		t.Errorf("expected 7 violations, got %d", len(violations))
	}

	ok, violations, err = v.ValidateSchema(schema)
	if err != nil {
		t.Error(err)
	}

	if ok {
		t.Error("expected validation errors")
	}

	if len(violations) != 7 {
		t.Errorf("expected 7 violations, got %d", len(violations))
	}
}

func StartTestDatahub(location string, port string) (*dh.DatahubInstance, error) {

	cfg, err := dh.LoadConfig("")
	if err != nil {
		return nil, err
	}
	cfg.Port = port
	cfg.StoreLocation = location + "/store"
	cfg.SecurityStorageLocation = location + "/security"

	dhi, err := dh.NewDatahubInstance(cfg)
	if err != nil {
		return nil, err
	}
	go dhi.Start()

	return dhi, err
}
