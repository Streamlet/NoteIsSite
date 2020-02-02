package note

import (
	"github.com/Streamlet/NoteIsSite/config"
	"github.com/Streamlet/NoteIsSite/note/translator"
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
	noteRoot     string
	templateRoot string

	uriNodeMap map[string]node
	lock       sync.RWMutex
	watcher    *watcher

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
	name    string
	index   string
	subUris []dirItem
}

type fileNode struct {
	isNote           bool
	templateExecutor template.Executor
	absolutePath     string
	uri              string
}

func NewRouter(noteRoot string, templateRoot string) (Router, error) {
	nr := new(notesRouter)
	nr.noteRoot = noteRoot
	nr.templateRoot = templateRoot

	var err error
	nr.templateExecutor, err = template.NewExecutor(templateRoot)
	if err != nil {
		return nil, err
	}

	err = nr.rebuild()
	if err != nil {
		return nil, err
	}

	return nr, nil
}

func (nr *notesRouter) rebuild() error {
	defer nr.lock.Unlock()
	nr.lock.Lock()

	if nr.watcher != nil {
		_ = nr.watcher.close()
		nr.watcher = nil
	}
	var err error
	nr.watcher, err = newWatcher()
	if err != nil {
		return err
	}

	nr.uriNodeMap = make(map[string]node)
	if err := nr.buildTree("/", nr.noteRoot, true, nil); err != nil {
		return err
	}
	if err := nr.watcher.addDirs(nr.noteRoot); err != nil {
		return err
	}
	for _, dir := range config.GetSiteConfig().Template.StaticDirs {
		if err := nr.buildTree("/"+dir+"/", nr.templateRoot+"/"+dir, false, nil); err != nil {
			return err
		}
	}
	if err := nr.watcher.addDirs(nr.templateRoot); err != nil {
		return err
	}

	nr.watcher.watch(nr)

	return nil
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

func (nr *notesRouter) buildTree(baseUri string, dir string, isNote bool, parent *dirNode) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	if parent == nil {
		parent = new(dirNode)
		parent.isNote = isNote
		parent.templateExecutor = nr.templateExecutor
		parent.absolutePath = dir
		parent.uri = baseUri
		if conf, err := config.GetCategoryConfig(parent.absolutePath); err == nil && conf != nil {
			parent.index = conf.Index
		}
		nr.uriNodeMap[parent.uri] = parent
	}

	for _, f := range files {
		if f.IsDir() {
			self := new(dirNode)
			self.isNote = isNote
			self.templateExecutor = nr.templateExecutor
			self.absolutePath = dir + "/" + f.Name()
			self.name = f.Name()
			if isNote {
				conf, err := config.GetCategoryConfig(self.absolutePath)
				if err != nil || conf == nil {
					continue
				}
				if conf.Name != "" {
					self.name = conf.Name
				}
				if conf.Index != "" {
					self.index = conf.Index
				}
			}
			self.uri = baseUri + self.name + "/"
			nr.uriNodeMap[self.uri] = self
			parent.subUris = append(parent.subUris, dirItem{self.name, true})
			if err := nr.buildTree(self.uri, dir+"/"+f.Name(), isNote, self); err != nil {
				return err
			}
		} else {
			subUri := f.Name()
			if isNote {
				matches := config.GetSiteConfig().Note.NoteFileRegExp.FindAllStringSubmatch(f.Name(), -1)
				if matches == nil {
					continue
				}
				if len(matches) > 0 && len(matches[0]) > 1 {
					subUri = matches[0][1]
				}
				parent.subUris = append(parent.subUris, dirItem{subUri, false})
			}
			self := new(fileNode)
			self.isNote = isNote
			self.templateExecutor = nr.templateExecutor
			self.absolutePath = dir + "/" + f.Name()
			self.uri = baseUri + subUri
			nr.uriNodeMap[self.uri] = self
		}
	}
	return nil
}

func (nr *notesRouter) findParent(parentPath string) (baseUri string, parent *dirNode) {
	for uri, node := range nr.uriNodeMap {
		if n, ok := node.(*dirNode); ok && n.absolutePath == parentPath {
			parent = n
			baseUri = uri
		}
	}
	return
}

func (nr *notesRouter) FileCreated(path string) {
	nr.fsNotify(path)
}

func (nr *notesRouter) FileRemoved(path string) {
	nr.fsNotify(path)
}

func (nr *notesRouter) FileChanged(path string) {
	nr.fsNotify(path)
}

func (nr *notesRouter) fsNotify(path string) {
	basename := filepath.Base(path)
	parentPath := strings.TrimSuffix(path, "/"+basename)
	if !strings.HasPrefix(path, nr.noteRoot) && parentPath == nr.templateRoot {
		if err := nr.templateExecutor.Update(nr.templateRoot); err != nil {
			log.Println(err.Error())
		}
	} else {
		_ = nr.rebuild()
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
	if n.index != "" {
		t := translator.New(n.absolutePath + "/" + n.index)
		content, err := t.Translate()
		if err != nil {
			if os.IsNotExist(err) {
				return n.templateExecutor.Get404(), err
			} else {
				return n.templateExecutor.Get500(), err
			}
		}
		data.Content = string(content)
	}
	if n.uri == "/" {
		return n.templateExecutor.GetIndex(data.IndexData)
	} else {
		data.Name = filepath.Base(n.uri)
		return n.templateExecutor.GetCategory(data)
	}
}

func (n fileNode) GetContent() ([]byte, error) {
	t := translator.New(n.absolutePath)
	content, err := t.Translate()
	if err != nil {
		if os.IsNotExist(err) {
			return n.templateExecutor.Get404(), err
		} else {
			return n.templateExecutor.Get500(), err
		}
	}
	if !n.isNote {
		return content, nil
	}
	var data template.ContentData
	data.Title = filepath.Base(n.uri)
	data.Content = string(content)
	return n.templateExecutor.GetContent(data)
}
