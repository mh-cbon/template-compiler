This is a template!

{{if ne 0 (len .Items)}}
  <ul>
  {{range .Items}}
    <li>{{.}}</li>
  {{end}}
  </ul>
{{else}}
No items!
{{end}}
