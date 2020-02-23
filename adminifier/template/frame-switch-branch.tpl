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
</ul>

New Branch:
<form action="{{.Root}}/func/create-branch" method="post">
    <input type="text" name="branch" />
    <input type="submit" name="submit" value="Create" />
</form>