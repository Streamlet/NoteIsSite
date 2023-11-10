package translator

import "os"

type defaultTranslator struct {
	path string
}

func newDefaultTranslator(path string) *defaultTranslator {
	t := new(defaultTranslator)
	t.path = path
	return t
}

func (t defaultTranslator) Translate() ([]byte, error) {
	return os.ReadFile(t.path)
}
