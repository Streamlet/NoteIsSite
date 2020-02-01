package notes

import (
	"github.com/Streamlet/NoteIsSite/template"
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

	templateExecutor template.Executor
}

type node interface {
	GetContent() ([]byte, error)
}

type dirItem struct {
	name  string
	isDir bool
}

type dirNode struct {
	fileNode
	subUris []dirItem
}

type fileNode struct {
	isNote           bool
	templateExecutor template.Executor
	absolutePath     string
	uri              string
}

func NewRouter(noteDir string, templateDir string) (Router, error) {
	w, err := newWatcher()
	if err != nil {
		return nil, err
	}

	nr := new(notesRouter)
	nr.noteDir = noteDir
	nr.templateDir = templateDir
	nr.uriNodeMap = make(map[string]node)
	nr.templateExecutor, err = template.NewExecutor(templateDir)
	if err != nil {
		return nil, err
	}

	if err := nr.buildTree("/", noteDir, true); err != nil {
		return nil, err
	}
	if err := w.addDirs(noteDir); err != nil {
		return nil, err
	}
	if err := nr.buildTree("/assets/", templateDir+"/assets", false); err != nil {
		return nil, err
	}
	if err := w.addDirs(templateDir); err != nil {
		return nil, err
	}

	w.watch(nr)

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

func (nr *notesRouter) buildTree(baseUri string, dir string, isNote bool) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	self := new(dirNode)
	self.isNote = isNote
	self.templateExecutor = nr.templateExecutor
	self.absolutePath = dir
	self.uri = baseUri
	nr.uriNodeMap[baseUri] = self
	for _, f := range files {
		if f.IsDir() {
			subUri := f.Name()
			if isNote {
				self.subUris = append(self.subUris, dirItem{subUri, true})
			}
			if err := nr.buildTree(baseUri+subUri+"/", dir+"/"+f.Name(), isNote); err != nil {
				return err
			}
		} else {
			subUri := f.Name()
			if isNote {
				self.subUris = append(self.subUris, dirItem{subUri, false})
			}
			n := new(fileNode)
			n.isNote = isNote
			n.templateExecutor = nr.templateExecutor
			n.absolutePath = dir + "/" + f.Name()
			n.uri = baseUri + subUri
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

	isNote := strings.HasPrefix(path, nr.noteDir)
	relPath := ""
	if isNote {
		relPath = strings.TrimPrefix(path, nr.noteDir)
	} else {
		relPath = strings.TrimPrefix(path, nr.templateDir)
	}
	subUri := filepath.Base(relPath)
	baseUri := strings.TrimSuffix(relPath, subUri)

	if !isNote && baseUri == "/" {
		if err := nr.templateExecutor.Update(nr.templateDir); err != nil {
			log.Println(err.Error())
		}
		return
	}

	defer nr.lock.Unlock()
	nr.lock.Lock()
	node, ok := nr.uriNodeMap[baseUri]
	if !ok {
		return
	}
	parent, ok := node.(*dirNode)
	util.Assert(ok, "%s is not dir?", baseUri)
	if f.IsDir() {
		if isNote {
			parent.subUris = append(parent.subUris, dirItem{subUri, true})
		}
		if err := nr.buildTree(baseUri+subUri+"/", path, isNote); err != nil {
			log.Println(err.Error())
		}
	} else {
		if isNote {
			parent.subUris = append(parent.subUris, dirItem{subUri, false})
		}
		n := new(fileNode)
		n.isNote = isNote
		n.templateExecutor = nr.templateExecutor
		n.absolutePath = path
		n.uri = baseUri + subUri
		nr.uriNodeMap[baseUri+subUri] = n
	}
}

func (nr *notesRouter) FileRemoved(path string) {
	isNote := strings.HasPrefix(path, nr.noteDir)
	relPath := ""
	if isNote {
		relPath = strings.TrimPrefix(path, nr.noteDir)
	} else {
		relPath = strings.TrimPrefix(path, nr.templateDir)
	}
	subUri := filepath.Base(relPath)
	baseUri := strings.TrimSuffix(relPath, subUri)

	if !isNote && baseUri == "/" {
		if err := nr.templateExecutor.Update(nr.templateDir); err != nil {
			log.Println(err.Error())
		}
		return
	}

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
			if isNote {
				parent.subUris = append(parent.subUris[:i], parent.subUris[i+1:]...)
			}
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
	isNote := strings.HasPrefix(path, nr.noteDir)
	relPath := ""
	if isNote {
		relPath = strings.TrimPrefix(path, nr.noteDir)
	} else {
		relPath = strings.TrimPrefix(path, nr.templateDir)
	}
	subUri := filepath.Base(relPath)
	baseUri := strings.TrimSuffix(relPath, subUri)

	if !isNote && baseUri == "/" {
		if err := nr.templateExecutor.Update(nr.templateDir); err != nil {
			log.Println(err.Error())
		}
		return
	}
}

func (n dirNode) GetContent() ([]byte, error) {
	util.Assert(n.isNote, "check code")
	var data template.CategoryData
	for _, subUri := range n.subUris {
		if subUri.isDir {
			data.SubCategories = append(data.SubCategories, template.SubItem{Name: subUri.name, Uri: subUri.name + "/"})
		} else {
			data.Contents = append(data.Contents, template.SubItem{Name: subUri.name, Uri: subUri.name})
		}
	}
	if n.uri == "/" {
		return n.templateExecutor.GetIndex(data.IndexData)
	} else {
		data.Name = filepath.Base(n.uri)
		return n.templateExecutor.GetCategory(data)
	}
}

func (n fileNode) GetContent() ([]byte, error) {
	c, err := ioutil.ReadFile(n.absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return n.templateExecutor.Get404(), err
		} else {
			return n.templateExecutor.Get500(), err
		}
	}
	if !n.isNote {
		return c, nil
	}
	var data template.ContentData
	data.Title = filepath.Base(n.uri)
	data.Content = string(c)
	return n.templateExecutor.GetContent(data)
}
