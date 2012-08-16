package gotemp

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"bytes"
	"net/http"
	"text/template"
)

func init() {
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/edit/modify", modifyEditHandler)
	http.HandleFunc("/edit/post", postEditHandler)
}

func forceAdmin(c appengine.Context, w http.ResponseWriter, r *http.Request) (isAdmin bool) {
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return false
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return false
	}
	if !user.IsAdmin(c) {
		w.WriteHeader(http.StatusForbidden)
		return false
	}

	return true
}

//      dD      d88888b d8888b. d888888b d888888b 
//     d8'      88'     88  `8D   `88'   `~~88~~' 
//    d8'       88ooooo 88   88    88       88    
//   d8'        88~~~~~ 88   88    88       88    
//  d8'         88.     88  .8D   .88.      88    
// C8'          Y88888P Y8888D' Y888888P    YP    

var edit_handler_template = template.New("/Edit")

func init() {
	var edit_handler_template_string string
	{
		buf := bytes.NewBuffer(nil)
		const form_body = `<form action="/edit/modify" method="post">
			<fieldset>
				<legend>Select A Template</legend>
				<select name="Name" class="field">
					<optgroup label="Templates">{{range .}}
						<option value="{{html .StringID}}">{{html .StringID}}</option>{{end}}
					</optgroup>
					<option value="New Template">New Template</option>
				</select>
			</fieldset>
			<div>
				<input name="Submit" id="Submit" type="hidden" />
				<button type="button" onClick="SubmitHandler(this.form)">Select</button>
				<button type="button" onClick="RemoveHandler(this.form)">Remove</button>
			</div>
		</form>`

		const confirm_script = `<script type="text/javascript">
function SubmitHandler(form) {
		$("#Submit").val("Select");
		form.submit();	
}
function RemoveHandler(form) {
	if (confirm("Are you sure you want to submit the form?")) {
		$("#Submit").val("Remove");
		form.submit();
	}
}
		</script>`

		form := prettyTemplateData{
			Title:   `Edit Templates`,
			Body:    []string{form_body},
			Scripts: []string{confirm_script},
		}

		if err := pretty_html_template.Execute(buf, form); err != nil {
			panic(err)
		}
		edit_handler_template_string = buf.String()
	}
	template.Must(edit_handler_template.Parse(edit_handler_template_string))
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if !forceAdmin(c, w, r) {
		return
	}

	keys, err := datastore.NewQuery("Name").KeysOnly().GetAll(c, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := edit_handler_template.Execute(w, keys); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//      dD      d88888b d8888b. d888888b d888888b           dD      .88b  d88.  .d88b.  d8888b. d888888b d88888b db    db 
//     d8'      88'     88  `8D   `88'   `~~88~~'          d8'      88'YbdP`88 .8P  Y8. 88  `8D   `88'   88'     `8b  d8' 
//    d8'       88ooooo 88   88    88       88            d8'       88  88  88 88    88 88   88    88    88ooo    `8bd8'  
//   d8'        88~~~~~ 88   88    88       88           d8'        88  88  88 88    88 88   88    88    88~~~      88    
//  d8'         88.     88  .8D   .88.      88          d8'         88  88  88 `8b  d8' 88  .8D   .88.   88         88    
// C8'          Y88888P Y8888D' Y888888P    YP         C8'          YP  YP  YP  `Y88P'  Y8888D' Y888888P YP         YP    
// Modify initialization
var modify_edit_handler_template = template.New("/Edit/Modify")

func init() {
	var input_adder_script string
	{
		var input_fields = []string{
			`<input type="text" class="field" name="Inputs" cols="80"/>`,
			`<input type="button" class="remove" value="-" onclick="$(this).parent().remove()"/>`,
		}

		buf := bytes.NewBuffer(nil)
		data := fieldAdderScriptData{
			FunctionName: "AddInputFields",
			Fields:       input_fields,
		}
		if err := field_adder_template.Execute(buf, data); err != nil {
			panic(err)
		}
		input_adder_script = buf.String()
	}

	var input_dependencies_adder_script string
	{
		var input_dependency_fields = []string{
			`<input type="text" class="field" name="InputDependencies" cols="80"/>`,
			`<input type="button" class="remove" value="-" onclick="$(this).parent().remove()"/>`,
		}

		buf := bytes.NewBuffer(nil)
		data := fieldAdderScriptData{
			FunctionName: "AddInputDependencyFields",
			Fields:       input_dependency_fields,
		}
		if err := field_adder_template.Execute(buf, data); err != nil {
			panic(err)
		}
		input_dependencies_adder_script = buf.String()
	}

	var dependencies_adder_script string
	{
		var dependency_fields = []string{
			`<input type="text" class="field" name="Dependencies" cols="80"/>`,
			`<input type="button" class="remove" value="-" onclick="$(this).parent().remove()"/>`,
		}

		buf := bytes.NewBuffer(nil)
		data := fieldAdderScriptData{
			FunctionName: "AddDependencyFields",
			Fields:       dependency_fields,
		}
		if err := field_adder_template.Execute(buf, data); err != nil {
			panic(err)
		}
		dependencies_adder_script = buf.String()
	}

	var modify_edit_handler_template_string string
	{
		buf := bytes.NewBuffer(nil)
		const form_body = `<form action="/edit/post" method="post">
				<fieldset>
					<legend>Name</legend>
					<input type="text" class="field" name="Name" cols="80" value="{{html .Name}}"/>
				</fieldset>
				<fieldset>
					<legend>Description</legend>
					<textarea name="Description" rows="20" cols="80">{{html .Description}}</textarea>
				</fieldset>
				<fieldset>
					<legend>Inputs</legend>
					<input type="button" value="Add an Input" class="add" onclick="AddInputFields($(this))" />{{range .Inputs}}
					<div class="fieldwrapper">
						<input type="text" class="field" name="Inputs" cols="80" value="{{html .}}"/>
						<input type="button" class="remove" value="-" onclick="$(this).parent().remove()"/>
					</div>{{end}}
				</fieldset>
				<fieldset>
					<legend>Input Dependencies</legend>
					<input type="button" value="Add a Input Dependency" class="add" onclick="AddInputDependencyFields($(this))" />{{range .InputDependencies}}
					<div class="fieldwrapper">
						<input type="text" class="field" name="InputDependencies" cols="80" value="{{html .}}"/>
						<input type="button" class="remove" value="-" onclick="$(this).parent().remove()"/>
					</div>{{end}}
				</fieldset>
				<fieldset>
					<legend>Dependencies</legend>
					<input type="button" value="Add a Dependency" class="add" onclick="AddDependencyFields($(this))" />{{range .Dependencies}}
					<div class="fieldwrapper">
						<input type="text" class="field" name="Dependencies" cols="80" value="{{html .}}"/>
						<input type="button" class="remove" value="-" onclick="$(this).parent().remove()"/>
					</div>{{end}}
				</fieldset>
				<fieldset>
					<legend>Data</legend>
					<textarea name="Data" rows="20" cols="80">{{html .Data}}</textarea>
				</fieldset>
				<div><input type="submit" value="Submit"></div>
			</form>`

		form := prettyTemplateData{
			Title: `Modify Template`,
			Scripts: []string{
				field_adder_script,
				input_adder_script,
				input_dependencies_adder_script,
				dependencies_adder_script,
			},
			Body: []string{form_body},
		}

		if err := pretty_html_template.Execute(buf, form); err != nil {
			panic(err)
		}
		modify_edit_handler_template_string = buf.String()
	}
	template.Must(modify_edit_handler_template.Parse(modify_edit_handler_template_string))
}
func modifyEditHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if !forceAdmin(c, w, r) {
		return
	}

	Name := r.FormValue("Name")

	switch r.FormValue("Submit") {
	case "Select":
		var template Template
		if Name != "New Template" {
			var err error
			if template, err = (contextTemplateLoader{c}).LoadTemplate(Name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := modify_edit_handler_template.Execute(w, template); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "Remove":
		if err := datastore.Delete(c, datastore.NewKey(c, "Name", Name, 0, nil)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "form parsing error", http.StatusInternalServerError)
		return
	}
}

func postEditHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	ctl := contextTemplateLoader{c}

	if !forceAdmin(c, w, r) {
		return
	}

	Name := r.FormValue("Name")
	Description := r.FormValue("Description")
	Data := r.FormValue("Data")
	Inputs := r.Form["Inputs"]
	InputDependencies := r.Form["InputDependencies"]
	Dependencies := r.Form["Dependencies"]

	if len(Name) == 0 {
		http.Error(w, "No Name Provided", http.StatusInternalServerError)
		return
	}
	if len(Data) == 0 {
		http.Error(w, "No Data Provided", http.StatusInternalServerError)
		return
	}

	template := Template{
		Name: Name, Data: Data, Description: Description,
		Inputs: Inputs, InputDependencies: InputDependencies, Dependencies: Dependencies,
	}

	if err := ctl.SaveTemplate(template); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
