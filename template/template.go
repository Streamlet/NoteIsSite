package template

import (
	"bytes"
	"io/ioutil"
	"sync"
	"text/template"
)

type IndexData struct {
	SubCategories []SubItem
	Contents      []SubItem
}

type CategoryData struct {
	Name string
	IndexData
}

type ContentData struct {
	Title   string
	Content string
}

type SubItem struct {
	Name string
	Uri  string
}

type Executor interface {
	Update(templateDir string) error

	GetIndex(data IndexData) ([]byte, error)
	GetCategory(data CategoryData) ([]byte, error)
	GetContent(data ContentData) ([]byte, error)

	Get404() []byte
	Get500() []byte
}

func NewExecutor(templateDir string) (Executor, error) {
	td := new(templateData)
	err := td.Update(templateDir)
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

func (td *templateData) Update(templateDir string) error {
	index, err := ioutil.ReadFile(templateDir + "/index.template.html")
	if err != nil {
		return err
	}
	category, err := ioutil.ReadFile(templateDir + "/category.template.html")
	if err != nil {
		return err
	}
	content, err := ioutil.ReadFile(templateDir + "/content.template.html")
	if err != nil {
		return err
	}
	err404, err := ioutil.ReadFile(templateDir + "/404.html")
	if err != nil {
		return err
	}
	err500, err := ioutil.ReadFile(templateDir + "/500.html")
	if err != nil {
		return err
	}

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
