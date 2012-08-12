package gotemp

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"net/http"
	"text/template"
)

func init() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/select", selectHandler)
	http.HandleFunc("/expand/", expandHandler)
}

func expandHandler(w http.ResponseWriter, r *http.Request) {
	const url_prefix = "/expand/"

	if r.URL.Path == url_prefix {
		rootHandler(w, r)
		return
	}

	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	path := r.URL.Path[len(url_prefix):]
	template, err := LoadTemplates(path, contextTemplateLoader{c})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.ParseForm()

	if err := template.Execute(w, r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var root_handler_template = template.New("/")

func init() {
	var root_handler_template_string string
	{
		buf := bytes.NewBuffer(nil)
		const form_body = `<form action="/select" method="post">
			<fieldset id="Template Select">
				<legend>Select A Template</legend>
				<select name="Name" class="field">{{range .}}
					<option value="{{html .StringID}}">{{html .StringID}}</option>{{end}}
				</select>
			</fieldset>
			<div>
				<input name="Submit" id="Submit" type="hidden">
				<button type="button" onClick="SubmitHandler(this.form)">Select</button>
			</div>
		</form>`

		const submit_script = `<script type="text/javascript">
function SubmitHandler(form) {
		$("#Submit").val("Select");
		form.submit();	
}
		</script>`

		form := prettyTemplateData{
			Title:   `Select Template`,
			Body:    []string{form_body},
			Scripts: []string{submit_script},
		}

		if err := pretty_html_template.Execute(buf, form); err != nil {
			panic(err)
		}
		root_handler_template_string = buf.String()
	}
	template.Must(root_handler_template.Parse(root_handler_template_string))
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	keys, err := datastore.NewQuery("Name").KeysOnly().GetAll(c, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := root_handler_template.Execute(w, keys); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Modify initialization
var select_handler_template = template.New("/Select")

func init() {
	var select_handler_template_string string
	{
		buf := bytes.NewBuffer(nil)
		form_body := `<form action="/expand/{{.Name}}" method="get">
			<fieldset id="Description">
				<legend>Description</legend>
				<p>{{.Description}}</p>
			</fieldset>{{range .Inputs}}
			<fieldset id="{{.}}">
				<legend>{{.}}</legend>
				<input type="button" value="+{{.}}" class="add" id="add" onclick="AddRemoveableFieldHelper($(this), '{{js .}}')" />
			</fieldset>{{end}}
			<div>
				<button type="submit">Expand</button>
			</div>
		</form>`

		field_adder_helper := `<script type="text/javascript">
function AddRemoveableFieldHelper(self, name) {
		AddRemoveableField(self.parent(), [
			"` + template.JSEscapeString(`<input type="text" class="field" name="`) + `" + name + "` + template.JSEscapeString(`" cols="80"/>`) + `",
			"` + template.JSEscapeString(`<input type="button" class="remove" value="-" onclick="$(this).parent().remove()"/>`) + `",
		]);
}
		</script>`

		form := prettyTemplateData{
			Title:   `Expand Template`,
			Scripts: []string{field_adder_script, field_adder_helper},
			Body:    []string{form_body},
		}

		if err := pretty_html_template.Execute(buf, form); err != nil {
			panic(err)
		}
		select_handler_template_string = buf.String()
	}
	template.Must(select_handler_template.Parse(select_handler_template_string))
}

func selectHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	Name := r.FormValue("Name")

	template, err := loadInputTemplate(Name, contextTemplateLoader{c})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := select_handler_template.Execute(w, template); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
