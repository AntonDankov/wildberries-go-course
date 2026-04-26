<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import {
		RefreshCw,
		CircleAlert,
		LoaderCircle,
		Users,
		Clock,
		BookOpen,
		ChevronLeft
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	type BookStatus = 0 | 1 | 2;

	type EventDTO = {
		id: number;
		name: string;
		seats: number;
		book_second_max_time: number;
	};

	type BookDTO = {
		id: number;
		event_id: number;
		book_status: BookStatus;
		created_at: string;
		updated_at: string;
	};

	const eventId = $page.params.id;

	let event = $state<EventDTO | null>(null);
	let books = $state<BookDTO[]>([]);
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

	const statusConfig: Record<BookStatus, { label: string; class: string }> = {
		0: { label: 'Pending', class: 'border-amber-500   text-amber-500' },
		1: { label: 'Confirmed', class: 'border-emerald-500 text-emerald-500' },
		2: { label: 'Cancelled', class: 'border-destructive  text-destructive' }
	};

	function startTimer() {
		if (timer) clearInterval(timer);
		timer = setInterval(() => fetchAll(), refreshInterval);
	}

	function changeInterval(ms: number) {
		refreshInterval = ms;
		startTimer();
	}

	async function fetchAll() {
		loading = true;
		error = null;
		try {
			const [evRes, bkRes] = await Promise.all([
				fetch(`http://localhost:8080/events/${eventId}`),
				fetch(`http://localhost:8080/events/${eventId}/book`)
			]);
			if (!evRes.ok) throw new Error(`Event fetch failed: ${evRes.status}`);
			if (!bkRes.ok) throw new Error(`Books fetch failed: ${bkRes.status}`);
			const evData = await evRes.json();
			const bkData = await bkRes.json();
			event = evData.event;
			books = bkData.books ?? [];
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
		fetchAll();
		startTimer();
	});
	onDestroy(() => {
		if (timer) clearInterval(timer);
	});
</script>

<svelte:head>
	<title>{event ? event.name : 'Event'} — Admin</title>
</svelte:head>

<main class="min-h-screen px-4 py-12">
	<div class="mx-auto max-w-2xl">
		<div class="mb-8 flex items-center justify-between">
			<a
				href="/view/admin/events"
				class="text-muted-foreground hover:text-foreground flex items-center gap-1.5 text-sm transition-colors"
			>
				<ChevronLeft size={15} />
				Back to events
			</a>
		</div>

		{#if loading && !event}
			<div class="text-muted-foreground flex flex-col items-center justify-center gap-3 py-16">
				<LoaderCircle size={32} class="animate-spin" />
				<p class="text-sm">Loading…</p>
			</div>
		{:else if error && !event}
			<div class="flex flex-col items-center justify-center gap-3 py-16">
				<p class="text-destructive flex items-center gap-1.5 text-sm">
					<CircleAlert size={14} />
					{error}
				</p>
				<Button
					variant="outline"
					class="border-destructive text-destructive hover:bg-destructive/10"
					onclick={fetchAll}
				>
					Retry
				</Button>
			</div>
		{:else if event}
			<section class="bg-card border-border mb-8 flex flex-col gap-4 rounded-xl border p-6">
				<div class="flex items-start justify-between gap-4">
					<div class="flex flex-col gap-1">
						<h1 class="text-2xl font-bold tracking-tight">{event.name}</h1>
						<Badge variant="outline" class="text-muted-foreground w-fit font-mono text-xs">
							#{event.id}
						</Badge>
					</div>
				</div>
				<Separator />
				<div class="flex flex-wrap gap-6">
					<div class="flex items-center gap-2">
						<Users size={15} class="text-muted-foreground shrink-0" />
						<div class="flex flex-col">
							<span class="text-muted-foreground text-xs">Seats</span>
							<span class="text-sm font-medium">{event.seats}</span>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<Clock size={15} class="text-muted-foreground shrink-0" />
						<div class="flex flex-col">
							<span class="text-muted-foreground text-xs">Book timeout</span>
							<span class="text-sm font-medium">{formatBookTime(event.book_second_max_time)}</span>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<BookOpen size={15} class="text-muted-foreground shrink-0" />
						<div class="flex flex-col">
							<span class="text-muted-foreground text-xs">Total bookings</span>
							<span class="text-sm font-medium">{books.length}</span>
						</div>
					</div>
				</div>
			</section>

			<Separator class="mb-6" />

			<div class="mb-4 flex flex-wrap items-center justify-between gap-4">
				<h2 class="text-lg font-semibold">Bookings</h2>
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

					<Button variant="outline" size="icon" class="size-8" onclick={fetchAll}>
						<RefreshCw size={14} class={loading ? 'animate-spin' : ''} />
					</Button>
				</div>
			</div>

			{#if error}
				<div class="text-destructive mb-4 flex items-center gap-1.5 text-sm">
					<CircleAlert size={14} />
					{error}
				</div>
			{/if}

			{#if books.length === 0}
				<div class="text-muted-foreground flex items-center justify-center py-16">
					<p class="text-sm">No bookings for this event yet.</p>
				</div>
			{:else}
				<ul class="flex flex-col gap-2.5">
					{#each books as book (book.id)}
						{@const status = statusConfig[book.book_status]}
						<li
							class="bg-card border-border flex items-center justify-between rounded-xl border px-5 py-4"
						>
							<div class="flex flex-col gap-1">
								<span class="text-sm font-medium">Booking #{book.id}</span>
								<span class="text-muted-foreground text-xs">
									{#if book.created_at === book.updated_at}
										Created {new Date(book.created_at).toLocaleString()}
									{:else}
										Created {new Date(book.created_at).toLocaleString()}
										| Updated {new Date(book.updated_at).toLocaleString()}
									{/if}
								</span>
							</div>
							<Badge variant="outline" class="shrink-0 text-xs {status.class}">
								{status.label}
							</Badge>
						</li>
					{/each}
				</ul>
			{/if}
		{/if}
	</div>
</main>
