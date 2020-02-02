package translator

import "io/ioutil"

type defaultTranslator struct {
	path string
}

func newDefaultTranslator(path string) *defaultTranslator {
	t := new(defaultTranslator)
	t.path = path
	return t
}

func (t defaultTranslator) Translate() ([]byte, error) {
	return ioutil.ReadFile(t.path)
}
