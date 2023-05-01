package aristoteles

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/odysseia-greek/aristoteles/models"
	"io/ioutil"
	"log"
	"strings"
)

type IndexImpl struct {
	es *elasticsearch.Client
}

func NewIndexImpl(suppliedClient *elasticsearch.Client) (*IndexImpl, error) {
	return &IndexImpl{es: suppliedClient}, nil
}

func (i *IndexImpl) CreateDocument(index string, body []byte) (*models.CreateResult, error) {
	var elasticResult models.CreateResult
	bodyString := strings.NewReader(string(body))

	esRequest := esapi.IndexRequest{
		Body:       bodyString,
		Refresh:    "true",
		Index:      index,
		DocumentID: "",
	}

	res, err := esRequest.Do(context.Background(), i.es)
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

func (i *IndexImpl) Create(index string, request map[string]interface{}) (*models.IndexCreateResult, error) {
	query, err := toBuffer(request)
	if err != nil {
		return nil, err
	}

	var elasticResult models.IndexCreateResult
	indexRequest := esapi.IndicesCreateRequest{
		Index: index,
		Body:  &query,
	}

	res, err := indexRequest.Do(context.Background(), i.es)
	if err != nil {
		return &elasticResult, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("%s: %s", errorMessage, res.Status())
	}

	jsonBody, _ := ioutil.ReadAll(res.Body)
	elasticResult, err = models.UnmarshalIndexCreateResult(jsonBody)
	if err != nil {
		return nil, err
	}

	return &elasticResult, nil
}

func (i *IndexImpl) Update(index string, request map[string]interface{}) (*models.IndexCreateResult, error) {
	query, err := toBuffer(request)
	if err != nil {
		return nil, err
	}

	var elasticResult models.IndexCreateResult
	indexRequest := esapi.IndicesCreateRequest{
		Index: index,
		Body:  &query,
	}

	res, err := indexRequest.Do(context.Background(), i.es)
	if err != nil {
		return &elasticResult, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("%s: %s", errorMessage, res.Status())
	}

	jsonBody, _ := ioutil.ReadAll(res.Body)
	elasticResult, err = models.UnmarshalIndexCreateResult(jsonBody)
	if err != nil {
		return nil, err
	}

	return &elasticResult, nil
}

func (i *IndexImpl) Delete(index string) (bool, error) {
	log.Printf("deleting index: %s", index)

	res, err := i.es.Indices.Delete([]string{index})
	if err != nil {
		return false, err
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return false, err
	}

	if res.IsError() {
		l, _ := json.Marshal(r)
		return false, fmt.Errorf(string(l))
	}

	return r["acknowledged"].(bool), nil
}
