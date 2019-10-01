{{template "header.tpl" .}}
{{range $i, $p := .Pages}}
    {{with $alternate := odd $i }}
        <div class="main-wrapper alternate">
    {{else}}
        <div class="main-wrapper">
    {{end}}
        {{$p.HTMLContent}}
    </div>
{{end}}
{{template "footer.tpl" .}}