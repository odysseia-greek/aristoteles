package aristoteles

import (
	"fmt"
	"github.com/odysseia-greek/aristoteles/models"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	certPathInPod            = "/app/config/certs/elastic-certificate.pem"
	elasticServiceDefault    = "http://localhost:9200"
	elasticServiceDefaultTlS = "https://localhost:9200"
	elasticUsernameDefault   = "elastic"
	elasticPasswordDefault   = "odysseia"
	EnvElasticService        = "ELASTIC_SEARCH_SERVICE"
	EnvElasticUser           = "ELASTIC_SEARCH_USER"
	EnvElasticPassword       = "ELASTIC_SEARCH_PASSWORD"
)

func HealthCheck(client Client) error {
	standardTicks := 120 * time.Second
	tick := 1 * time.Second

	healthy := client.Health().Check(standardTicks, tick)
	if !healthy {
		return fmt.Errorf("elasticClient unhealthy after %s ticks", standardTicks)
	}

	return nil
}

func ElasticService(tls bool) string {
	elasticService := os.Getenv(EnvElasticService)
	if elasticService == "" {
		if tls {
			log.Printf("setting %s to default: %s", EnvElasticService, elasticServiceDefaultTlS)
			elasticService = elasticServiceDefaultTlS
		} else {
			log.Printf("setting %s to default: %s", EnvElasticService, elasticServiceDefault)
			elasticService = elasticServiceDefault
		}
	}
	return elasticService
}

func ElasticConfig(env string, testOverwrite, tls bool) models.Config {
	elasticUser := os.Getenv(EnvElasticUser)
	if elasticUser == "" {
		log.Printf("setting %s to default: %s", EnvElasticUser, elasticUsernameDefault)
		elasticUser = elasticUsernameDefault
	}
	elasticPassword := os.Getenv(EnvElasticPassword)
	if elasticPassword == "" {
		log.Printf("setting %s to default: %s", EnvElasticPassword, elasticPasswordDefault)
		elasticPassword = elasticPasswordDefault
	}

	var elasticCert string
	if tls {
		elasticCert = string(GetCert(env, testOverwrite))
	}

	elasticService := ElasticService(tls)

	esConf := models.Config{
		Service:     elasticService,
		Username:    elasticUser,
		Password:    elasticPassword,
		ElasticCERT: elasticCert,
	}

	return esConf
}

func GetCert(env string, testOverWrite bool) []byte {
	var cert []byte
	if env == "LOCAL" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		certPath := filepath.Join(homeDir, ".odysseia", "current", "elastic-certificate.pem")

		cert, _ = ioutil.ReadFile(certPath)

		return cert
	}

	if testOverWrite {
		log.Print("trying to read cert file from file")
		certPath := filepath.Join("eratosthenes", "elastic-test-cert.pem")

		cert, _ = ioutil.ReadFile(certPath)

		return cert
	}

	log.Print("trying to read cert file from pod")
	cert, _ = ioutil.ReadFile(certPathInPod)

	return cert
}
