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
{{range $n := .PageNumbers}}
    <a class="page-number{{if eq $.PageN $n}} active{{end}}" href="{{$.Root.Category}}/{{$.Name}}/{{$n}}">{{$n}}</a>
{{end}}
{{template "footer.tpl" .}}