<script lang="ts">
	import { page } from "$app/state";
	import { mdiBookOpenVariant, mdiHome } from "@mdi/js";
	import {
		AppBar,
		AppLayout,
		Button,
		Icon,
		NavItem,
		settings,
		ThemeSelect,
		Tooltip
	} from "svelte-ux";
	import "../app.css";
  
	let { children } = $props();

	settings({
	  components: {
		AppBar: {
		  classes: "bg-neutral-400 text-white shadow-md text-primary-content p-3",
		},
		AppLayout: {
		  classes: {
			nav: "bg-neutral-800 py-2",
		  },
		},
		NavItem: {
		  classes: {
			root: "text-sm text-gray-400 pl-6 py-2 hover:text-white hover:bg-gray-300/10 [&:where(.is-active)]:text-sky-400 [&:where(.is-active)]:bg-gray-500/10",
		  },
		},
	  },
	});
  </script>
  
  <AppLayout navWidth={200}>
   
	<svelte:fragment slot="nav">
		<!-- Nav menu -->
		<NavItem
			path="/"
			text="Home"
			icon={mdiHome}
			currentUrl={page.url}
		/>
	</svelte:fragment>	

	<AppBar title="Svelte-ux-App" class="">
	  <div slot="actions" class="flex gap-3 p-4">
		<!-- App actions -->
		<Tooltip title="Svelte-ux documentation" placement="bottom" offset={2}>
		  <Button
			class="p-2"
			target="_blank"
			href="https://svelte-ux.techniq.dev/"
		  >
			<Icon data={mdiBookOpenVariant} size="2em" />
		  </Button>
		</Tooltip>  
		<Tooltip title="Color mode" placement="bottom" offset={2}>
		  <ThemeSelect />
		</Tooltip>
	  </div>
	</AppBar>
	<main>
	  <div class="flex justify-center max-w-lg flex-col mx-auto p-4">
		{@render children()}
	  </div>
	</main>
  </AppLayout>