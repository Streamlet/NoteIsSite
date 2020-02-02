package translator

import "path/filepath"

type Translator interface {
	Translate() ([]byte, error)
}

func New(path string) Translator {
	switch filepath.Ext(path) {
	case ".md":
		return newMarkdownTranslator(path)
	default:
		return newDefaultTranslator(path)
	}
}
