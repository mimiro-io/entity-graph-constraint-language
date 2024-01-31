package egcl

import (
	"bytes"
	"encoding/json"
	datahub "github.com/mimiro-io/datahub-client-sdk-go"
	egdm "github.com/mimiro-io/entity-graph-data-model"
)

type RemoteDataProvider struct {
	client *datahub.Client
}

func NewRemoteDataProvider(client *datahub.Client) (*RemoteDataProvider, error) {
	resolver := &RemoteDataProvider{
		client: client,
	}

	return resolver, nil
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

func (r *RemoteDataProvider) GetDatasetEntities(name string) (datahub.EntityIterator, error) {
	return r.client.GetEntitiesStream(name, "", -1, false, true)
}

func (r *RemoteDataProvider) Hop(sourceEntityId string, reference string, datasets []string, inverse bool, limit int) (datahub.EntityIterator, error) {
	return r.client.RunHopQuery(sourceEntityId, reference, datasets, inverse, limit)
}
