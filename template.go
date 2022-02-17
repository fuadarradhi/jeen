// The following source code was copied from the goview package, we cannot use
// the package directly because goview only supports escaped html.
//
// Source: https://github.com/foolin/goview/blob/master/view.go
//
package jeen

import (
	"bytes"
	"fmt"
	_html "html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	_text "text/template"
)

// default html content-type
var HTMLContentType = []string{"text/html; charset=utf-8"}

// engine used to render html or text output
type TemplateEngine struct {
	template   *Template
	tplMapHtml map[string]*_html.Template
	tplMapText map[string]*_text.Template
	tplMutex   sync.RWMutex
}

// template is config of template
type Template struct {
	Root         string
	Master       string
	Partials     []string
	Funcs        Map
	DisableCache bool
	Delims       *Delims
}

// delimeter used in html or text template
type Delims struct {
	Left  string
	Right string
}

// create new template engine
func NewTemplateEngine(template *Template) *TemplateEngine {
	return &TemplateEngine{
		template:   template,
		tplMapHtml: make(map[string]*_html.Template),
		tplMapText: make(map[string]*_text.Template),
		tplMutex:   sync.RWMutex{},
	}
}

// render output to responseWriter
func (e *TemplateEngine) Render(w http.ResponseWriter, statusCode int, name string, data interface{}, escape bool) error {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = HTMLContentType
	}
	w.WriteHeader(statusCode)
	return e.executeRender(w, name, data, escape)
}

// render output to io.Writer, so we can use that output before render to browser
func (e *TemplateEngine) RenderWriter(w io.Writer, name string, data interface{}, escape bool) error {
	return e.executeRender(w, name, data, escape)
}

// shortcut to render html
func (e *TemplateEngine) executeRender(out io.Writer, name string, data interface{}, escape bool) error {
	useMaster := true
	if filepath.Ext(name) == ".html" {
		useMaster = false
		name = strings.TrimSuffix(name, ".html")
	}

	if escape {
		return e.executeHtmlTemplate(out, name, data, useMaster)
	}
	return e.executeTextTemplate(out, name, data, useMaster)
}

// execute html template with escaped
func (e *TemplateEngine) executeHtmlTemplate(out io.Writer, name string, data interface{}, useMaster bool) error {
	var tpl *_html.Template
	var err error
	var ok bool

	allFuncs := make(_html.FuncMap, 0)
	allFuncs["include"] = func(layout string) (_html.HTML, error) {
		buf := new(bytes.Buffer)
		err := e.executeHtmlTemplate(buf, layout, data, false)
		return _html.HTML(buf.String()), err
	}

	// Get the plugin collection
	for k, v := range e.template.Funcs {
		allFuncs[k] = v
	}

	e.tplMutex.RLock()
	tpl, ok = e.tplMapHtml[name]
	e.tplMutex.RUnlock()

	exeName := name
	if useMaster && e.template.Master != "" {
		exeName = e.template.Master
	}

	if !ok || e.template.DisableCache {
		tplList := make([]string, 0)
		if useMaster {
			//render()
			if e.template.Master != "" {
				tplList = append(tplList, e.template.Master)
			}
		}
		tplList = append(tplList, name)
		tplList = append(tplList, e.template.Partials...)

		// Loop through each template and test the full path
		tpl = _html.New(name).Funcs(allFuncs).Delims(e.template.Delims.Left, e.template.Delims.Right)
		for _, v := range tplList {
			var data string
			data, err = e.fileHandler(e.template, v)
			if err != nil {
				return err
			}
			var tmpl *_html.Template
			if v == name {
				tmpl = tpl
			} else {
				tmpl = tpl.New(v)
			}
			_, err = tmpl.Parse(data)
			if err != nil {
				return fmt.Errorf("ViewEngine render parser name:%v, error: %v", v, err)
			}
		}
		e.tplMutex.Lock()
		e.tplMapHtml[name] = tpl
		e.tplMutex.Unlock()
	}

	// Display the content to the screen
	err = tpl.Funcs(allFuncs).ExecuteTemplate(out, exeName, data)
	if err != nil {
		return fmt.Errorf("ViewEngine execute template error: %v", err)
	}

	return nil
}

// execute html template without escaped
func (e *TemplateEngine) executeTextTemplate(out io.Writer, name string, data interface{}, useMaster bool) error {
	var tpl *_text.Template
	var err error
	var ok bool

	allFuncs := make(_text.FuncMap, 0)
	allFuncs["include"] = func(layout string) (string, error) {
		buf := new(bytes.Buffer)
		err := e.executeHtmlTemplate(buf, layout, data, false)
		return buf.String(), err
	}

	// Get the plugin collection
	for k, v := range e.template.Funcs {
		allFuncs[k] = v
	}

	e.tplMutex.RLock()
	tpl, ok = e.tplMapText[name]
	e.tplMutex.RUnlock()

	exeName := name
	if useMaster && e.template.Master != "" {
		exeName = e.template.Master
	}

	if !ok || e.template.DisableCache {
		tplList := make([]string, 0)
		if useMaster {
			//render()
			if e.template.Master != "" {
				tplList = append(tplList, e.template.Master)
			}
		}
		tplList = append(tplList, name)
		tplList = append(tplList, e.template.Partials...)

		// Loop through each template and test the full path
		tpl = _text.New(name).Funcs(allFuncs).Delims(e.template.Delims.Left, e.template.Delims.Right)
		for _, v := range tplList {
			var data string
			data, err = e.fileHandler(e.template, v)
			if err != nil {
				return err
			}
			var tmpl *_text.Template
			if v == name {
				tmpl = tpl
			} else {
				tmpl = tpl.New(v)
			}
			_, err = tmpl.Parse(data)
			if err != nil {
				return fmt.Errorf("ViewEngine render parser name:%v, error: %v", v, err)
			}
		}
		e.tplMutex.Lock()
		e.tplMapText[name] = tpl
		e.tplMutex.Unlock()
	}

	// Display the content to the screen
	err = tpl.Funcs(allFuncs).ExecuteTemplate(out, exeName, data)
	if err != nil {
		return fmt.Errorf("ViewEngine execute template error: %v", err)
	}

	return nil
}

// filehandler
func (e *TemplateEngine) fileHandler(config *Template, tplFile string) (content string, err error) {
	// Get the absolute path of the root template
	path, err := filepath.Abs(config.Root + string(os.PathSeparator) + tplFile + ".html")
	if err != nil {
		return "", fmt.Errorf("ViewEngine path:%v error: %v", path, err)
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("ViewEngine render read name:%v, path:%v, error: %v", tplFile, path, err)
	}
	return string(data), nil
}
