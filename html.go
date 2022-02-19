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

type Html struct {
	// response writer
	writer http.ResponseWriter

	// template engine
	engine *HtmlEngine
}

// create new html response
func newHtml(rw http.ResponseWriter, e *HtmlEngine) *Html {
	return &Html{
		writer: rw,
		engine: e,
	}
}

// Success is shortcut for Response with StatusOK = 200,
// use escape = false if don't need html escape (default `true`)
func (h *Html) Success(filename string, data Map, escape ...bool) error {
	return h.Response(http.StatusOK, filename, data, escape...)
}

// Error is shortcut for Response with StatusInternalServerError = 500,
// use escape = false if don't need html escape (default `true`)
func (h *Html) Error(filename string, data Map, escape ...bool) error {
	return h.Response(http.StatusInternalServerError, filename, data, escape...)
}

// Timeout is shortcut for Response with StatusGatewayTimeout = 504,
// use escape = false if don't need html escape (default `true`)
func (h *Html) Timeout(filename string, data Map, escape ...bool) error {
	return h.Response(http.StatusGatewayTimeout, filename, data, escape...)
}

// Forbidden is shortcut for Response with StatusForbidden = 403,
// use escape = false if don't need html escape (default `true`)
func (h *Html) Forbidden(filename string, data Map, escape ...bool) error {
	return h.Response(http.StatusForbidden, filename, data, escape...)
}

// NotFound is shortcut for Response with StatusNotFound = 404,
// use escape = false if don't need html escape (default `true`)
func (h *Html) NotFound(filename string, data Map, escape ...bool) error {
	return h.Response(http.StatusNotFound, filename, data, escape...)
}

// Unauthorized is shortcut for Response with StatusUnauthorized = 401,
// use escape = false if don't need html escape (default `true`)
func (h *Html) Unauthorized(filename string, data Map, escape ...bool) error {
	return h.Response(http.StatusUnauthorized, filename, data, escape...)
}

// Response render html from filename, output to browser,
// use escape = false if don't need html escape (default `true`)
func (h *Html) Response(statusCode int, filename string, data Map, escape ...bool) error {
	return h.engine.Response(h.writer, statusCode, filename, data, len(escape) == 0 || escape[0])
}

// Render render to io.Writer without output to browser,
// example:
//
//  var b bytes.Buffer
//  err = res.Html.Output(&b, ...)
//  fmt.Println(b.String())
func (h *Html) Render(out io.Writer, filename string, data Map, escape ...bool) error {
	return h.engine.Render(out, filename, data, len(escape) == 0 || escape[0])
}

// ResponseString response to browser from statuscode and string content
func (h *Html) ResponseString(statusCode int, text string) error {
	h.writer.WriteHeader(statusCode)
	_, err := h.writer.Write([]byte(text))
	return err
}

// ResponseByte render to browser from statuscode and byte content
func (h *Html) ResponseByte(statusCode int, text []byte) error {
	h.writer.WriteHeader(statusCode)
	_, err := h.writer.Write(text)
	return err
}

// StatusText response output to browser, with header and default string from statuscode
func (h *Html) StatusText(statusCode int) error {
	return h.ResponseString(statusCode, http.StatusText(statusCode))
}

//
// The following source code was copied from the goview package, we cannot use
// the package directly because goview only supports escaped html.
//
// Source: https://github.com/foolin/goview/blob/master/view.go
//
// engine used to render html or text output
type HtmlEngine struct {
	template   *Template
	tplMapHtml map[string]*_html.Template
	tplMapText map[string]*_text.Template
	tplMutex   sync.RWMutex
}

// create new engine from template config
func newTemplateEngine(template *Template) *HtmlEngine {
	return &HtmlEngine{
		template:   template,
		tplMapHtml: make(map[string]*_html.Template),
		tplMapText: make(map[string]*_text.Template),
		tplMutex:   sync.RWMutex{},
	}
}

// Response render output to responseWriter
func (e *HtmlEngine) Response(w http.ResponseWriter, statusCode int, name string, data Map, escape bool) error {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}
	w.WriteHeader(statusCode)
	return e.executeRender(w, name, data, escape)
}

// render output to io.Writer, so we can use that output before render to browser
func (e *HtmlEngine) Render(w io.Writer, name string, data Map, escape bool) error {
	return e.executeRender(w, name, data, escape)
}

// shortcut to render html
func (e *HtmlEngine) executeRender(out io.Writer, name string, data Map, escape bool) error {
	useMaster := true
	if filepath.Ext(name) == ".html" {
		useMaster = false
		name = strings.TrimSuffix(name, ".html")
	}

	if escape {
		return e.htmlEscape(out, name, data, useMaster)
	}
	return e.htmlString(out, name, data, useMaster)
}

// execute html template with escaped
func (e *HtmlEngine) htmlEscape(out io.Writer, name string, data Map, useMaster bool) error {
	var tpl *_html.Template
	var err error
	var ok bool

	allFuncs := make(_html.FuncMap, 0)
	allFuncs["include"] = func(layout string) (_html.HTML, error) {
		buf := new(bytes.Buffer)
		err := e.htmlEscape(buf, layout, data, false)
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
func (e *HtmlEngine) htmlString(out io.Writer, name string, data Map, useMaster bool) error {
	var tpl *_text.Template
	var err error
	var ok bool

	allFuncs := make(_text.FuncMap, 0)
	allFuncs["include"] = func(layout string) (string, error) {
		buf := new(bytes.Buffer)
		err := e.htmlString(buf, layout, data, false)
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
func (e *HtmlEngine) fileHandler(config *Template, tplFile string) (content string, err error) {
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
