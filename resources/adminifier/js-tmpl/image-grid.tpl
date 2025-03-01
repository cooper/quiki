<template id="tmpl-image-grid-item">
    <a href="func/image/{%= o.file %}">
        <img alt="{%= o.file %}" src="func/image/{%= o.file %}?{%= o.dimension %}={%= o.dimValue %}" />
        <span>{%= o.file %}</span>
    </a>
</template>
