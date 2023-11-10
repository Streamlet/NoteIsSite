package translator

import (
	"bytes"
	"os"
)

type textTranslator struct {
	defaultTranslator
}

func newTextTranslator(path string) *textTranslator {
	t := new(textTranslator)
	t.path = path
	return t
}

func (t textTranslator) Translate() ([]byte, error) {
	content, err := os.ReadFile(t.path)
	if err != nil {
		return nil, err
	}

	htmlContent := htmlTrans(content)

	return htmlContent, nil
}

func htmlTrans(content []byte) []byte {
	content = bytes.Replace(content, []byte("&"), []byte("&amp;"), -1)
	content = bytes.Replace(content, []byte("<"), []byte("&lt;"), -1)
	content = bytes.Replace(content, []byte(">"), []byte("&gt;"), -1)
	content = bytes.Replace(content, []byte("\""), []byte("&quot;"), -1)
	content = bytes.Replace(content, []byte(" "), []byte("&nbsp;"), -1)
	content = bytes.Replace(content, []byte("\n"), []byte("<br />"), -1)
	return content
}
