package langfuse

import (
	"github.com/git-hulk/langfuse-go/pkg/comments"
	"github.com/git-hulk/langfuse-go/pkg/datasets"
	"github.com/git-hulk/langfuse-go/pkg/models"
	"github.com/git-hulk/langfuse-go/pkg/prompts"
	"github.com/git-hulk/langfuse-go/pkg/traces"

	"github.com/go-resty/resty/v2"
)

type LangFuse struct {
	ingestor *traces.Ingestor
	prompt   *prompts.Client
	model    *models.Client
	comment  *comments.Client
	dataset  *datasets.Client
	restyCli *resty.Client
}

func NewClient(host string, publicKey string, secretKey string) *LangFuse {
	restyCli := resty.New().
		SetBaseURL(host+"/api/public").
		SetBasicAuth(publicKey, secretKey)

	return &LangFuse{
		ingestor: traces.NewIngestor(restyCli),
		prompt:   prompts.NewClient(restyCli),
		model:    models.NewClient(restyCli),
		comment:  comments.NewClient(restyCli),
		dataset:  datasets.NewClient(restyCli),
		restyCli: restyCli,
	}
}

func (c *LangFuse) StartTrace(name string) *traces.Trace {
	return c.ingestor.StartTrace(name)
}

func (c *LangFuse) Prompts() *prompts.Client {
	return c.prompt
}

func (c *LangFuse) Models() *models.Client {
	return c.model
}

func (c *LangFuse) Comments() *comments.Client {
	return c.comment
}

func (c *LangFuse) Datasets() *datasets.Client {
	return c.dataset
}

func (c *LangFuse) Close() error {
	return c.ingestor.Close()
}
