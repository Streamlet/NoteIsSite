package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/Streamlet/NoteIsSite/util"
)

type SiteConfig interface {
	GetNoteDir() string
	GetTemplateDir() string
	GetPort() uint
	GetSock() string
}

type siteConfig struct {
	NoteDir     string `toml:"note_dir"`
	TemplateDir string `toml:"template_dir"`
	Port        uint   `toml:"port"` // If port is specified, sock MUST be empty string.
	Sock        string `toml:"sock"` // sIf sock is specified, port MUST be 0.
}

var siteConf *siteConfig

func LoadSiteConfig(configPath string) (SiteConfig, error) {
	util.Assert(siteConf == nil, "duplicately loading site config")
	conf := new(siteConfig)
	if _, err := toml.DecodeFile(configPath, conf); err != nil {
		return nil, err
	}
	if conf.NoteDir == "" {
		return nil, fmt.Errorf("note_dir MUST be set")
	}
	if conf.TemplateDir == "" {
		return nil, fmt.Errorf("template_dir MUST be set")
	}
	if conf.Port == 0 && conf.Sock == "" {
		return nil, fmt.Errorf("port or sock MUST be set")
	}
	if conf.Port > 0 && conf.Sock != "" {
		return nil, fmt.Errorf("port and sock can NOT be both set")
	}
	siteConf = conf
	return siteConf, nil
}

func GetSiteConfig() SiteConfig {
	util.Assert(siteConf != nil, "use site config before loading")
	return siteConf
}

func (c siteConfig) GetNoteDir() string {
	return c.NoteDir
}

func (c siteConfig) GetTemplateDir() string {
	return c.TemplateDir
}

func (c siteConfig) GetPort() uint {
	return c.Port
}
func (c siteConfig) GetSock() string {
	return c.Sock
}
