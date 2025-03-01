<meta
    data-title="Switch branch"
    data-icon="git-alt"
    data-icon-b="yes"
/>

<h1>Switch Branch</h1>

<p>Branching is experimental.</p>

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