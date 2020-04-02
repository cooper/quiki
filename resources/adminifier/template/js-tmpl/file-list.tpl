<script type="text/x-tmpl" id="tmpl-filter-editor">
    <div class="filter-editor-title">
        <i class="fa fa-filter"></i> Filter
        <a href="#"><i class="fa fa-times"></i></a>
    </div>
</script>

<script type="text/x-tmpl" id="tmpl-filter-text">
    <span><input type="checkbox" /> {%= o.column %}</span>
    <div class="filter-row-inner">
        <form>
            <input type="radio" name="mode" data-mode="Contains" checked /> Contains
            <input type="radio" name="mode" data-mode="Is" /> Is<br />
        </form>
        <i class="fa fa-plus-circle fa-lg" style="color: chartreuse;"></i>
        <input type="text" />
    </div>
</script>

<script type="text/x-tmpl" id="tmpl-filter-date">
    <span><input type="checkbox" /> {%= o.column %}</span>
    <div class="filter-row-inner">
        <form>
            <input type="radio" name="mode" data-mode="Is" checked /> Is
            <input type="radio" name="mode" data-mode="Before" /> Before
            <input type="radio" name="mode" data-mode="After" /> After<br />
        </form>
        <i class="fa fa-plus-circle fa-lg" style="color: chartreuse;"></i>
        <input type="text" />
    </div>
</script>

<script type="text/x-tmpl" id="tmpl-filter-state">
    <span><input type="checkbox" /> {%= o.stateName %}</span>
</script>

<script type="text/x-tmpl" id="tmpl-filter-item">
    <i class="fa fa-minus-circle fa-lg" style="color: #FF7070;"></i>
    {%= o.mode %} &quot;{%= o.item %}&quot;
</script>
