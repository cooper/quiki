<meta
    data-title="Switch branch"
    data-icon="git-alt"
    data-icon-b="yes"
/>

<ul>
{{$root := .Root}}
{{range .Branches}}
    <li><a href="{{$root}}/func/switch-branch/{{.}}">{{.}}</a></li>
{{end}}
<ul>