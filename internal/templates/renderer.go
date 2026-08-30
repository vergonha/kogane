package templates

import (
	"html/template"
	"net/http"
)

type Renderer struct {
	development bool
	glob        string
	tmpl        *template.Template
}

func New(glob string, development bool) (*Renderer, error) {
	renderer := &Renderer{
		development: development,
		glob:        glob,
	}

	if development {
		return renderer, nil
	}

	tmpl, err := template.ParseGlob(glob)
	if err != nil {
		return nil, err
	}

	renderer.tmpl = tmpl
	return renderer, nil
}

func (r *Renderer) Render(
	w http.ResponseWriter,
	name string,
	data any,
) error {
	if r.development {
		tmpl, err := template.ParseGlob(r.glob)
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(w, name, data)
	}

	return r.tmpl.ExecuteTemplate(w, name, data)
}
