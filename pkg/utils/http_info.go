package utils

import (
	"github.com/eirsyl/apollo/pkg"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

type context struct {
	Version   string
	BuildDate string
	BuildInfo map[string]string
}

// BuildInformationHandler returns a http site containing information about
// the apollo binary.
func BuildInformationHandler(info map[string]string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		tmpl, err := template.New("buildInfo").Parse(`
		<h2>apollo</h2>
		<p>{{.Version}} <i>{{.BuildDate}}</i></p>
		<table style="border: 1px solid;">
		<tr>
		<th>Name</th>
		<th>Value</th>
		</tr>

        {{ range $key, $value := .BuildInfo  }}
		<tr>
		<td>{{ $key  }}</td>
		<td>{{ $value }}</td>
		</tr>
		{{ end }}

		</table>
		`)

		if err != nil {
			log.Warnf("Could not parse buildInfo template: %v", err)
			return
		}

		c := context{
			Version:   pkg.Version,
			BuildDate: pkg.BuildDate,
			BuildInfo: info,
		}

		err = tmpl.Execute(w, c)
		if err != nil {
			log.Warnf("Could not render build info page: %v", err)
		}
	}
}
