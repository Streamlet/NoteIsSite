package translator

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type markdownTranslator struct {
	defaultTranslator
}

func newMarkdownTranslator(path string) *markdownTranslator {
	t := new(markdownTranslator)
	t.path = path
	return t
}

func (t markdownTranslator) Translate() ([]byte, error) {
	content, err := t.defaultTranslator.Translate()
	if err != nil {
		return nil, err
	}

	extensions := parser.CommonExtensions
	p := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlContent := markdown.ToHTML(content, p, renderer)
	return htmlContent, nil
}
