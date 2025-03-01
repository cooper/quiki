[[.JSON]]

<meta
[[if .Model]]
      data-nav="models"
      data-icon="cube"
[[else if .Config]]
      data-nav="settings"
      data-icon="cog"
[[else if .Category]]
    data-nav="categories"
    data-icon="list"
[[else]]
      data-nav="pages"
      data-icon="edit"
[[end]]
      data-title="[[or .Title .File]]"
      data-scripts="ace jquery editor"
      data-styles="editor colorpicker diff2html"
      data-flags="no-margin compact-sidebar"
/>

[[if not .Found]]
    Not found.
[[else]]
<div class="editor-toolbar-wrapper">
    <ul class="editor-toolbar">
        [[if and .Page .Info]][[if .Info.External]]
            <li class="readonly">READ ONLY</li>
        [[end]][[end]]
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
<div id="editor" data-file="[[.File]]">
[[- .Content -]]
</div>
[[end]]


<template id="tmpl-link-helper">
    <div id="editor-link-type-internal" class="editor-link-type editor-small-tab active" title="Page"><i class="fa fa-file-alt"></i></div>
    <div id="editor-link-type-external" class="editor-link-type editor-small-tab" title="External wiki page"><i class="fa fa-globe"></i></div>
    <div id="editor-link-type-category" class="editor-link-type editor-small-tab" title="Category"><i class="fa fa-list"></i></div>
    <div id="editor-link-type-url" class="editor-link-type editor-small-tab" title="External URL"><i class="fa fa-external-link"></i></div>
    <div style="clear: both;"></div>
    <div id="editor-link-wrapper">
    <span id="editor-link-title1">Page target</span><br />
    <input id="editor-link-target" class="editor-full-width-input" type="text" placeholder="My Page" />
    <span id="editor-link-title2">Display text</span><br />
    <input id="editor-link-display" class="editor-full-width-input" type="text" placeholder="Click here" /><br/>
    </div>
    <div id="editor-link-insert" class="editor-tool-large-button">Insert page link</div>
</template>

<template id="tmpl-save-helper">
    <div id="editor-save-wrapper">
    Edit summary<br />
    <input id="editor-save-message" class="editor-full-width-input" type="text" placeholder="Updated {%= o.file %}" />
    </div>
    <div id="editor-save-commit" class="editor-tool-large-button">Commit changes</div>
</template>

<template id="tmpl-save-spinner">
    <div style="text-align: center;"><i class="fa fa-spinner fa-3x fa-spin center"></i></div>
</template>

<template id="tmpl-delete-confirm">
    <div id="editor-delete-wrapper">
        <i class="fa fa-3x center fa-question-circle"></i>
    </div>
    <div id="editor-delete-button" class="editor-tool-large-button">Are you sure?</div>
</template>

<template id="tmpl-color-helper">
    <div id="editor-color-type-hex" class="editor-color-type editor-small-tab active" title="Hex color picker">
        <i class="fa fa-hashtag"></i>
        Color picker
    </div>
    <div id="editor-color-type-list" class="editor-color-type editor-small-tab" title="Color list">
        <i class="fa fa-paint-brush"></i>
        Color list
    </div>
    <div id="editor-color-names"></div>
    <div id="editor-color-hex"></div>
</template>

<template id="tmpl-color-container">
    <table>
        <tr>
            <td valign="top">
                <div id="colorpicker-color-map"></div>
            </td>
            <td valign="top">
                <div id="colorpicker-color-bar"></div>
            </td>
            <td valign="top">
                <table>
                    <tr>
                        <td colspan="3">
                            <div id="colorpicker-preview" style="background-color: #fff; width: 90px; height: 60px; padding: 0; margin: 0;">
                                <br />
                            </div>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <input type="radio" id="colorpicker-hue-radio" name="colorpicker-mode" value="0" />
                        </td>
                        <td>
                            <label for="colorpicker-hue-radio">H</label>
                        </td>
                        <td>
                            <input type="text" id="colorpicker-hue" value="0" style="width: 40px;" />
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <input type="radio" id="colorpicker-saturation-radio" name="colorpicker-mode" value="1" />
                        </td>
                        <td>
                            <label for="colorpicker-saturation-radio">S</label>
                        </td>
                        <td>
                            <input type="text" id="colorpicker-saturation" value="100" style="width: 40px;" />
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <input type="radio" id="colorpicker-brightness-radio" name="colorpicker-mode" value="2" />
                        </td>
                        <td>
                            <label for="colorpicker-brightness-radio">B</label>
                        </td>
                        <td>
                            <input type="text" id="colorpicker-brightness" value="100" style="width: 40px;" />
                        </td>
                    </tr>
                    <tr>
                        <td colspan="3" height="5">
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <input type="radio" id="colorpicker-red-radio" name="colorpicker-mode" value="r" />
                        </td>
                        <td>
                            <label for="colorpicker-red-radio">R</label>
                        </td>
                        <td>
                            <input type="text" id="colorpicker-red" value="255" style="width: 40px;" />
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <input type="radio" id="colorpicker-green-radio" name="colorpicker-mode" value="g" />
                        </td>
                        <td>
                            <label for="colorpicker-greenRadio">G</label>
                        </td>
                        <td>
                            <input type="text" id="colorpicker-green" value="0" style="width: 40px;" />
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <input type="radio" id="colorpicker-blue-radio" name="colorpicker-mode" value="b" />
                        </td>
                        <td>
                            <label for="colorpicker-blue-radio">B</label>
                        </td>
                        <td>
                            <input type="text" id="colorpicker-blue" value="0" style="width: 40px;" />
                        </td>
                    </tr>
                    <tr>
                        <td>
                            #
                        </td>
                        <td colspan="2">
                            <input type="text" id="colorpicker-hex" value="FF0000" style="width: 57px;" />
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</template>

<template id="tmpl-revision-viewer">
    <div id="editor-revisions"></div>
</template>

<template id="tmpl-revision-row">
    <b>{%= o.message.replace(/^Updated (.*?):/, '') %}</b><br />
    {%= o.author %}<br />
    {%= o.date %}
</template>

<template id="tmpl-revision-overlay">
    <div>
        <div class="editor-revision-diff-button" title="View the page at this point in history">
            <i class="fa fa-binoculars"></i>
            View on wiki
        </div>
        <div class="editor-revision-diff-button" title="View the page source at this point in history">
            <i class="fa fa-code"></i>
            View source
        </div>
    </div>
    <div>
        <div class="editor-revision-diff-button" title="Compare this revision to the current version">
            <i class="fa fa-circle-o"></i>
            Compare to current
        </div>
        <div class="editor-revision-diff-button" title="Compare this revision to the one before it">
            <i class="fa fa-clock-o"></i>
            Compare to previous
        </div>
    </div>
    <div>
        <div class="editor-revision-diff-button" title="Revert this change">
            <i class="fa fa-undo"></i>
            Revert
        </div>
        <div class="editor-revision-diff-button" title="Revert all changes since this version">
            <i class="fa fa-share"></i>
            Restore
        </div>
    </div>
    <div>
        <div class="editor-revision-diff-button" title="Return to the revision list">
            <i class="fa fa-chevron-left"></i>
            Back
        </div>
    </div>
</template>

<template id="tmpl-color-name">
    <span style="padding-left: 10px;">{%= o.colorName %}</span>
</template>

<template id="tmpl-page-options">
    <h3>Settings</h3>
    <table><tbody>
        <tr>
            <td class="left"><span title="Human-readable page title.">Title</span></td>
            <td><input class="title" type="text" value="{%= o.title %}" /></td>
        </tr>
        <tr>
            <td class="left"><span title="Page author/primary maintainer. Not used for revision tracking but may be displayed on the page.">Author</span></td>
            <td><input class="author" type="text" value="{%= o.author %}" /></td>
        </tr>
        <tr>
            <td class="left"><span title="Marks the page as a draft; unauthenticated users may not view it.">Draft</span></td>
            <td><input class="draft" type="checkbox"[["{%= o.draft ? ' checked' : '' %}"]] /></td>
        </tr>
    </tbody></table>
    <h3>Categories</h3>
    <table><tbody>
        <tr class="add-category"><td>
            <i class="fa fa-plus-circle" style="color: #00B545;"></i>
            <input type="text" placeholder="Add category" />
        </td></tr>
    </tbody></table>
</template>

<template id="tmpl-page-category">
    <td>
        <i class="fa fa-minus-circle" style="color: #EB2F42;"></i>
        <span>{%= o.catName %}</span>
    </td>
</template>

<template id="tmpl-model-options">
    <h3>Settings</h3>
    <table><tbody>
        <tr>
            <td class="left"><span title="Human-readable model title.">Title</span></td>
            <td><input class="title" type="text" value="{%= o.title %}" /></td>
        </tr>
        <tr>
            <td class="left"><span title="Model author/primary maintainer.">Author</span></td>
            <td><input class="author" type="text" value="{%= o.author %}" /></td>
        </tr>
    </tbody></table>
</template>