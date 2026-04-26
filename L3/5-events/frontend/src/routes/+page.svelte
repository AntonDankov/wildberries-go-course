<script lang="ts">
	import { User, ShieldCheck } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
</script>

<svelte:head><title>Event Booking</title></svelte:head>

<main class="min-h-screen px-4 py-12">
	<div class="mx-auto flex max-w-2xl flex-col items-center justify-center gap-8 py-32">
		<div class="flex flex-col items-center gap-2 text-center">
			<h1 class="text-3xl font-bold tracking-tight">Event Booking</h1>
			<p class="text-muted-foreground text-sm">We recommed to call instead</p>
		</div>

		<div class="flex items-center gap-4">
			<Button
				variant="outline"
				class="flex h-auto min-w-36 items-center gap-2 rounded-2xl px-10 py-6"
				href="/view/admin/events"
			>
				<span class="text-base font-medium">Admin</span>
				<ShieldCheck size={18} />
			</Button>
			<Button
				variant="outline"
				class="flex h-auto min-w-36 items-center gap-2 rounded-2xl px-10 py-6"
				href="/view/client/events"
			>
				<span class="text-base font-medium">Client</span>
				<User size={18} />
			</Button>
		</div>
	</div>
</main>
