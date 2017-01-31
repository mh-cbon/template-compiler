This is a template!

{{if ne 0 (len .MethodItems)}}
  <ul>
  {{range .MethodItems}}
    <li>{{.}}</li>
  {{end}}
  </ul>
{{else}}
No items!
{{end}}
