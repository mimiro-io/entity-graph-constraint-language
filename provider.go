package egcl

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	datahub "github.com/mimiro-io/datahub-client-sdk-go"
	egdm "github.com/mimiro-io/entity-graph-data-model"
)

type RemoteDataProvider struct {
	client *datahub.Client
}

func NewRemoteDatahubQueryHelper(endpoint string) (*RemoteDataProvider, error) {
	client, err := datahub.NewClient(endpoint)
	if err != nil {
		return nil, err
	}

	resolver := &RemoteDataProvider{
		client: client,
	}

	return resolver, nil
}

func (r *RemoteDataProvider) WithClientKeyAndSecretAuth(authorizer string, audience string, key string, secret string) *RemoteDataProvider {
	r.client.WithClientKeyAndSecretAuth(authorizer, audience, key, secret)
	return r
}

func (r *RemoteDataProvider) WithPublicKeyAuth(clientID string, privateKey *rsa.PrivateKey) *RemoteDataProvider {
	r.client.WithPublicKeyAuth(clientID, privateKey)
	return r
}

func (r *RemoteDataProvider) WithAdminAuth(clientID string, clientSecret string) *RemoteDataProvider {
	r.client.WithAdminAuth(clientID, clientSecret)
	return r
}

func (r *RemoteDataProvider) GetEntity(id string, datasets []string) (*egdm.Entity, error) {
	qb := datahub.NewQueryBuilder()
	qb.WithEntityId(id)
	qb.WithDatasets(datasets)
	qb.WithNoPartialMerging(true)
	result, err := r.client.RunQuery(qb.Build())
	if err != nil {
		return nil, err
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	nsm := egdm.NewNamespaceContext()
	parser := egdm.NewEntityParser(nsm)
	reader := bytes.NewReader(jsonResult)
	egc, err := parser.LoadEntityCollection(reader)
	if err != nil {
		return nil, err
	}

	if len(egc.Entities) == 0 {
		return nil, nil
	}

	return egc.Entities[0], nil
}

func (r *RemoteDataProvider) GetReferencingEntities(id string, referenceClass string) ([]*egdm.Entity, error) {
	return nil, nil
}

/*
type ConstraintViolation struct {
	Constraint any
	Message    string
}

func NewConstraintViolation(constraint any, message string) *ConstraintViolation {
	return &ConstraintViolation{
		Constraint: constraint,
		Message:    message,
	}
} */
