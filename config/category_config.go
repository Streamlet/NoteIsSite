package config

import "github.com/BurntSushi/toml"

type CategoryConfig struct {
	Name  string `toml:"name"`
	Index string `toml:"index"`
}

func GetCategoryConfig(dirPath string) (*CategoryConfig, error) {
	configPath := dirPath + "/" + GetSiteConfig().Note.CategoryConfigFile
	conf := new(CategoryConfig)
	if _, err := toml.DecodeFile(configPath, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
