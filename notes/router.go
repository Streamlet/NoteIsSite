package notes

import (
	"fmt"
	"github.com/Streamlet/NoteIsSite/util"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Router interface {
	Route(uri string) ([]byte, error)
}

type notesRouter struct {
	noteDir     string
	templateDir string

	uriNodeMap map[string]node
	lock       sync.RWMutex
}

type node interface {
	GetContent() ([]byte, error)
}

type dirItem struct {
	name  string
	isDir bool
}

type dirNode struct {
	absolutePath string
	subUris      []dirItem
}

type fileNode struct {
	absolutePath string
}

func NewRouter(noteDir string, templateDir string) (Router, error) {
	nr := new(notesRouter)
	nr.noteDir = noteDir
	nr.templateDir = templateDir
	nr.uriNodeMap = make(map[string]node)

	if err := nr.buildNotes("/", noteDir); err != nil {
		return nil, err
	}
	if err := watchDir(noteDir, nr); err != nil {
		return nil, err
	}

	return nr, nil
}

func (nr notesRouter) Route(uri string) (content []byte, err error) {
	nr.lock.RLock()
	n, ok := nr.uriNodeMap[uri]
	nr.lock.RUnlock()
	if !ok {
		return nil, os.ErrNotExist
	}
	return n.GetContent()
}

func (nr *notesRouter) buildNotes(baseUri string, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	self := new(dirNode)
	self.absolutePath = dir
	nr.uriNodeMap[baseUri] = self
	for _, f := range files {
		if f.IsDir() {
			subUri := f.Name()
			self.subUris = append(self.subUris, dirItem{subUri, true})
			if err := nr.buildNotes(baseUri+subUri+"/", dir+"/"+f.Name()); err != nil {
				return err
			}
		} else {
			subUri := f.Name()
			self.subUris = append(self.subUris, dirItem{subUri, false})
			n := new(fileNode)
			n.absolutePath = dir + "/" + f.Name()
			nr.uriNodeMap[baseUri+subUri] = n
		}
	}
	return nil
}

func (nr *notesRouter) FileCreated(path string) {
	f, err := os.Stat(path)
	if err != nil {
		return
	}

	relPath := strings.TrimPrefix(path, nr.noteDir)
	subUri := filepath.Base(relPath)
	baseUri := strings.TrimSuffix(relPath, subUri)

	defer nr.lock.Unlock()
	nr.lock.Lock()
	node, ok := nr.uriNodeMap[baseUri]
	if !ok {
		return
	}
	parent, ok := node.(*dirNode)
	util.Assert(ok, "%s is not dir?", baseUri)
	if f.IsDir() {
		parent.subUris = append(parent.subUris, dirItem{subUri, true})
		if err := nr.buildNotes(baseUri+subUri+"/", path); err != nil {
			log.Println(err.Error())
		}
	} else {
		parent.subUris = append(parent.subUris, dirItem{subUri, false})
		n := new(fileNode)
		n.absolutePath = path
		nr.uriNodeMap[baseUri+subUri] = n
	}
}

func (nr *notesRouter) FileRemoved(path string) {
	relPath := strings.TrimPrefix(path, nr.noteDir)
	subUri := filepath.Base(relPath)
	baseUri := strings.TrimSuffix(relPath, subUri)

	defer nr.lock.Unlock()
	nr.lock.Lock()

	node, ok := nr.uriNodeMap[baseUri]
	if !ok {
		return
	}
	parent, ok := node.(*dirNode)
	util.Assert(ok, "%s is not dir?", baseUri)
	for i, di := range parent.subUris {
		if di.name == subUri {
			parent.subUris = append(parent.subUris[:i], parent.subUris[i+1:]...)
			break
		}
	}
	for uri, _ := range nr.uriNodeMap {
		if strings.HasPrefix(uri, relPath) {
			delete(nr.uriNodeMap, uri)
		}
	}

}

func (nr *notesRouter) FileChanged(path string) {

}

func (n dirNode) GetContent() ([]byte, error) {
	sb := new(strings.Builder)
	for _, subUri := range n.subUris {
		href := subUri.name
		if subUri.isDir {
			href += "/"
		}
		sb.WriteString(fmt.Sprintf(`<a href="%s">%s</a><br />`, href, subUri.name))
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
