package langfuse

import (
	"github.com/git-hulk/langfuse-go/pkg/prompts"
	"github.com/git-hulk/langfuse-go/pkg/traces"

	"github.com/go-resty/resty/v2"
)

type LangFuse struct {
	ingestor *traces.Ingestor
	prompt   *prompts.Client
	restyCli *resty.Client
}

func NewClient(host string, secretKey string, publicKey string) *LangFuse {
	restyCli := resty.New().
		SetBaseURL(host+"/api/public").
		SetBasicAuth(publicKey, secretKey)

	return &LangFuse{
		ingestor: traces.NewIngestor(restyCli),
		prompt:   prompts.NewClient(restyCli),
		restyCli: restyCli,
	}
}

func (c *LangFuse) StartTrace(name string) *traces.Trace {
	return c.ingestor.StartTrace(name)
}

func (c *LangFuse) Prompts() *prompts.Client {
	return c.prompt
}
