package config

import "github.com/BurntSushi/toml"

type ResourceConfig struct {
	Name    string `toml:"name"`
}

func GetResourceConfig(dirPath string) (*ResourceConfig, error) {
	configPath := dirPath + "/" + GetSiteConfig().Note.ResourceConfigFile
	conf := new(ResourceConfig)
	if _, err := toml.DecodeFile(configPath, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
