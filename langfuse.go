package main

import (
	"github.com/git-hulk/langfuse-go/trace"
	"github.com/go-resty/resty/v2"
)

type LangFuse struct {
	ingestor *trace.Ingestor
	restyCli *resty.Client
}

func NewClient(host string, secretKey string, publicKey string) *LangFuse {
	restyCli := resty.New().
		SetBaseURL(host+"/api/public").
		SetBasicAuth(publicKey, secretKey)

	return &LangFuse{
		ingestor: trace.NewIngestor(restyCli),
		restyCli: restyCli,
	}
}

func (c *LangFuse) StartTrace(name string) *trace.Trace {
	return c.ingestor.StartTrace(name)
}
