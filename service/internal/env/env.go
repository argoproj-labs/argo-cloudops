package env

import (
	"errors"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

const appPrefix = "ARGO_CLOUDOPS"

type Vars struct {
	AdminSecret    string `split_words:"true" required:"true"`
	VaultRole      string `envconfig:"VAULT_ROLE" required:"true"`
	VaultSecret    string `envconfig:"VAULT_SECRET" required:"true"`
	VaultAddress   string `envconfig:"VAULT_ADDR" required:"true"`
	ArgoAddress    string `envconfig:"ARGO_ADDR" required:"true"`
	ArgoNamespace  string `envconfig:"WORKFLOW_EXECUTION_NAMESPACE" default:"argo"`
	ConfigFilePath string `envconfig:"CONFIG" default:"argo-cloudops.yaml"`
	SSHPEMFile     string `envconfig:"SSH_PEM_FILE" required:"true"`
	LogLevel       string `split_words:"true"`
	Port           int    `default:"8443"`
	DbHost         string `split_words:"true" required:"true"`
	DbUser         string `split_words:"true" required:"true"`
	DbPassword     string `split_words:"true" required:"true"`
	DbName         string `split_words:"true" required:"true"`
}

var (
	instance Vars
	once     sync.Once
	err      error
)

func GetEnv() (Vars, error) {
	once.Do(func() {
		err = envconfig.Process(appPrefix, &instance)
		if err != nil {
			return
		}
		err = instance.validate()
	})
	return instance, err
}

func (values Vars) validate() error {
	if len(values.AdminSecret) < 16 {
		return errors.New("admin secret must be at least 16 characers long")
	}
	return nil
}
