package template

import (
	"bytes"
	"github.com/Streamlet/NoteIsSite/config"
	"io/ioutil"
	"sync"
	"text/template"
)

type IndexData struct {
	HasSubItems
	HasContent
}

type CategoryData struct {
	Name string
	HasSubItems
	HasContent
	HasParentItems
}

type ContentData struct {
	Title string
	HasContent
	HasParentItems
}

type HasSubItems struct {
	SubItems []BasicItem
}

func (data HasSubItems) HasChildren() bool {
	return len(data.SubItems) > 0
}

func (data HasSubItems) Children() []BasicItem {
	return data.SubItems
}

type HasContent struct {
	Content string
}

type HasParentItems struct {
	// Parents[0] is the root items, that is, sub items of index
	// ...
	// Parents[len(Parent)-3] is the grand parent items
	// Parents[len(Parent)-2] is the parent items
	// Parents[len(Parent)-1] is the brother items
	Parents [][]ParentItem
}

func (data HasParentItems) Roots() []ParentItem {
	if len(data.Parents) > 0 {
		return data.Parents[0]
	} else {
		return nil
	}
}

func (data HasParentItems) Brothers() []ParentItem {
	if len(data.Parents) > 0 {
		return data.Parents[len(data.Parents)-1]
	} else {
		return nil
	}
}

func (data HasParentItems) Ancestors() []BasicItem {
	ancestors := make([]BasicItem, len(data.Parents))
	for i, l0 := range data.Parents {
		for _, l1 := range l0 {
			if l1.IsAncestor {
				ancestors[i] = l1.BasicItem
			}
		}
	}
	return ancestors
}

type BasicItem struct {
	Name  string
	Uri   string
	IsDir bool
}

type ParentItem struct {
	BasicItem
	IsAncestor bool
}

type Executor interface {
	Update(templateRoot string) error

	GetIndex(data IndexData) ([]byte, error)
	GetCategory(data CategoryData) ([]byte, error)
	GetContent(data ContentData) ([]byte, error)

	Get404() []byte
	Get500() []byte
}

func NewExecutor(templateRoot string) (Executor, error) {
	td := new(templateData)
	err := td.Update(templateRoot)
	if err != nil {
		return nil, err
	}
	return td, nil
}

type templateData struct {
	lock             sync.RWMutex
	indexTemplate    string
	categoryTemplate string
	contentTemplate  string
	err404           []byte
	err500           []byte
}

func (td *templateData) Update(templateRoot string) error {
	c := config.GetSiteConfig().Template
	index, err := ioutil.ReadFile(templateRoot + "/" + c.IndexTemplate)
	if err != nil {
		return err
	}
	category, err := ioutil.ReadFile(templateRoot + "/" + c.CategoryTemplate)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadFile(templateRoot + "/" + c.ContentTemplate)
	if err != nil {
		return err
	}
	err404, _ := ioutil.ReadFile(templateRoot + "/" + c.ErrorPage404)
	err500, _ := ioutil.ReadFile(templateRoot + "/" + c.ErrorPage500)

	defer td.lock.Unlock()
	td.lock.Lock()

	td.indexTemplate = string(index)
	td.categoryTemplate = string(category)
	td.contentTemplate = string(content)
	td.err404 = err404
	td.err500 = err500

	return nil
}

func (td templateData) GetIndex(data IndexData) ([]byte, error) {
	defer td.lock.RUnlock()
	td.lock.RLock()

	return td.execute(td.indexTemplate, data)
}

func (td templateData) GetCategory(data CategoryData) ([]byte, error) {
	defer td.lock.RUnlock()
	td.lock.RLock()

	return td.execute(td.categoryTemplate, data)
}

func (td templateData) GetContent(data ContentData) ([]byte, error) {
	defer td.lock.RUnlock()
	td.lock.RLock()

	return td.execute(td.contentTemplate, data)
}

func (td templateData) execute(tmpl string, data interface{}) ([]byte, error) {
	tt := template.New("")
	_, err := tt.Parse(tmpl)
	if err != nil {
		return td.err500, err
	}

	var buffer []byte
	w := bytes.NewBuffer(buffer)
	err = tt.Execute(w, data)
	if err != nil {
		return td.err500, err
	}
	return w.Bytes(), nil
}

func (td templateData) Get404() []byte {
	defer td.lock.RUnlock()
	td.lock.RLock()

	return td.err404
}

func (td templateData) Get500() []byte {
	defer td.lock.RUnlock()
	td.lock.RLock()

	return td.err500
}
