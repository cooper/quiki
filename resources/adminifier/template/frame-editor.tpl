{{.JSON}}

<meta
{{if .Model}}
      data-nav="models"
      data-icon="cube"
{{else if .Config}}
      data-nav="settings"
      data-icon="cog"
{{else if .Category}}
    data-nav="categories"
    data-icon="list"
{{else}}
      data-nav="pages"
      data-icon="edit"
{{end}}
      data-title="{{or .Title .File}}"
      data-scripts="ace jquery editor"
      data-styles="editor colorpicker diff2html"
      data-flags="no-margin compact-sidebar"
/>

{{if not .Found}}
    Not found.
{{else}}
<div class="editor-toolbar-wrapper">
    <ul class="editor-toolbar">
        {{if and .Page .Info}}{{if .Info.External}}
            <li class="readonly">READ ONLY</li>
        {{end}}{{end}}
        <li data-action="save" class="right"><i class="fa right fa-save"></i> <span>Save</span></li>
        <li data-action="delete" class="right"><i class="fa right fa-trash"></i> Delete</li>
        <li data-action="revisions" class="right"><i class="fa right fa-history"></i> Revisions</li>
        <li data-action="view" class="right"><i class="fa right fa-binoculars"></i> View</li>
        <li class="hidden right" data-action="options"><i class="fa right fa-wrench"></i> Options</li>
        <li id="toolbar-redo" data-action="redo" class="right disabled"><i class="fa right fa-redo"></i> Redo</li>
        <li id="toolbar-undo" data-action="undo" class="right disabled"><i class="fa right fa-undo"></i> Undo</li>

        <!--
            <li style="float: right;"><i class="fa fa-paste"></i> Paste</li>
            <li style="float: right;"><i class="fa fa-copy"></i> Copy</li>
            <li style="float: right;"><i class="fa fa-cut"></i> Cut</li>
        -->

        <li class="hidden" data-action="bold"><i class="fa fa-bold"></i> Bold</li>
        <li class="hidden" data-action="italic"><i class="fa fa-italic"></i> Italic</li>
        <li class="hidden" data-action="underline"><i class="fa fa-underline"></i> Underline</li>
        <li class="hidden" data-action="strike"><i class="fa fa-strikethrough"></i> Strike</li>
        <li class="hidden" data-action="font"><i class="fa fa-font"></i> Color</li>
        <!--<li data-action="header"><i class="fa fa-header"></i> Header</li>-->
        <li class="hidden" data-action="image"><i class="fa fa-image"></i> Image</li>
        <li class="hidden" data-action="link"><i class="fa fa-link"></i> Link</li>
        <li class="hidden" data-action="emoji"><i class="fa fa-smile"></i> Emoticons</li>
        <li class="hidden" data-action="file"><i class="fa fa-paperclip"></i> File</li>
        <li class="hidden" data-action="list"><i class="fa fa-list-ul"></i> List</li>
        <li class="hidden" data-action="list"><i class="fa fa-list-ol"></i> Num list</li>
        <li class="hidden" data-action="model"><i class="fa fa-cube"></i> Model</li>
        <li class="hidden" data-action="infobox"><i class="fa fa-info-circle"></i> Infobox</li>
        <li class="hidden" data-action="code"><i class="fa fa-code"></i> Code</li>
        <li class="hidden" data-action="cite"><i class="fa fa-copyright"></i> Citation</li>

    </ul>
</div>
<div id="editor" data-file="{{.File}}">
{{- .Content -}}
</div>
{{end}}