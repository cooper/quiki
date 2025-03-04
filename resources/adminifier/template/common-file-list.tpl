<template id="tmpl-filter-editor">
    <div class="filter-editor-title">
        <i class="fa fa-filter"></i> Filter
        <a href="#"><i class="fa fa-times"></i></a>
    </div>
</template>

<template id="tmpl-filter-text">
    <span><input type="checkbox" /> {%= o.column %}</span>
    <div class="filter-row-inner">
        <form>
            <input type="radio" name="mode" data-mode="Contains" checked /> Contains
            <input type="radio" name="mode" data-mode="Matches" /> Matches
            <input type="radio" name="mode" data-mode="Is" /> Is<br />
        </form>
        <i class="fa fa-plus-circle fa-lg" style="color: chartreuse;"></i>
        <input type="text" />
    </div>
</template>

<template id="tmpl-filter-date">
    <span><input type="checkbox" /> {%= o.column %}</span>
    <div class="filter-row-inner">
        <form>
            <input type="radio" name="mode" data-mode="Before" checked /> Before
            <input type="radio" name="mode" data-mode="After" /> After
            <input type="radio" name="mode" data-mode="Is" /> Is<br />
        </form>
        <i class="fa fa-plus-circle fa-lg" style="color: chartreuse;"></i>
        <input type="text" />
    </div>
</template>

<template id="tmpl-filter-state">
    <span><input type="checkbox" /> {%= o.stateName %}</span>
</template>

<template id="tmpl-filter-item">
    <i class="fa fa-minus-circle fa-lg" style="color: #FF7070;"></i>
    {%= o.mode %} &quot;{%= o.item %}&quot;
</template>

<template id="tmpl-create-folder">
    <form action="{{.Root}}/func/create-{%= o.mode %}-folder" method="post">
        <label for="name">Folder Name:</label>
        <input type="text" name="name" />
        <input type="hidden" name="dir" value="{{.Cd}}" />
        <input type="submit" value="Create" />
    </form>
</template>