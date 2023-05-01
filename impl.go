package aristoteles

import (
	"crypto/x509"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/odysseia-greek/aristoteles/models"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client interface {
	Query() Query
	Document() Document
	Index() Index
	Builder() Builder
	Health() Health
	Access() Access
}

type Query interface {
	Match(index string, request map[string]interface{}) (*models.Response, error)
	MatchWithSort(index, mode, sort string, size int, request map[string]interface{}) (*models.Response, error)
	MatchWithScroll(index string, request map[string]interface{}) (*models.Response, error)
	MatchAggregate(index string, request map[string]interface{}) (*models.Aggregations, error)
}

type Document interface {
	Create(index string, body []byte) (*models.CreateResult, error)
	Update(index, id string, body []byte) (*models.CreateResult, error)
}

type Index interface {
	CreateDocument(index string, body []byte) (*models.CreateResult, error)
	Create(index string, request map[string]interface{}) (*models.IndexCreateResult, error)
	Delete(index string) (bool, error)
}

type Builder interface {
	MatchQuery(term, queryWord string) map[string]interface{}
	MatchAll() map[string]interface{}
	MultipleMatch(mappedFields []map[string]string) map[string]interface{}
	MultiMatchWithGram(queryWord string) map[string]interface{}
	Aggregate(aggregate, field string) map[string]interface{}
	FilteredAggregate(term, queryWord, aggregate, field string) map[string]interface{}
	SearchAsYouTypeIndex(searchWord string) map[string]interface{}
	Index() map[string]interface{}
}

type Health interface {
	Check(ticks, tick time.Duration) bool
	Info() (elasticHealth models.DatabaseHealth)
}

type Access interface {
	CreateRole(name string, roleRequest models.CreateRoleRequest) (bool, error)
	CreateUser(name string, userCreation models.CreateUserRequest) (bool, error)
}

type Elastic struct {
	document *DocumentImpl
	query    *QueryImpl
	index    *IndexImpl
	builder  *BuilderImpl
	health   *HealthImpl
	access   *AccessImpl
}

func NewClient(config models.Config) (Client, error) {
	//https://patorjk.com/software/taag/#p=display&f=Crawford2&t=ARISTOTELES
	log.Print("\n  ____  ____   ____ _____ ______   ___   ______    ___  _        ___  _____\n /    ||    \\ |    / ___/|      | /   \\ |      |  /  _]| |      /  _]/ ___/\n|  o  ||  D  ) |  (   \\_ |      ||     ||      | /  [_ | |     /  [_(   \\_ \n|     ||    /  |  |\\__  ||_|  |_||  O  ||_|  |_||    _]| |___ |    _]\\__  |\n|  _  ||    \\  |  |/  \\ |  |  |  |     |  |  |  |   [_ |     ||   [_ /  \\ |\n|  |  ||  .  \\ |  |\\    |  |  |  |     |  |  |  |     ||     ||     |\\    |\n|__|__||__|\\_||____|\\___|  |__|   \\___/   |__|  |_____||_____||_____| \\___|\n                                                                           \n")
	log.Print(strings.Repeat("~", 37))
	log.Print("\"Τριών δει παιδεία: φύσεως, μαθήσεως, ασκήσεως.\"")
	log.Print("\"Education needs these three: natural endowment, study, practice.\"")
	log.Print(strings.Repeat("~", 37))

	var err error
	var esClient *elasticsearch.Client
	if config.ElasticCERT != "" {
		esClient, err = createWithTLS(config)
		if err != nil {
			return nil, err
		}
	} else {
		esClient, err = create(config)
		if err != nil {
			return nil, err
		}
	}

	query, err := NewQueryImpl(esClient)
	if err != nil {
		return nil, err
	}

	index, err := NewIndexImpl(esClient)
	if err != nil {
		return nil, err
	}

	health, err := NewHealthImpl(esClient)
	if err != nil {
		return nil, err
	}

	access, err := NewAccessImpl(esClient)
	if err != nil {
		return nil, err
	}

	document, err := NewDocumentImpl(esClient)
	if err != nil {
		return nil, err
	}

	builder := NewBuilderImpl()

	es := &Elastic{query: query, index: index, builder: builder, health: health, access: access, document: document}

	return es, nil
}

func NewMockClient(fixtureFile string, statusCode int) (Client, error) {
	esClient, err := CreateMockClient(fixtureFile, statusCode)
	if err != nil {
		return nil, err
	}

	query, err := NewQueryImpl(esClient)
	if err != nil {
		return nil, err
	}

	index, err := NewIndexImpl(esClient)
	if err != nil {
		return nil, err
	}

	health, err := NewHealthImpl(esClient)
	if err != nil {
		return nil, err
	}

	access, err := NewAccessImpl(esClient)
	if err != nil {
		return nil, err
	}

	document, err := NewDocumentImpl(esClient)
	if err != nil {
		return nil, err
	}

	builder := NewBuilderImpl()

	es := &Elastic{query: query, index: index, builder: builder, health: health, access: access, document: document}

	return es, nil
}

func create(config models.Config) (*elasticsearch.Client, error) {
	log.Print("creating elasticClient")

	cfg := elasticsearch.Config{
		Username:  config.Username,
		Password:  config.Password,
		Addresses: []string{config.Service},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Printf("Error creating the client: %s", err)
		return nil, err
	}

	return es, nil
}

func createWithTLS(config models.Config) (*elasticsearch.Client, error) {
	log.Print("creating elasticClient with tls")

	caCert := []byte(config.ElasticCERT)

	// --> Clone the default HTTP transport

	tp := http.DefaultTransport.(*http.Transport).Clone()

	// --> Initialize the set of root certificate authorities
	//
	var err error

	if tp.TLSClientConfig.RootCAs, err = x509.SystemCertPool(); err != nil {
		log.Fatalf("ERROR: Problem adding system CA: %s", err)
	}

	// --> Add the custom certificate authority
	//
	if ok := tp.TLSClientConfig.RootCAs.AppendCertsFromPEM(caCert); !ok {
		log.Fatalf("ERROR: Problem adding CA from file %q", caCert)
	}

	cfg := elasticsearch.Config{
		Username:  config.Username,
		Password:  config.Password,
		Addresses: []string{config.Service},
		Transport: tp,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Printf("Error creating the client: %s", err)
		return nil, err
	}

	return es, nil
}

func (e *Elastic) Query() Query {
	if e == nil {
		return nil
	}
	return e.query
}

func (e *Elastic) Document() Document {
	if e == nil {
		return nil
	}
	return e.document
}

func (e *Elastic) Index() Index {
	if e == nil {
		return nil
	}
	return e.index
}

func (e *Elastic) Health() Health {
	if e == nil {
		return nil
	}
	return e.health
}

func (e *Elastic) Builder() Builder {
	if e == nil {
		return nil
	}
	return e.builder
}

func (e *Elastic) Access() Access {
	if e == nil {
		return nil
	}
	return e.access
}
