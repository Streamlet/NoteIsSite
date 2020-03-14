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
	"regexp"
	"strings"
	"sync"
)

type Router interface {
	Route(uri string) ([]byte, error)
}

type notesRouter struct {
	noteRoot     string
	templateRoot string

	uriNodeMap map[string]*node
	lock       sync.RWMutex
	watcher    *watcher

	templateExecutor template.Executor
}

type subItem struct {
	uri   string
	isDir bool
}

type node struct {
	isNote           bool
	templateExecutor template.Executor
	absolutePath     string
	absoluteUri      string
	name             string
	parent           *node

	// dir node only
	subItems []*node
	index    string
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

	nr.uriNodeMap = make(map[string]*node)
	for _, dir := range config.GetSiteConfig().Template.StaticDirs {
		if err := nr.buildTree("/", filepath.Join(nr.templateRoot, dir), false, nil, nil); err != nil {
			return err
		}
	}
	if err := nr.watcher.addDirs(nr.templateRoot); err != nil {
		return err
	}
	if err := nr.buildTree("/", nr.noteRoot, true, config.GetSiteConfig().Note.NoteFileRegExp, nil); err != nil {
		return err
	}
	if err := nr.watcher.addDirs(nr.noteRoot); err != nil {
		return err
	}

	nr.watcher.watch(nr)

	return nil
}

func (nr notesRouter) Route(uri string) (content []byte, err error) {
	nr.lock.RLock()
	n, ok := nr.uriNodeMap[strings.ToLower(uri)]
	nr.lock.RUnlock()
	if !ok {
		return nr.templateExecutor.Get404(), os.ErrNotExist
	}
	return n.GetContent()
}

func (nr *notesRouter) buildTree(baseUri string, dir string, isNote bool, pattern *regexp.Regexp, parent *node) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	if parent == nil {
		parent = new(node)
		parent.isNote = isNote
		parent.templateExecutor = nr.templateExecutor
		parent.absolutePath = dir
		parent.absoluteUri = baseUri
		if conf, err := config.GetCategoryConfig(parent.absolutePath); err == nil && conf != nil {
			if conf.Index != "" {
				parent.index = conf.Index
			}
			if conf.NoteFileRegExp != nil {
				pattern = conf.NoteFileRegExp
			}
		}
		parent.subItems = make([]*node, 0)
		nr.uriNodeMap[parent.absoluteUri] = parent
	}
	for _, f := range files {
		self := new(node)
		self.isNote = isNote
		self.templateExecutor = nr.templateExecutor
		self.absolutePath = filepath.Join(dir, f.Name())
		self.name = f.Name()
		self.parent = parent
		if f.IsDir() {
			self.subItems = make([]*node, 0)
			subIsNote := isNote
			uriName := self.name
			if isNote {
				if conf, err := config.GetCategoryConfig(self.absolutePath); err == nil && conf != nil {
					subIsNote = true
					if conf.Name != "" {
						uriName = conf.Name
						self.name = conf.Name
					}
					if conf.DisplayName != "" {
						self.name = conf.DisplayName
					}
					if conf.Index != "" {
						self.index = conf.Index
					}
					if conf.NoteFileRegExp != nil {
						pattern = conf.NoteFileRegExp
					}
				} else if conf, err := config.GetResourceConfig(self.absolutePath); err == nil && conf != nil {
					subIsNote = false
					if conf.Name != "" {
						self.name = conf.Name
					}
				} else {
					continue
				}
			}
			self.absoluteUri = baseUri + strings.ToLower(uriName) + "/"
			if !(isNote && !subIsNote) {
				nr.uriNodeMap[self.absoluteUri] = self
				parent.subItems = append(parent.subItems, self)
			}
			if err := nr.buildTree(self.absoluteUri, self.absolutePath, subIsNote, pattern, self); err != nil {
				return err
			}
		} else {
			uriName := self.name
			if isNote {
				matches := pattern.FindAllStringSubmatch(f.Name(), -1)
				if matches == nil {
					continue
				}
				if len(matches) > 0 {
					if len(matches[0]) > 1 && len(matches[0][1]) > 0 {
						uriName = matches[0][1]
						if strings.HasSuffix(uriName, ".") {
							uriName = uriName[:len(uriName)-1]
							self.name = uriName
							uriName += "/"
						} else {
							self.name = uriName
						}
					}
					if len(matches[0]) > 2 && len(matches[0][2]) > 0 {
						self.name = matches[0][2]
					}
				}
				parent.subItems = append(parent.subItems, self)
			}
			self.absoluteUri = baseUri + strings.ToLower(uriName)
			nr.uriNodeMap[self.absoluteUri] = self
		}
	}
	return nil
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

func (n *node) toTemplateItem(parent *template.BasicItem, find *node) (itemForThis *template.BasicItem, itemForFind *template.BasicItem) {
	itemForThis = new(template.BasicItem)
	itemForThis.Parent = parent
	itemForThis.Uri = n.absoluteUri
	itemForThis.Name = n.name
	if n.subItems != nil {
		itemForThis.Children = make([]*template.BasicItem, 0)
		for _, c := range n.subItems {
			item, found := c.toTemplateItem(itemForThis, find)
			itemForThis.Children = append(itemForThis.Children, item)
			if found != nil {
				itemForFind = found
			}
		}
	}
	if n == find {
		util.Assert(itemForFind == nil, "data error")
		itemForFind = itemForThis
	}
	return
}

func (n *node) GetContent() ([]byte, error) {

	var pageData *template.PageData
	if n.isNote {
		root := n
		for root.parent != nil {
			root = root.parent
		}
		_, item := root.toTemplateItem(nil, n)
		for p := item; p != nil; p = p.Parent {
			p.IsAncestor = true
		}
		pageData = new(template.PageData)
		pageData.BasicItem = item
	}
	if n.subItems == nil {
		t := translator.New(n.absolutePath)
		content, err := t.Translate()
		if err != nil {
			if os.IsNotExist(err) {
				return n.templateExecutor.Get404(), err
			} else {
				return n.templateExecutor.Get500(), err
			}
		}
		if n.isNote {
			pageData.Content = string(content)
			return n.templateExecutor.GetContent(*pageData)
		} else {
			return content, nil
		}
	} else {
		util.Assert(n.isNote, "check code")
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
			pageData.Content = string(content)
		}
		if n.parent == nil {
			return n.templateExecutor.GetIndex(*pageData)
		} else {
			return n.templateExecutor.GetCategory(*pageData)
		}

	}

}
