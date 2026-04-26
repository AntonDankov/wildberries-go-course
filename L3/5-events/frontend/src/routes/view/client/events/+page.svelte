<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { RefreshCw, CircleAlert, LoaderCircle, Users, Clock } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	type EventDTO = {
		id: number;
		name: string;
		seats: number;
		book_second_max_time: number;
	};

	let events = $state<EventDTO[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let lastUpdated = $state<Date | null>(null);

	const refreshOptions = [
		{ label: '5s', ms: 5_000 },
		{ label: '15s', ms: 15_000 },
		{ label: '30s', ms: 30_000 },
		{ label: '1min', ms: 60_000 },
		{ label: '5min', ms: 300_000 }
	];
	let refreshInterval = $state(15_000);
	let timer: ReturnType<typeof setInterval> | null = null;

	function startTimer() {
		if (timer) clearInterval(timer);
		timer = setInterval(() => fetchEvents(), refreshInterval);
	}

	function changeInterval(ms: number) {
		refreshInterval = ms;
		startTimer();
	}

	async function fetchEvents() {
		loading = true;
		error = null;
		try {
			const res = await fetch('http://localhost:8080/events');
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			const data = await res.json();
			events = data.events ?? [];
			lastUpdated = new Date();
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	function formatBookTime(seconds: number): string {
		if (!seconds) return '—';
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		return s > 0 ? `${m}m ${s}s` : `${m}m`;
	}

	onMount(() => {
		fetchEvents();
		startTimer();
	});
	onDestroy(() => {
		if (timer) clearInterval(timer);
	});
</script>

<svelte:head><title>Events</title></svelte:head>

<main class="min-h-screen px-4 py-12">
	<div class="mx-auto max-w-2xl">
		<h1 class="mb-8 text-3xl font-bold tracking-tight">Events</h1>

		<Separator class="mb-6" />

		<div class="mb-4 flex flex-wrap items-center justify-between gap-4">
			<h2 class="text-lg font-semibold">Available Events</h2>
			<div class="flex flex-wrap items-center justify-end gap-3">
				{#if lastUpdated}
					<span class="text-muted-foreground text-xs whitespace-nowrap">
						Updated: {lastUpdated.toLocaleTimeString()}
					</span>
				{/if}

				<Select.Root
					type="single"
					value={String(refreshInterval)}
					onValueChange={(v) => changeInterval(Number(v))}
				>
					<Select.Trigger class="h-8 w-16 text-xs">
						{refreshOptions.find((o) => o.ms === refreshInterval)?.label ?? 'Auto'}
					</Select.Trigger>
					<Select.Content>
						{#each refreshOptions as opt}
							<Select.Item value={String(opt.ms)} class="text-xs">{opt.label}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>

				<Button variant="outline" size="icon" class="size-8" onclick={fetchEvents}>
					<RefreshCw size={14} class={loading ? 'animate-spin' : ''} />
				</Button>
			</div>
		</div>

		{#if loading && events.length === 0}
			<div class="text-muted-foreground flex flex-col items-center justify-center gap-3 py-16">
				<LoaderCircle size={32} class="animate-spin" />
				<p class="text-sm">Loading…</p>
			</div>
		{:else if error}
			<div class="flex flex-col items-center justify-center gap-3 py-16">
				<p class="text-destructive flex items-center gap-1.5 text-sm">
					<CircleAlert size={14} />
					{error}
				</p>
				<Button
					variant="outline"
					class="border-destructive text-destructive hover:bg-destructive/10"
					onclick={fetchEvents}
				>
					Retry
				</Button>
			</div>
		{:else if events.length === 0}
			<div class="text-muted-foreground flex items-center justify-center py-16">
				<p class="text-sm">No events available at the moment.</p>
			</div>
		{:else}
			<ul class="flex flex-col gap-2.5">
				{#each events as event (event.id)}
					<li
						class="bg-card border-border hover:bg-muted/50 relative flex items-center justify-between rounded-xl border px-5 py-4 transition-colors"
					>
						<a
							href="/view/client/events/{event.id}"
							class="focus-visible:ring-ring absolute inset-0 rounded-xl focus-visible:ring-2 focus-visible:outline-none"
							aria-label="View event {event.name}"
						></a>
						<div class="flex min-w-0 flex-col gap-1">
							<span class="truncate font-medium">{event.name}</span>
							<div class="text-muted-foreground flex items-center gap-3 text-xs">
								<span class="flex items-center gap-1">
									<Users size={11} />
									{event.seats} seats
								</span>
								<span class="flex items-center gap-1">
									<Clock size={11} />
									{formatBookTime(event.book_second_max_time)}
								</span>
							</div>
						</div>
						<Badge
							variant="outline"
							class="text-muted-foreground relative z-10 shrink-0 font-mono text-xs"
						>
							#{event.id}
						</Badge>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</main>
