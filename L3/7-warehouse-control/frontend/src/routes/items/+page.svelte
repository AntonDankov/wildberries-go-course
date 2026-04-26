<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		RefreshCw,
		CircleAlert,
		CircleCheck,
		LoaderCircle,
		Plus,
		Pencil,
		Trash2,
		History,
		DollarSign,
		Package,
		LogOut,
		User,
		Check,
		X
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	type ItemDTO = {
		id: number;
		owner_id: number;
		name: string;
		price: number;
		amount: number;
		created_at: string;
		updated_at: string;
	};

	type ActionType = 0 | 1 | 2;

	type ItemHistory = {
		id: number;
		item_id: number;
		name?: string;
		price?: number;
		amount?: number;
		action: ActionType;
		user_id: number;
		username: string;
		changed_at: string;
	};

	const OWNER = 2;
	const MANAGER = 4;
	const ADMIN = 8;

	const roleLabels: Record<number, string> = {
		1: 'Viewer',
		2: 'Owner',
		4: 'Manager',
		8: 'Admin'
	};

	const actionConfig: Record<ActionType, { label: string; class: string }> = {
		0: { label: 'Insert', class: 'border-emerald-500 text-emerald-500' },
		1: { label: 'Update', class: 'border-amber-500 text-amber-500' },
		2: { label: 'Delete', class: 'border-destructive text-destructive' }
	};

	let userName = $state('');
	let userRole = $state(0);
	let userId = $state(0);
	let token = $state('');

	let canMutate = $derived((userRole & (OWNER | MANAGER | ADMIN)) !== 0);
	let canViewHistory = $derived((userRole & (MANAGER | ADMIN)) !== 0);

	// Items
	let items = $state<ItemDTO[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let lastUpdated = $state<Date | null>(null);

	// Debounced spinner
	let showRefreshSpinner = $state(false);
	$effect(() => {
		if (loading) {
			const t = setTimeout(() => (showRefreshSpinner = true), 300);
			return () => clearTimeout(t);
		} else {
			showRefreshSpinner = false;
		}
	});

	// Auto-refresh
	const refreshOptions = [
		{ label: '5s', ms: 5_000 },
		{ label: '15s', ms: 15_000 },
		{ label: '30s', ms: 30_000 },
		{ label: '1min', ms: 60_000 },
		{ label: '5min', ms: 300_000 }
	];
	let refreshInterval = $state(15_000);
	let timer: ReturnType<typeof setInterval> | null = null;

	// Create form
	let formName = $state('');
	let formPrice = $state('');
	let formAmount = $state('');
	let submitting = $state(false);
	let formError = $state<string | null>(null);
	let formSuccess = $state<string | null>(null);

	// Edit in-place
	let editingId = $state<number | null>(null);
	let editName = $state('');
	let editPrice = $state('');
	let editAmount = $state('');
	let updatingId = $state<number | null>(null);
	let updateError = $state<string | null>(null);

	// Delete
	let deletingId = $state<number | null>(null);

	// History
	let expandedHistoryId = $state<number | null>(null);
	let historyMap = $state<Record<number, ItemHistory[]>>({});
	let historyLoadingId = $state<number | null>(null);
	let historyError = $state<string | null>(null);

	function authHeaders(): HeadersInit {
		return {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`
		};
	}

	function handleUnauthorized() {
		localStorage.clear();
		goto('/');
	}

	function startTimer() {
		if (timer) clearInterval(timer);
		timer = setInterval(() => fetchItems(), refreshInterval);
	}

	function changeInterval(ms: number) {
		refreshInterval = ms;
		startTimer();
	}

	function canMutateItem(item: ItemDTO): boolean {
		if (!canMutate) return false;
		if (userRole & (MANAGER | ADMIN)) return true;
		return item.owner_id === userId;
	}

	async function fetchItems() {
		loading = true;
		error = null;
		try {
			const res = await fetch('http://localhost:8080/items');
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			const data = await res.json();
			items = data.items ?? [];
			lastUpdated = new Date();
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function createItem() {
		if (!formName || !formPrice || !formAmount) {
			formError = 'All fields are required';
			return;
		}
		const priceNum = Number(formPrice);
		const amountNum = Number(formAmount);
		if (priceNum < 0) {
			formError = 'Price must be non-negative';
			return;
		}
		if (amountNum < 0) {
			formError = 'Amount must be non-negative';
			return;
		}

		submitting = true;
		formError = null;
		formSuccess = null;
		try {
			const res = await fetch('http://localhost:8080/items', {
				method: 'POST',
				headers: authHeaders(),
				body: JSON.stringify({ name: formName, price: priceNum, amount: amountNum })
			});
			if (res.status === 401) {
				handleUnauthorized();
				return;
			}
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}
			formSuccess = 'Item created!';
			formName = '';
			formPrice = '';
			formAmount = '';
			await fetchItems();
		} catch (e: any) {
			formError = e.message;
		} finally {
			submitting = false;
		}
	}

	function startEdit(item: ItemDTO) {
		editingId = item.id;
		editName = item.name;
		editPrice = String(item.price);
		editAmount = String(item.amount);
		updateError = null;
	}

	function cancelEdit() {
		editingId = null;
		updateError = null;
	}

	async function saveEdit(id: number) {
		if (!editName || !editPrice || !editAmount) {
			updateError = 'All fields are required';
			return;
		}
		const priceNum = Number(editPrice);
		const amountNum = Number(editAmount);
		if (priceNum < 0) {
			updateError = 'Price must be non-negative';
			return;
		}
		if (amountNum < 0) {
			updateError = 'Amount must be non-negative';
			return;
		}

		updatingId = id;
		updateError = null;
		try {
			const res = await fetch(`http://localhost:8080/items/${id}`, {
				method: 'PUT',
				headers: authHeaders(),
				body: JSON.stringify({ name: editName, price: priceNum, amount: amountNum })
			});
			if (res.status === 401) {
				handleUnauthorized();
				return;
			}
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}
			editingId = null;
			await fetchItems();
			if (expandedHistoryId === id) {
				await updateHistory(id);
			}
		} catch (e: any) {
			updateError = e.message;
		} finally {
			updatingId = null;
		}
	}

	async function deleteItem(id: number) {
		deletingId = id;
		try {
			const res = await fetch(`http://localhost:8080/items/${id}`, {
				method: 'DELETE',
				headers: authHeaders()
			});
			if (res.status === 401) {
				handleUnauthorized();
				return;
			}
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}
			if (expandedHistoryId === id) expandedHistoryId = null;
			if (editingId === id) editingId = null;
			await fetchItems();
		} catch (e: any) {
			error = e.message;
		} finally {
			deletingId = null;
		}
	}

	async function toggleHistory(id: number) {
		if (expandedHistoryId === id) {
			expandedHistoryId = null;
			return;
		}
		updateHistory(id);
	}

	async function updateHistory(id: number) {
		expandedHistoryId = id;
		historyLoadingId = id;
		historyError = null;
		try {
			const res = await fetch(`http://localhost:8080/analytics/${id}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.status === 401) {
				handleUnauthorized();
				return;
			}
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}
			const data = await res.json();
			historyMap = { ...historyMap, [id]: data.history ?? [] };
		} catch (e: any) {
			historyError = e.message;
			expandedHistoryId = null;
		} finally {
			historyLoadingId = null;
		}
	}

	function formatAmount(n: number | null | undefined): string {
		if (n == null || isNaN(Number(n))) return '—';
		return Number(n).toLocaleString('en-US', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		});
	}

	function formatDate(s: string): string {
		return new Date(s).toLocaleString();
	}

	function logout() {
		localStorage.clear();
		goto('/');
	}

	onMount(() => {
		const t = localStorage.getItem('token');
		if (!t) {
			goto('/');
			return;
		}
		token = t;
		userName = localStorage.getItem('name') ?? '';
		userRole = Number(localStorage.getItem('role') ?? '0');
		userId = Number(localStorage.getItem('user_id') ?? '0');
		fetchItems();
		startTimer();
	});

	onDestroy(() => {
		if (timer) clearInterval(timer);
	});
</script>

<svelte:head><title>Warehouse</title></svelte:head>

<main class="min-h-screen px-4 py-12">
	<div class="mx-auto max-w-2xl">
		<!-- Header -->
		<div class="mb-8 flex items-center justify-between gap-4">
			<h1 class="text-3xl font-bold tracking-tight">Warehouse</h1>
			<div class="flex items-center gap-2">
				<User size={13} class="text-muted-foreground" />
				<span class="text-sm font-medium">{userName}</span>
				<Badge variant="outline" class="text-muted-foreground text-xs">
					{roleLabels[userRole] ?? `Role ${userRole}`}
				</Badge>
				<Button variant="ghost" size="icon" class="size-8" onclick={logout}>
					<LogOut size={14} />
				</Button>
			</div>
		</div>

		<!-- Create form -->
		{#if canMutate}
			<section class="bg-card border-border mb-8 flex flex-col gap-4 rounded-xl border p-6">
				<h2 class="text-base font-semibold">New Item</h2>

				<Input placeholder="Item name" bind:value={formName} />

				<div class="flex items-center gap-2">
					<div class="relative flex-1">
						<DollarSign
							size={14}
							class="text-muted-foreground absolute top-1/2 left-3 -translate-y-1/2"
						/>
						<Input
							type="number"
							placeholder="Price"
							bind:value={formPrice}
							min="0"
							step="0.01"
							class="[appearance:textfield] pl-8 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
						/>
					</div>
					<div class="relative flex-1">
						<Package
							size={14}
							class="text-muted-foreground absolute top-1/2 left-3 -translate-y-1/2"
						/>
						<Input
							type="number"
							placeholder="Amount"
							bind:value={formAmount}
							min="0"
							class="[appearance:textfield] pl-8 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
						/>
					</div>
				</div>

				{#if formError}
					<div class="text-destructive flex items-center gap-1.5 text-sm">
						<CircleAlert size={14} />{formError}
					</div>
				{/if}
				{#if formSuccess}
					<div class="flex items-center gap-1.5 text-sm text-emerald-500">
						<CircleCheck size={14} />{formSuccess}
					</div>
				{/if}

				<Button
					class="self-end rounded-full"
					disabled={!formName || !formPrice || !formAmount || submitting}
					onclick={createItem}
				>
					{#if submitting}<LoaderCircle size={14} class="animate-spin" />{:else}<Plus
							size={14}
						/>{/if}
					{submitting ? 'Creating…' : 'Create Item'}
				</Button>
			</section>
		{/if}

		<Separator class="mb-6" />

		<!-- List header -->
		<div class="mb-4 flex flex-wrap items-center justify-between gap-4">
			<div class="flex items-center gap-2">
				<h2 class="text-lg font-semibold">Items</h2>
				{#if showRefreshSpinner && items.length > 0}
					<LoaderCircle size={14} class="text-muted-foreground animate-spin" />
				{/if}
			</div>
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
				<Button variant="outline" size="icon" class="size-8" onclick={fetchItems}>
					<RefreshCw size={14} class={showRefreshSpinner ? 'animate-spin' : ''} />
				</Button>
			</div>
		</div>

		<!-- Items list -->
		{#if loading && items.length === 0}
			<div class="text-muted-foreground flex flex-col items-center justify-center gap-3 py-16">
				<LoaderCircle size={32} class="animate-spin" />
				<p class="text-sm">Loading…</p>
			</div>
		{:else if error && items.length === 0}
			<div class="flex flex-col items-center justify-center gap-3 py-16">
				<p class="text-destructive flex items-center gap-1.5 text-sm">
					<CircleAlert size={14} />{error}
				</p>
				<Button
					variant="outline"
					class="border-destructive text-destructive hover:bg-destructive/10"
					onclick={fetchItems}
				>
					Retry
				</Button>
			</div>
		{:else if items.length === 0}
			<div class="text-muted-foreground flex items-center justify-center py-16">
				<p class="text-sm">No items yet.{canMutate ? ' Create one above.' : ''}</p>
			</div>
		{:else}
			<ul class="flex flex-col gap-2.5">
				{#each items as item (item.id)}
					{@const isEditing = editingId === item.id}
					{@const isHistoryOpen = expandedHistoryId === item.id}

					<li
						class="bg-card flex flex-col rounded-xl border px-5 py-4 transition-colors {isEditing
							? 'border-amber-400'
							: 'border-border hover:bg-muted/50'}"
					>
						<!-- Content row -->
						<div class="flex items-start justify-between gap-3">
							{#if isEditing}
								<!-- Edit inputs -->
								<div class="flex flex-1 flex-col gap-2">
									<Input placeholder="Item name" bind:value={editName} class="h-8 text-sm" />
									<div class="flex gap-2">
										<div class="relative flex-1">
											<DollarSign
												size={12}
												class="text-muted-foreground absolute top-1/2 left-3 -translate-y-1/2"
											/>
											<Input
												type="number"
												placeholder="Price"
												bind:value={editPrice}
												min="0"
												step="0.01"
												class="h-8 [appearance:textfield] pl-8 text-sm [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
											/>
										</div>
										<div class="relative flex-1">
											<Package
												size={12}
												class="text-muted-foreground absolute top-1/2 left-3 -translate-y-1/2"
											/>
											<Input
												type="number"
												placeholder="Amount"
												bind:value={editAmount}
												min="0"
												class="h-8 [appearance:textfield] pl-8 text-sm [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
											/>
										</div>
									</div>
									{#if updateError}
										<div class="text-destructive flex items-center gap-1.5 text-xs">
											<CircleAlert size={12} />{updateError}
										</div>
									{/if}
								</div>
							{:else}
								<!-- Display -->
								<div class="flex min-w-0 flex-col gap-1">
									<span class="truncate font-medium">{item.name}</span>
									<div class="text-muted-foreground flex flex-wrap items-center gap-3 text-xs">
										<span class="flex items-center gap-1">
											<DollarSign size={11} />{formatAmount(item.price)}
										</span>
										<span class="flex items-center gap-1">
											<Package size={11} />{item.amount} units
										</span>
										<span>{formatDate(item.created_at)}</span>
									</div>
								</div>
							{/if}

							<Badge variant="outline" class="text-muted-foreground shrink-0 font-mono text-xs">
								#{item.id}
							</Badge>
						</div>

						<!-- Action row -->
						{#if canMutate || canViewHistory}
							<div class="border-border mt-3 flex items-center gap-1 border-t pt-2.5">
								{#if canMutateItem(item)}
									{#if isEditing}
										<Button
											variant="ghost"
											size="sm"
											class="h-7 rounded-full px-3 text-xs text-emerald-500 hover:bg-emerald-500/10 hover:text-emerald-500"
											disabled={updatingId === item.id}
											onclick={() => saveEdit(item.id)}
										>
											{#if updatingId === item.id}
												<LoaderCircle size={12} class="animate-spin" />
											{:else}
												<Check size={12} />
											{/if}
											{updatingId === item.id ? 'Saving…' : 'Save'}
										</Button>

										<Button
											variant="ghost"
											size="sm"
											class="text-destructive hover:bg-destructive/10 hover:text-destructive h-7 rounded-full px-3 text-xs"
											onclick={cancelEdit}
										>
											<X size={12} />
											Cancel
										</Button>
									{:else}
										<Button
											variant="ghost"
											size="sm"
											class="text-muted-foreground hover:text-foreground h-7 rounded-full px-3 text-xs"
											onclick={() => startEdit(item)}
										>
											<Pencil size={12} />
											Edit
										</Button>
										<Button
											variant="ghost"
											size="sm"
											class="text-muted-foreground hover:text-destructive hover:bg-destructive/10 h-7 rounded-full px-3 text-xs"
											disabled={deletingId === item.id}
											onclick={() => deleteItem(item.id)}
										>
											{#if deletingId === item.id}
												<LoaderCircle size={12} class="animate-spin" />
											{:else}
												<Trash2 size={12} />
											{/if}
											{deletingId === item.id ? 'Deleting…' : 'Delete'}
										</Button>
									{/if}
								{/if}

								{#if canViewHistory && !isEditing}
									<Button
										variant="ghost"
										size="sm"
										class="text-muted-foreground hover:text-foreground h-7 rounded-full px-3 text-xs {isHistoryOpen
											? 'bg-muted text-foreground'
											: ''}"
										disabled={historyLoadingId === item.id}
										onclick={() => toggleHistory(item.id)}
									>
										{#if historyLoadingId === item.id}
											<LoaderCircle size={12} class="animate-spin" />
										{:else}
											<History size={12} />
										{/if}
										{isHistoryOpen ? 'Hide History' : 'History'}
									</Button>
								{/if}
							</div>
						{/if}

						<!-- History panel -->
						{#if isHistoryOpen && historyMap[item.id]}
							{@const history = historyMap[item.id]}
							<div class="border-border mt-3 border-t pt-3">
								{#if history.length === 0}
									<p class="text-muted-foreground py-3 text-center text-xs">No history yet.</p>
								{:else}
									<div class="overflow-x-auto">
										<table class="w-full text-xs">
											<thead>
												<tr class="text-muted-foreground border-border border-b">
													<th class="pb-2 text-left font-medium">Action</th>
													<th class="pb-2 text-left font-medium">Name</th>
													<th class="pb-2 text-left font-medium">Price</th>
													<th class="pb-2 text-left font-medium">Amount</th>
													<th class="pb-2 text-left font-medium">By</th>
													<th class="pb-2 text-left font-medium">Date</th>
												</tr>
											</thead>
											<tbody>
												{#each history as h (h.id)}
													{@const ac = actionConfig[h.action]}
													<tr class="border-border border-b last:border-0">
														<td class="py-2 pr-3">
															<Badge variant="outline" class="text-xs {ac.class}">
																{ac.label}
															</Badge>
														</td>
														<td class="text-muted-foreground py-2 pr-3">
															{h.name ?? '—'}
														</td>
														<td class="text-muted-foreground py-2 pr-3">
															{h.price != null ? `$${formatAmount(h.price)}` : '—'}
														</td>
														<td class="text-muted-foreground py-2 pr-3">
															{h.amount ?? '—'}
														</td>

														<td class="text-muted-foreground py-2 pr-3">
															{h.username}#{h.user_id}
														</td>

														<td class="text-muted-foreground py-2 whitespace-nowrap">
															{formatDate(h.changed_at)}
														</td>
													</tr>
												{/each}
											</tbody>
										</table>
									</div>
								{/if}
							</div>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</main>
