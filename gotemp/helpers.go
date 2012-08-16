package gotemp

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"log"
	"text/template"
	"time"
)

var pretty_html_template = template.Must(template.New("Pretty").Parse(`<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
		<script type="text/javascript" src="/js/jquery-1.7.2.min.js"></script>{{range .Scripts}}
		{{.}}{{end}}
		<link rel="stylesheet" type="text/css" href="/css/pretty.css"/>
	</head>
	<body>{{range .Body}}
		{{.}}{{end}}
	</body>
</html>`,
))

type prettyTemplateData struct {
	Title         string
	Body, Scripts []string
}

type contextTemplateLoader struct {
	appengine.Context
}

const (
	Kibi = 1 << ((1 + iota) * 10)
	Mebi
	Gibi
)

type ctlTemplate struct {
	Name                                    string
	Description, Data                       []byte
	Inputs, InputDependencies, Dependencies []string
}

func toCtlTemplate(t Template) ctlTemplate {
	return ctlTemplate{
		Name:              t.Name,
		Description:       []byte(t.Description),
		Data:              []byte(t.Data),
		Inputs:            t.Inputs,
		InputDependencies: t.InputDependencies,
		Dependencies:      t.Dependencies,
	}
}
func fromCtlTemplate(t ctlTemplate) Template {
	return Template{
		Name:              t.Name,
		Description:       string(t.Description),
		Data:              string(t.Data),
		Inputs:            t.Inputs,
		InputDependencies: t.InputDependencies,
		Dependencies:      t.Dependencies,
	}
}

func (ctl contextTemplateLoader) SaveTemplate(t Template) error {
	c := ctl.Context

	if _, err := template.New(t.Name).Funcs(builtins).Parse(t.Data); err != nil {
		return err
	}

	c_temp := toCtlTemplate(t)

	if _, err := datastore.Put(c, datastore.NewKey(c, "Name", c_temp.Name, 0, nil), &c_temp); err != nil {
		return err
	}

	if err := memcache.Delete(c, c_temp.Name); err != memcache.ErrCacheMiss {
		log.Println(err)
	}

	return nil
}

func (ctl contextTemplateLoader) LoadTemplate(name string) (Template, error) {
	c := ctl.Context

	var c_temp ctlTemplate

	// check the memcache
	if _, err := memcache.Gob.Get(c, name, &c_temp); err == nil {
		return fromCtlTemplate(c_temp), nil
	} else if err != memcache.ErrCacheMiss {
		return Template{}, err // if encounter any error but cache miss return the failure
	}

	//  read the value from the datastore
	if err := datastore.Get(c, datastore.NewKey(c, "Name", name, 0, nil), &c_temp); err != nil {
		return Template{}, err
	}

	const max_cache_size = 20 * Mebi
	const min_cache_time = 20 // in seconds

	// if the oldest item in the cache is too young, 
	if stats, err := memcache.Stats(c); err != nil {
		log.Println(err)
	} else if stats == nil || stats.Bytes < max_cache_size || stats.Oldest > min_cache_time {
		// cache the value from the datastore if we're under the max size or the oldest item is over the min time
		err := memcache.Gob.Set(c, &memcache.Item{Key: name, Object: &c_temp, Expiration: 1 * time.Hour})
		log.Println(err)
	}
	return fromCtlTemplate(c_temp), nil
}

type fieldAdderScriptData struct {
	FunctionName string
	Fields       []string
}

const field_adder_script = `<script type="text/javascript">
function AddRemoveableField(parent, fields) {
	var fieldWrapper = $("<div class=\"fieldwrapper\" />");
	for(i in fields) {
		fieldWrapper.append($(fields[i]));
	}
	parent.append(fieldWrapper);
}
		</script>`

var field_adder_template = template.Must(template.New("FieldAdder").Parse(`<script type="text/javascript">
function {{.FunctionName}}(self) {
	AddRemoveableField(self.parent(), [{{range .Fields}}
		"{{js .}}",{{end}}
	]);
}
		</script>`))
