<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { MessageSquare, Trash2, LoaderCircle, CircleAlert, Search } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import Pagination from '$lib/components/Pagination.svelte';

	const PAGE_SIZE = 3;

	type Comment = {
		id: number;
		text: string;
		depth: number;
		parent_id: number;
		amount_of_replies: number;
	};

	let activeTab = $state<'topics' | 'search'>(
		(page.url.searchParams.get('tab') as 'topics' | 'search') ?? 'topics'
	);

	let currentPage = $state(Number(page.url.searchParams.get('page') ?? '0'));
	let loading = $state(false);
	let error = $state<string | null>(null);
	let totalPages = $state(1);
	let comments = $state<Comment[]>([]);
	let newTopicText = $state('');
	let creating = $state(false);
	let createError = $state<string | null>(null);

	let searchText = $state(page.url.searchParams.get('q') ?? '');
	let searchResults = $state<Comment[]>([]);
	let searchLoading = $state(false);
	let searchError = $state<string | null>(null);
	let searchPage = $state(Number(page.url.searchParams.get('searchPage') ?? '0'));
	let searchTotalPages = $state(1);
	let searchDone = $state(false);

	async function fetchTopics(p: number) {
		loading = true;
		error = null;
		try {
			const res = await fetch(`http://localhost:8080/comment?page=${p}&pageSize=${PAGE_SIZE}`);
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			const data = await res.json();
			comments = data.comments ?? [];
			totalPages = Math.max(1, data.totalPages ?? 1);
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function goToPage(p: number) {
		if (p < 0 || p >= totalPages) return;
		currentPage = p;
		await goto(`?tab=topics&page=${p}`, { replaceState: true, noScroll: true, keepFocus: true });
		await fetchTopics(p);
	}

	async function createTopic() {
		if (!newTopicText.trim()) return;
		creating = true;
		createError = null;
		try {
			const res = await fetch('http://localhost:8080/comment', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ text: newTopicText.trim(), parent_id: -1 })
			});
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			newTopicText = '';
			await fetchTopics(currentPage);
		} catch (e: any) {
			createError = e.message;
		} finally {
			creating = false;
		}
	}

	async function deleteTopic(id: number) {
		try {
			const res = await fetch(`http://localhost:8080/comment/${id}`, { method: 'DELETE' });
			if (!res.ok) throw new Error(`Server error: ${res.status}`);

			const nextPage = comments.length === 1 && currentPage > 0 ? currentPage - 1 : currentPage;
			await goToPage(nextPage);
		} catch (e: any) {
			error = e.message;
		}
	}

	async function doSearch(p: number) {
		if (!searchText.trim()) return;
		searchLoading = true;
		searchError = null;
		searchDone = false;
		try {
			const res = await fetch(
				`http://localhost:8080/comment/search?searchText=${encodeURIComponent(searchText.trim())}&page=${p}&pageSize=${PAGE_SIZE}`
			);
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			const data = await res.json();
			searchResults = data.comments ?? [];
			searchTotalPages = Math.max(1, data.totalPages ?? 1);
			searchDone = true;
		} catch (e: any) {
			searchError = e.message;
		} finally {
			searchLoading = false;
		}
	}

	async function goToSearchPage(p: number) {
		if (p < 0 || p >= searchTotalPages) return;
		searchPage = p;
		await goto(`?tab=search&q=${encodeURIComponent(searchText)}&searchPage=${p}`, {
			replaceState: true,
			noScroll: true,
			keepFocus: true
		});
		await doSearch(p);
	}

	async function switchTab(tab: 'topics' | 'search') {
		activeTab = tab;
		await goto(tab === 'topics' ? `?tab=topics&page=${currentPage}` : `?tab=search`, {
			replaceState: true,
			noScroll: true,
			keepFocus: true
		});
	}

	$effect(() => {
		fetchTopics(currentPage);
	});
</script>

<svelte:head><title>Topico</title></svelte:head>

<main class="min-h-screen px-4 py-12">
	<div class="mx-auto max-w-2xl">
		<div class="mb-6 flex items-baseline gap-3">
			<h1 class="text-3xl font-bold tracking-tight">Topico</h1>
			<p class="text-sm text-muted-foreground">Threads killer</p>
		</div>

		<Tabs.Root value={activeTab} onValueChange={(v: any) => switchTab(v as 'topics' | 'search')}>
			<Tabs.List class="mb-6 w-full">
				<Tabs.Trigger value="topics" class="flex-1">Topics</Tabs.Trigger>
				<Tabs.Trigger value="search" class="flex-1">Search</Tabs.Trigger>
			</Tabs.List>

			<Tabs.Content value="topics">
				<div class="mb-8 flex flex-col gap-3 rounded-xl border border-border bg-card p-4">
					<Textarea
						placeholder="It's a place to write some writing"
						rows={3}
						bind:value={newTopicText}
						disabled={creating}
						class="resize-none border-none bg-transparent shadow-none focus-visible:ring-0"
						onkeydown={(e: KeyboardEvent) => {
							if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) createTopic();
						}}
					/>

					{#if createError}
						<div class="flex items-center gap-1.5 text-sm text-destructive">
							<CircleAlert size={14} />{createError}
						</div>
					{/if}

					<div class="flex items-center justify-end gap-3 border-t border-border pt-3">
						<span
							class="text-sm {newTopicText.length > 240
								? 'text-destructive'
								: 'text-muted-foreground'}"
						>
							{newTopicText.length}
						</span>
						<Button
							class="rounded-full"
							disabled={creating || !newTopicText.trim()}
							onclick={createTopic}
						>
							{#if creating}<LoaderCircle size={14} class="animate-spin" />{/if}
							{creating ? 'Posting…' : 'Post'}
						</Button>
					</div>
				</div>

				{#if loading}
					<div class="flex flex-col items-center justify-center gap-3 py-16 text-muted-foreground">
						<LoaderCircle size={32} class="animate-spin" />
						<p class="text-sm">Loading topics…</p>
					</div>
				{:else if error}
					<div class="flex flex-col items-center gap-3 py-16 text-destructive">
						<div class="flex items-center gap-1.5 text-sm"><CircleAlert size={14} />{error}</div>
						<Button
							variant="outline"
							class="border-destructive text-destructive hover:bg-destructive/10"
							onclick={() => fetchTopics(currentPage)}
						>
							Retry
						</Button>
					</div>
				{:else if comments.length === 0}
					<div class="flex justify-center py-16 text-sm text-muted-foreground">No topics yet.</div>
				{:else}
					<ul class="flex flex-col gap-3">
						{#each comments as comment (comment.id)}
							<li
								class="group rounded-xl border border-border bg-card transition-colors hover:border-primary"
							>
								<a href="/view/comment/{comment.id}" class="flex flex-col gap-1.5 p-4 no-underline">
									<span class="text-sm font-semibold">Topic #{comment.id}</span>
									<p class="m-0 text-sm leading-relaxed break-words text-muted-foreground">
										{comment.text}
									</p>
									<div class="mt-1 flex items-center justify-between">
										<span
											class="flex items-center gap-1.5 text-xs text-muted-foreground transition-colors group-hover:text-primary"
										>
											<MessageSquare size={14} />{comment.amount_of_replies}
										</span>
										<Button
											variant="ghost"
											size="icon"
											class="size-7 hover:text-destructive"
											onclick={(e: MouseEvent) => {
												e.preventDefault();
												deleteTopic(comment.id);
											}}
										>
											<Trash2 size={14} />
										</Button>
									</div>
								</a>
							</li>
						{/each}
					</ul>
					<Pagination {currentPage} {totalPages} onpage={goToPage} />
				{/if}
			</Tabs.Content>

			<Tabs.Content value="search">
				<div class="mb-6 flex gap-3">
					<Input
						type="text"
						placeholder="Search topics…"
						bind:value={searchText}
						onkeydown={(e: KeyboardEvent) => {
							if (e.key === 'Enter') {
								searchPage = 0;
								doSearch(0);
							}
						}}
					/>
					<Button
						class="rounded-full"
						disabled={searchLoading || !searchText.trim()}
						onclick={() => {
							searchPage = 0;
							doSearch(0);
						}}
					>
						{#if searchLoading}
							<LoaderCircle size={14} class="animate-spin" />
						{:else}
							<Search size={14} />
						{/if}
						{searchLoading ? 'Searching…' : 'Search'}
					</Button>
				</div>

				{#if searchLoading}
					<div class="flex flex-col items-center gap-3 py-16 text-muted-foreground">
						<LoaderCircle size={32} class="animate-spin" />
						<p class="text-sm">Searching…</p>
					</div>
				{:else if searchError}
					<div class="flex items-center justify-center gap-1.5 py-16 text-sm text-destructive">
						<CircleAlert size={14} />{searchError}
					</div>
				{:else if searchDone && searchResults.length === 0}
					<div class="flex justify-center py-16 text-sm text-muted-foreground">
						No results for "<strong>{searchText}</strong>".
					</div>
				{:else if searchResults.length > 0}
					<ul class="flex flex-col gap-3">
						{#each searchResults as comment (comment.id)}
							<li
								class="group rounded-xl border border-border bg-card transition-colors hover:border-primary"
							>
								<a href="/view/comment/{comment.id}" class="flex flex-col gap-1.5 p-4 no-underline">
									<span class="text-sm font-semibold">Topic #{comment.id}</span>
									<p class="m-0 text-sm leading-relaxed break-words text-muted-foreground">
										{comment.text}
									</p>
									<span
										class="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground transition-colors group-hover:text-primary"
									>
										<MessageSquare size={14} />{comment.amount_of_replies}
									</span>
								</a>
							</li>
						{/each}
					</ul>
					<Pagination
						currentPage={searchPage}
						totalPages={searchTotalPages}
						onpage={goToSearchPage}
					/>
				{/if}
			</Tabs.Content>
		</Tabs.Root>
	</div>
</main>
