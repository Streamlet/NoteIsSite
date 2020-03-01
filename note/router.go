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

type subItem struct {
	uri   string
	isDir bool
}

type dirNode struct {
	nodeBasic
	index    string
	subItems []subItem
}

type fileNode struct {
	nodeBasic
}

type nodeBasic struct {
	isNote           bool
	templateExecutor template.Executor
	absolutePath     string
	absoluteUri      string
	name             string
	parent           *dirNode
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
		if err := nr.buildTree("/"+dir+"/", filepath.Join(nr.templateRoot, dir), false, nil); err != nil {
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
		return nr.templateExecutor.Get404(), os.ErrNotExist
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
		parent.absoluteUri = baseUri
		if conf, err := config.GetCategoryConfig(parent.absolutePath); err == nil && conf != nil {
			parent.index = conf.Index
		}
		nr.uriNodeMap[parent.absoluteUri] = parent
	}
	for _, f := range files {
		if f.IsDir() {
			self := new(dirNode)
			self.isNote = isNote
			self.templateExecutor = nr.templateExecutor
			self.absolutePath = filepath.Join(dir, f.Name())
			self.name = f.Name()
			self.parent = parent
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
			self.absoluteUri = baseUri + self.name + "/"
			nr.uriNodeMap[self.absoluteUri] = self
			parent.subItems = append(parent.subItems, subItem{self.name, true})
			if err := nr.buildTree(self.absoluteUri, self.absolutePath, isNote, self); err != nil {
				return err
			}
		} else {
			self := new(fileNode)
			self.isNote = isNote
			self.templateExecutor = nr.templateExecutor
			self.absolutePath = filepath.Join(dir, f.Name())
			self.name = f.Name()
			self.parent = parent
			if isNote {
				matches := config.GetSiteConfig().Note.NoteFileRegExp.FindAllStringSubmatch(f.Name(), -1)
				if matches == nil {
					continue
				}
				if len(matches) > 0 && len(matches[0]) > 1 {
					self.name = matches[0][1]
				}
				parent.subItems = append(parent.subItems, subItem{self.name, false})
			}
			self.absoluteUri = baseUri + self.name
			nr.uriNodeMap[self.absoluteUri] = self
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
	parentPath := strings.TrimSuffix(path, string(filepath.Separator)+basename)
	if !strings.HasPrefix(path, nr.noteRoot) && parentPath == nr.templateRoot {
		if err := nr.templateExecutor.Update(nr.templateRoot); err != nil {
			log.Println(err.Error())
		}
	} else {
		_ = nr.rebuild()
	}
}

func (n nodeBasic) GetParents() template.HasParentItems {
	var parents template.HasParentItems
	parents.Parents = make([][]template.ParentItem, 0)

	currentUri := strings.TrimRight(n.absoluteUri, "/")
	for p := n.parent; p != nil; p = p.parent {
		currentSubUri := strings.TrimPrefix(currentUri, p.absoluteUri)
		currentUri = strings.TrimRight(p.absoluteUri, "/")
		items := make([]template.ParentItem, 0, len(p.subItems))
		for _, sub := range p.subItems {
			var item template.ParentItem
			item.Name = sub.uri
			item.Uri = currentUri + "/" + sub.uri
			if sub.isDir {
				item.Uri += "/"
			}
			item.IsAncestor = sub.uri == currentSubUri
			items = append(items, item)
		}
		parents.Parents = append([][]template.ParentItem{items}, parents.Parents...)
	}
	return parents
}

func (n dirNode) GetContent() ([]byte, error) {
	util.Assert(n.isNote, "check code")
	var subData template.HasSubItems
	for _, subUri := range n.subItems {
		uri := subUri.uri
		if subUri.isDir {
			uri += "/"
		}
		subData.SubItems = append(subData.SubItems, template.BasicItem{Name: subUri.uri, Uri: uri, IsDir: subUri.isDir})
	}
	var contentData template.HasContent
	if n.index != "" {
		t := translator.New(filepath.Join(n.absolutePath, n.index))
		content, err := t.Translate()
		if err != nil {
			if os.IsNotExist(err) {
				return n.templateExecutor.Get404(), err
			} else {
				return n.templateExecutor.Get500(), err
			}
		}
		contentData.Content = string(content)
	}
	if n.absoluteUri == "/" {
		var data template.IndexData
		data.HasSubItems = subData
		data.HasContent = contentData
		return n.templateExecutor.GetIndex(data)
	} else {
		var data template.CategoryData
		data.Name = n.name
		data.HasSubItems = subData
		data.HasContent = contentData
		data.HasParentItems = n.GetParents()
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
	data.Title = n.name
	data.Content = string(content)
	data.HasParentItems = n.GetParents()
	return n.templateExecutor.GetContent(data)
}
