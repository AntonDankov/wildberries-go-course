<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		RefreshCw,
		CircleAlert,
		CircleCheck,
		LoaderCircle,
		Plus,
		Pencil,
		Trash2,
		BarChart3,
		TrendingUp,
		Hash,
		DollarSign,
		Activity,
		X,
		Filter
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	type RecordDTO = {
		id: number;
		type: number;
		category: number;
		amount: number;
		date: string;
	};

	type Analytics = {
		sum: number;
		average: number;
		count: number;
		median: number;
		percentile: number;
	};

	const typeLabels: Record<number, string> = { 0: 'Expense', 1: 'Income' };
	const categoryLabels: Record<number, string> = {
		0: 'Electronics',
		1: 'Food',
		2: 'Delivery',
		3: 'Taxes'
	};

	let filterFrom = $state('');
	let filterTo = $state('');
	let filterType = $state('all');
	let filterCategory = $state('all');

	let records = $state<RecordDTO[]>([]);
	let recordsLoading = $state(false);
	let recordsError = $state<string | null>(null);
	let lastUpdated = $state<Date | null>(null);

	let analytics = $state<Analytics | null>(null);
	let analyticsLoading = $state(false);
	let analyticsError = $state<string | null>(null);

	let showRefreshSpinner = $state(false);
	let _spinnerTimer: ReturnType<typeof setTimeout> | null = null;

	$effect(() => {
		const active = recordsLoading || analyticsLoading;
		if (active) {
			const t = setTimeout(() => (showRefreshSpinner = true), 300);
			return () => clearTimeout(t);
		} else {
			showRefreshSpinner = false;
		}
	});

	const refreshOptions = [
		{ label: '5s', ms: 5_000 },
		{ label: '15s', ms: 15_000 },
		{ label: '30s', ms: 30_000 },
		{ label: '1min', ms: 60_000 },
		{ label: '5min', ms: 300_000 }
	];
	let refreshInterval = $state(15_000);
	let timer: ReturnType<typeof setInterval> | null = null;

	let formType = $state('0');
	let formCategory = $state('0');
	let formAmount = $state('');
	let formDate = $state('');
	let editingId = $state<number | null>(null);
	let submitting = $state(false);
	let formError = $state<string | null>(null);
	let formSuccess = $state<string | null>(null);

	let deletingId = $state<number | null>(null);
	let deleteError = $state<string | null>(null);

	function buildFilterQuery(): string {
		const params = new URLSearchParams();
		if (filterFrom) params.set('from', filterFrom);
		if (filterTo) params.set('to', filterTo);
		if (filterType !== 'all') params.set('type', filterType);
		if (filterCategory !== 'all') params.set('category', filterCategory);
		const qs = params.toString();
		return qs ? `?${qs}` : '';
	}

	function startTimer() {
		if (timer) clearInterval(timer);
		timer = setInterval(() => fetchAll(), refreshInterval);
	}

	function changeInterval(ms: number) {
		refreshInterval = ms;
		startTimer();
	}

	async function fetchRecords() {
		recordsLoading = true;
		recordsError = null;
		try {
			const res = await fetch(`http://localhost:8080/items${buildFilterQuery()}`);
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			const data = await res.json();
			records = data.records ?? [];
			lastUpdated = new Date();
		} catch (e: any) {
			recordsError = e.message;
		} finally {
			recordsLoading = false;
		}
	}

	async function fetchAnalytics() {
		analyticsLoading = true;
		analyticsError = null;
		try {
			const res = await fetch(`http://localhost:8080/analytics${buildFilterQuery()}`);
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			analytics = await res.json();
		} catch (e: any) {
			analyticsError = e.message;
		} finally {
			analyticsLoading = false;
		}
	}

	async function fetchAll() {
		await Promise.all([fetchRecords(), fetchAnalytics()]);
	}

	async function submitForm() {
		if (!formAmount || !formDate) {
			formError = 'Amount and date are required';
			return;
		}
		const amountNum = Number(formAmount);
		if (amountNum < 0) {
			formError = 'Amount must be non-negative';
			return;
		}

		submitting = true;
		formError = null;
		formSuccess = null;

		try {
			const url = editingId
				? `http://localhost:8080/items/${editingId}`
				: 'http://localhost:8080/items';
			const method = editingId ? 'PUT' : 'POST';
			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					type: Number(formType),
					category: Number(formCategory),
					amount: amountNum,
					date: formDate
				})
			});
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}
			formSuccess = editingId ? 'Record updated!' : 'Record created!';
			resetForm();
			await fetchAll();
		} catch (e: any) {
			formError = e.message;
		} finally {
			submitting = false;
		}
	}

	function startEdit(record: RecordDTO) {
		editingId = record.id;
		formType = String(record.type);
		formCategory = String(record.category);
		formAmount = String(record.amount);
		formDate = record.date;
		formError = null;
		formSuccess = null;
	}

	function resetForm() {
		editingId = null;
		formType = '0';
		formCategory = '0';
		formAmount = '';
		formDate = '';
		formError = null;
	}

	async function deleteRecord(id: number) {
		deletingId = id;
		deleteError = null;
		try {
			const res = await fetch(`http://localhost:8080/items/${id}`, { method: 'DELETE' });
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}
			await fetchAll();
		} catch (e: any) {
			deleteError = e.message;
		} finally {
			deletingId = null;
		}
	}

	function formatAmount(n: number | null | undefined): string {
		if (n == null || isNaN(Number(n))) return '—';
		return Number(n).toLocaleString('en-US', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		});
	}

	onMount(() => {
		fetchAll();
		startTimer();
	});

	onDestroy(() => {
		if (timer) clearInterval(timer);
	});
</script>

<svelte:head><title>Sales Tracker</title></svelte:head>

<main class="min-h-screen px-4 py-12">
	<div class="mx-auto max-w-7xl">
		<h1 class="mb-8 text-3xl font-bold tracking-tight">Sales Tracker</h1>

		<section class="bg-card border-border mb-6 rounded-xl border p-5">
			<div class="flex flex-wrap items-end gap-3">
				<div class="flex flex-col gap-1.5">
					<label class="text-muted-foreground text-xs font-medium">From</label>
					<Input type="date" bind:value={filterFrom} class="h-8 w-36 text-xs" />
				</div>
				<div class="flex flex-col gap-1.5">
					<label class="text-muted-foreground text-xs font-medium">To</label>
					<Input type="date" bind:value={filterTo} class="h-8 w-36 text-xs" />
				</div>
				<div class="flex flex-col gap-1.5">
					<label class="text-muted-foreground text-xs font-medium">Type</label>
					<Select.Root type="single" value={filterType} onValueChange={(v) => (filterType = v)}>
						<Select.Trigger class="h-8 w-28 text-xs">
							{filterType === 'all' ? 'All types' : typeLabels[Number(filterType)]}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="all" class="text-xs">All</Select.Item>
							<Select.Item value="0" class="text-xs">Expense</Select.Item>
							<Select.Item value="1" class="text-xs">Income</Select.Item>
						</Select.Content>
					</Select.Root>
				</div>
				<div class="flex flex-col gap-1.5">
					<label class="text-muted-foreground text-xs font-medium">Category</label>
					<Select.Root
						type="single"
						value={filterCategory}
						onValueChange={(v) => (filterCategory = v)}
					>
						<Select.Trigger class="h-8 w-32 text-xs">
							{filterCategory === 'all' ? 'All categories' : categoryLabels[Number(filterCategory)]}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="all" class="text-xs">All</Select.Item>
							<Select.Item value="0" class="text-xs">Electronics</Select.Item>
							<Select.Item value="1" class="text-xs">Food</Select.Item>
							<Select.Item value="2" class="text-xs">Delivery</Select.Item>
							<Select.Item value="3" class="text-xs">Taxes</Select.Item>
						</Select.Content>
					</Select.Root>
				</div>

				<Button size="sm" class="h-8 rounded-full text-xs" onclick={fetchAll}>
					<Filter size={12} />
					Apply
				</Button>
				<Button
					size="sm"
					variant="outline"
					class="h-8 rounded-full text-xs"
					onclick={() => {
						filterFrom = '';
						filterTo = '';
						filterType = 'all';
						filterCategory = 'all';
						fetchAll();
					}}
				>
					<X size={12} />
					Reset
				</Button>

				<div class="ml-auto flex flex-wrap items-center justify-end gap-3">
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
						<RefreshCw size={14} class={showRefreshSpinner ? 'animate-spin' : ''} />
					</Button>
				</div>
			</div>
		</section>

		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<div class="flex flex-col gap-5">
				<section
					class="bg-card rounded-xl border p-5 transition-colors {editingId
						? 'border-amber-400'
						: 'border-border'}"
				>
					<h2 class="mb-4 text-base font-semibold">
						{editingId ? `Edit Record #${editingId}` : 'New Record'}
					</h2>
					<div class="flex flex-col gap-3">
						<div class="flex gap-2">
							<Select.Root type="single" value={formType} onValueChange={(v) => (formType = v)}>
								<Select.Trigger class="h-9 flex-1 text-sm">
									{typeLabels[Number(formType)]}
								</Select.Trigger>
								<Select.Content>
									<Select.Item value="0" class="text-sm">Expense</Select.Item>
									<Select.Item value="1" class="text-sm">Income</Select.Item>
								</Select.Content>
							</Select.Root>
							<Select.Root
								type="single"
								value={formCategory}
								onValueChange={(v) => (formCategory = v)}
							>
								<Select.Trigger class="h-9 flex-1 text-sm">
									{categoryLabels[Number(formCategory)]}
								</Select.Trigger>
								<Select.Content>
									<Select.Item value="0" class="text-sm">Electronics</Select.Item>
									<Select.Item value="1" class="text-sm">Food</Select.Item>
									<Select.Item value="2" class="text-sm">Delivery</Select.Item>
									<Select.Item value="3" class="text-sm">Taxes</Select.Item>
								</Select.Content>
							</Select.Root>
						</div>

						<div class="flex gap-2">
							<div class="relative flex-1">
								<DollarSign
									size={14}
									class="text-muted-foreground absolute top-1/2 left-3 -translate-y-1/2"
								/>
								<Input
									type="number"
									placeholder="Amount"
									bind:value={formAmount}
									min="0"
									step="0.01"
									class="[appearance:textfield] pl-8 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
								/>
							</div>
							<Input type="date" bind:value={formDate} class="flex-1" />
						</div>

						{#if formError}
							<div class="text-destructive flex items-center gap-1.5 text-sm">
								<CircleAlert size={14} />
								{formError}
							</div>
						{/if}
						{#if formSuccess}
							<div class="flex items-center gap-1.5 text-sm text-emerald-500">
								<CircleCheck size={14} />
								{formSuccess}
							</div>
						{/if}

						<div class="flex items-center justify-end gap-2">
							{#if editingId}
								<Button variant="ghost" size="sm" class="rounded-full text-xs" onclick={resetForm}>
									<X size={12} />
									Cancel
								</Button>
							{/if}
							<Button
								class="rounded-full"
								size="sm"
								disabled={!formAmount || !formDate || submitting}
								onclick={submitForm}
							>
								{#if submitting}
									<LoaderCircle size={14} class="animate-spin" />
								{:else if editingId}
									<Pencil size={14} />
								{:else}
									<Plus size={14} />
								{/if}
								{submitting
									? editingId
										? 'Updating…'
										: 'Creating…'
									: editingId
										? 'Update Record'
										: 'Create Record'}
							</Button>
						</div>
					</div>
				</section>

				<div>
					<!-- Analytics table -->
					<div class="mb-3">
						<h2 class="text-lg font-semibold">Analytics</h2>
					</div>

					{#if analyticsLoading && !analytics}
						<div
							class="text-muted-foreground flex flex-col items-center justify-center gap-3 py-16"
						>
							<LoaderCircle size={32} class="animate-spin" />
							<p class="text-sm">Loading…</p>
						</div>
					{:else if analyticsError && !analytics}
						<div class="flex flex-col items-center justify-center gap-3 py-12">
							<p class="text-destructive flex items-center gap-1.5 text-sm">
								<CircleAlert size={14} />
								{analyticsError}
							</p>
							<Button
								variant="outline"
								class="border-destructive text-destructive hover:bg-destructive/10"
								onclick={fetchAnalytics}
							>
								Retry
							</Button>
						</div>
					{:else if analytics}
						<section
							class="bg-card border-border relative rounded-xl border p-5 transition-opacity {showRefreshSpinner
								? 'opacity-60'
								: ''}"
						>
							{#if showRefreshSpinner}
								<div class="absolute inset-0 flex items-center justify-center rounded-xl">
									<LoaderCircle size={20} class="text-muted-foreground animate-spin" />
								</div>
							{/if}
							<ul class="divide-border flex flex-col divide-y">
								<li class="flex items-center justify-between py-3.5">
									<div class="text-muted-foreground flex items-center gap-2 text-sm">
										<DollarSign size={14} />
										<span>Sum</span>
									</div>
									<span class="font-semibold">${formatAmount(analytics.sum)}</span>
								</li>
								<li class="flex items-center justify-between py-3.5">
									<div class="text-muted-foreground flex items-center gap-2 text-sm">
										<BarChart3 size={14} />
										<span>Average</span>
									</div>
									<span class="font-semibold">${formatAmount(analytics.average)}</span>
								</li>
								<li class="flex items-center justify-between py-3.5">
									<div class="text-muted-foreground flex items-center gap-2 text-sm">
										<Hash size={14} />
										<span>Count</span>
									</div>
									<span class="font-semibold">{analytics.count}</span>
								</li>
								<li class="flex items-center justify-between py-3.5">
									<div class="text-muted-foreground flex items-center gap-2 text-sm">
										<TrendingUp size={14} />
										<span>Median</span>
									</div>
									<span class="font-semibold">${formatAmount(analytics.median)}</span>
								</li>
								<li class="flex items-center justify-between py-3.5">
									<div class="text-muted-foreground flex items-center gap-2 text-sm">
										<Activity size={14} />
										<span>95th Percentile</span>
									</div>
									<span class="font-semibold">${formatAmount(analytics.percentile)}</span>
								</li>
							</ul>
						</section>
					{/if}
					<!-- End of Analytics table -->
				</div>
			</div>

			<div>
				<!-- Records table -->
				<div class="mb-3 flex items-center justify-between">
					<h2 class="text-lg font-semibold">Records</h2>
					{#if showRefreshSpinner && records.length > 0}
						<LoaderCircle size={14} class="text-muted-foreground animate-spin" />
					{/if}
				</div>

				{#if deleteError}
					<div class="text-destructive mb-3 flex items-center gap-1.5 text-sm">
						<CircleAlert size={14} />
						{deleteError}
					</div>
				{/if}

				{#if recordsLoading && records.length === 0}
					<div class="text-muted-foreground flex flex-col items-center justify-center gap-3 py-16">
						<LoaderCircle size={32} class="animate-spin" />
						<p class="text-sm">Loading…</p>
					</div>
				{:else if recordsError && records.length === 0}
					<div class="flex flex-col items-center justify-center gap-3 py-12">
						<p class="text-destructive flex items-center gap-1.5 text-sm">
							<CircleAlert size={14} />
							{recordsError}
						</p>
						<Button
							variant="outline"
							class="border-destructive text-destructive hover:bg-destructive/10"
							onclick={fetchRecords}
						>
							Retry
						</Button>
					</div>
				{:else if records.length === 0}
					<div class="text-muted-foreground flex items-center justify-center py-16">
						<p class="text-sm">No records yet. Create one above.</p>
					</div>
				{:else}
					<ul class="flex flex-col gap-2">
						{#each records as record (record.id)}
							<li
								class="bg-card hover:bg-muted/50 flex items-center justify-between gap-3 rounded-xl border px-4 py-3 transition-colors {editingId ===
								record.id
									? 'border-amber-400'
									: 'border-border'}"
							>
								<div class="flex min-w-0 flex-col gap-1.5">
									<div class="flex items-center gap-3">
										<span class="text-foreground font-medium">${formatAmount(record.amount)}</span>
										<span class="text-muted-foreground text-xs"
											>{record.date.split('-').reverse().join('-')}</span
										>
									</div>
									<div class="flex flex-wrap items-center gap-1.5">
										<Badge
											variant="outline"
											class="text-xs {record.type === 1
												? 'border-emerald-500 text-emerald-500'
												: 'border-destructive text-destructive'}"
										>
											{typeLabels[record.type]}
										</Badge>
										<Badge variant="outline" class="text-muted-foreground text-xs">
											{categoryLabels[record.category]}
										</Badge>
									</div>
								</div>

								<div class="flex shrink-0 items-center gap-1">
									<Button
										variant="ghost"
										size="icon"
										class="size-7"
										onclick={() => startEdit(record)}
									>
										<Pencil size={12} />
									</Button>
									<Button
										variant="ghost"
										size="icon"
										class="text-destructive hover:bg-destructive/10 hover:text-destructive size-7"
										disabled={deletingId === record.id}
										onclick={() => deleteRecord(record.id)}
									>
										{#if deletingId === record.id}
											<LoaderCircle size={12} class="animate-spin" />
										{:else}
											<Trash2 size={12} />
										{/if}
									</Button>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
				<!-- End of Records table -->
			</div>
		</div>
	</div>
</main>
