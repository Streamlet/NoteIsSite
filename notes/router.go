package notes

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Router interface {
	Route(uri string) ([]byte, error)
}

type notesRouter struct {
	uriNodeMap map[string]node
}

type node interface {
	GetContent() ([]byte, error)
}

type dirNode struct {
	absolutePath string
	subUris      []string // { "sub_file1", "sub_dir1/", ... }
}

type fileNode struct {
	absolutePath string
}

func NewRouter(noteDir string, templateDir string) (Router, error) {
	r := new(notesRouter)
	r.uriNodeMap = make(map[string]node)
	if err := buildNotes(r, "/", noteDir); err != nil {
		return nil, err
	}
	return r, nil
}

func (nr notesRouter) Route(uri string) (content []byte, err error) {
	n, ok := nr.uriNodeMap[uri]
	if !ok {
		return nil, os.ErrNotExist
	}
	return n.GetContent()
}

func buildNotes(r *notesRouter, baseUri string, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	self := new(dirNode)
	self.absolutePath = dir
	r.uriNodeMap[baseUri] = self
	for _, f := range files {
		if f.IsDir() {
			subUri := f.Name() + "/"
			self.subUris = append(self.subUris, subUri)
			if err := buildNotes(r, baseUri+subUri, dir+"/"+f.Name()); err != nil {
				return err
			}
		} else {
			subUri := f.Name()
			self.subUris = append(self.subUris, subUri)
			n := new(fileNode)
			n.absolutePath = dir + "/" + f.Name()
			r.uriNodeMap[baseUri+subUri] = n
		}
	}
	return nil
}

func (n dirNode) GetContent() ([]byte, error) {
	sb := new(strings.Builder)
	for _, subUri := range n.subUris {
		sb.WriteString(fmt.Sprintf(`<a href="%s">%s</a><br />`, subUri, subUri))
	}
	return []byte(sb.String()), nil
}

func (n fileNode) GetContent() ([]byte, error) {
	c, err := ioutil.ReadFile(n.absolutePath)
	if err != nil {
		return nil, err
	}
	return c, nil
}
