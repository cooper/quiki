<meta
    data-title="Switch Branch"
    data-icon="git-alt"
    data-icon-b="yes"
/>

<p>Branching is experimental.</p>

<p><b>Choose Branch</b></p>
<ul>
{{$root := .Root}}
{{range .Branches}}
    <li><a href="{{$root}}/func/switch-branch/{{.}}">{{.}}</a></li>
{{end}}
</ul>

<p><b>New Branch</b></p>
<form action="{{.Root}}/func/create-branch" method="post">
    <input type="text" name="branch" />
    <input type="submit" name="submit" value="Create" />
</form>