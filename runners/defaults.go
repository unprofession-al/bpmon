package runners

import "text/template"

func Defaults() Runners {
	return Runners{
		"default": Runner{
			Description: `Print all check results in a short format and human readable`,
			ForEach:     true,
			Template: template.Must(template.New("default").Parse(`{{- range $index, $bp := .BP }}  {{ $bp.Status.Colorize $bp.Name }} {{ $bp.Status.Colorize "is" }} {{ $bp.Status.Colorize $bp.Status.String }}
  {{- range $index, $kpi := .Children }}
    {{ $kpi.Status.Colorize $kpi.Name }} {{ $kpi.Status.Colorize "is" }} {{ $kpi.Status.Colorize $kpi.Status.String }}
    {{- range $index, $svc := .Children }}
      {{ $svc.Status.Colorize $svc.Name }} {{ $svc.Status.Colorize "is" }} {{ $svc.Status.Colorize $svc.Status.String }}
    {{- end -}}
  {{ end }}
{{ end }}
`)),
		},
		"verbose": Runner{
			Description: `Print all check results in a long format and human readable`,
			ForEach:     true,
			Template: template.Must(template.New("verbose").Parse(`{{- range $index, $bp := .BP }}  {{ $bp.Status.Colorize $bp.Name }} {{ $bp.Status.Colorize "is" }} {{ $bp.Status.Colorize $bp.Status.String }}
            since: {{ $bp.Start.Format "2006-01-02 15:04:05" }}
      responsible: {{ $bp.Responsible }}
           values: {{ range $key, $val := $bp.Vals }}{{$key}}={{$val}} {{ end }}
  {{- range $index, $kpi := .Children }}
    {{ $kpi.Status.Colorize $kpi.Name }} {{ $kpi.Status.Colorize "is" }} {{ $kpi.Status.Colorize $kpi.Status.String }}
              since: {{ $kpi.Start.Format "2006-01-02 15:04:05" }}
        responsible: {{ $kpi.Responsible }}
    {{- range $index, $svc := .Children }}
      {{ $svc.Status.Colorize $svc.Name }} {{ $svc.Status.Colorize "is" }} {{ $svc.Status.Colorize $svc.Status.String }}
                since: {{ $svc.Start.Format "2006-01-02 15:04:05" }}
          responsible: {{ $svc.Responsible }}
               values: {{ range $key, $val := $svc.Vals }}{{$key}}={{$val}} {{ end }}
    {{- end -}}
  {{ end }}
{{ end }}
`)),
		},
		"issues": Runner{
			Description: `Print failed check results in a short format and human readable`,
			Template: template.Must(template.New("issues").Parse(`
{{- range $index, $bp := .BP -}}
  {{- if and (index $bp.Vals "in_availability") (ne $bp.Status 0) }}
  {{ $bp.Status.Colorize $bp.Name }} {{ $bp.Status.Colorize "is" }} {{ $bp.Status.Colorize $bp.Status.String }}
    {{- range $index, $kpi := .Children -}}
      {{- if ne $kpi.Status 0 }}
    {{ $kpi.Status.Colorize $kpi.Name }} {{ $kpi.Status.Colorize "is" }} {{ $kpi.Status.Colorize $kpi.Status.String }}
        {{- range $index, $svc := .Children -}}
          {{- if ne $svc.Status 0 }}
      {{ $svc.Status.Colorize $svc.Name }} {{ $svc.Status.Colorize "is" }} {{ $svc.Status.Colorize $svc.Status.String }}
          {{- end -}}
        {{ end }}
      {{- end -}}
    {{ end }}
  {{ end }}
{{- end -}}`)),
		},
		"issues_verbose": Runner{
			Description: `Print failed check results in a long format and human readable`,
			Template: template.Must(template.New("issues_verbose").Parse(`
{{- range $index, $bp := .BP -}}
  {{- if and (index $bp.Vals "in_availability") (ne $bp.Status 0) }}
  {{ $bp.Status.Colorize $bp.Name }} {{ $bp.Status.Colorize "is" }} {{ $bp.Status.Colorize $bp.Status.String }}
            since: {{ $bp.Start.Format "2006-01-02 15:04:05" }}
      responsible: {{ $bp.Responsible }}
           values: {{ range $key, $val := $bp.Vals }}{{$key}}={{$val}} {{ end }}
    {{- range $index, $kpi := .Children -}}
      {{- if ne $kpi.Status 0 }}
    {{ $kpi.Status.Colorize $kpi.Name }} {{ $kpi.Status.Colorize "is" }} {{ $kpi.Status.Colorize $kpi.Status.String }}
              since: {{ $kpi.Start.Format "2006-01-02 15:04:05" }}
        responsible: {{ $kpi.Responsible }}
        {{- range $index, $svc := .Children -}}
          {{- if ne $svc.Status 0 }}
      {{ $svc.Status.Colorize $svc.Name }} {{ $svc.Status.Colorize "is" }} {{ $svc.Status.Colorize $svc.Status.String }}
                since: {{ $svc.Start.Format "2006-01-02 15:04:05" }}
          responsible: {{ $svc.Responsible }}
               values: {{ range $key, $val := $svc.Vals }}{{$key}}={{$val}} {{ end }}
              message: {{ $svc.Output }}
                error: {{ $svc.Err }}
          {{- end -}}
        {{ end }}
      {{- end -}}
    {{ end }}
  {{ end }}
{{- end -}}`)),
		},
	}
}
