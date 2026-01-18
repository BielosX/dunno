package opensearch

import (
	"dunno/internal/aws"

	"github.com/caarlos0/env/v11"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v4/signer/awsv2"
)

type Config struct {
	OpenSearchAddress string `env:"OPEN_SEARCH_ADDRESS"`
}

var Client *opensearchapi.Client

func init() {
	var config Config
	err := env.Parse(&config)
	if err != nil {
		panic(err.Error())
	}
	signer, err := requestsigner.NewSignerWithService(aws.Config, "es")
	if err != nil {
		panic(err.Error())
	}
	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Signer:    signer,
			Addresses: []string{config.OpenSearchAddress},
		},
	})
	if err != nil {
		panic(err.Error())
	}
	Client = client
}
