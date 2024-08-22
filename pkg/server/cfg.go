package server

import (
	"fmt"
	"io/ioutil"

	"github.com/galdor/emaild/pkg/smtp"
	"github.com/galdor/go-ejson"
	"github.com/galdor/go-log"
	"go.n16f.net/eyaml"
)

type ServerCfg struct {
	BuildId string `json:"-"`

	Logger      *log.LoggerCfg             `json:"logger"`
	SMTPServers map[string]*smtp.ServerCfg `json:"smtp_servers"`
}

func (cfg *ServerCfg) ValidateJSON(v *ejson.Validator) {
	v.CheckOptionalObject("logger", cfg.Logger)

	v.WithChild("smtp_servers", func() {
		for name, cfg := range cfg.SMTPServers {
			v.CheckObject(name, cfg)
		}
	})
}

func (cfg *ServerCfg) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read %q: %w", filePath, err)
	}

	return eyaml.Load(data, cfg)
}
