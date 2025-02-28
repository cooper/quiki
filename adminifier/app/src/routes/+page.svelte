<script lang="ts">
    import {Button} from 'svelte-ux'
    let count = 0;
    function handleClick() {
        count += 1;
    }
</script>

<h1>Welcome to adminifier</h1>

<Button on:click={handleClick} variant="fill">Click me {count}</Button>