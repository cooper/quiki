<script type="text/x-tmpl" id="tmpl-image-grid-item">
    <a href="functions/image.php?file={%= o.file %}">
        <img alt="{%= o.file %}" src="functions/image.php?file={%= o.file %}&{%= o.dimension %}={%= o.dimValue %}" />
        <span>{%= o.file %}</span>
    </a>
</script>
