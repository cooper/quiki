<script lang="ts">
	import { Button } from "svelte-ux";
    let count = 0;
</script>

<h1 class="text-3xl font-bold underline">
    Hello world!
</h1>

<Button on:click={() => count += 1} variant="fill-outline">Click me {count}</Button>