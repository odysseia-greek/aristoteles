package aristoteles

import (
	"bytes"
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/odysseia-greek/aristoteles/models"
	"io/ioutil"
	"log"
)

type DocumentImpl struct {
	es *elasticsearch.Client
}

func NewDocumentImpl(suppliedClient *elasticsearch.Client) (*DocumentImpl, error) {
	return &DocumentImpl{es: suppliedClient}, nil
}

func (d *DocumentImpl) Create(index string, body []byte) (*models.CreateResult, error) {
	var elasticResult models.CreateResult

	ctx := context.Background()
	res, err := esapi.CreateRequest{
		Index: index,
		Body:  bytes.NewReader(body),
	}.Do(ctx, d.es)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("%s: %s", errorMessage, res.Status())
	}

	jsonBody, _ := ioutil.ReadAll(res.Body)
	elasticResult, err = models.UnmarshalCreateResult(jsonBody)
	if err != nil {
		return nil, err
	}

	return &elasticResult, nil
}

func (d *DocumentImpl) Update(index, id string, body []byte) (*models.CreateResult, error) {
	var elasticResult models.CreateResult

	ctx := context.Background()
	res, err := esapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader([]byte(fmt.Sprintf(`{"doc":%s}`, body))),
	}.Do(ctx, d.es)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		jsonBody, _ := ioutil.ReadAll(res.Body)
		log.Print(jsonBody)
		return nil, fmt.Errorf("%s: %s", errorMessage, res.Status())
	}

	jsonBody, _ := ioutil.ReadAll(res.Body)
	elasticResult, err = models.UnmarshalCreateResult(jsonBody)
	if err != nil {
		return nil, err
	}

	return &elasticResult, nil
}
