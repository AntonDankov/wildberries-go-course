<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { MessageSquare, Plus, X, LoaderCircle, CircleAlert, ChevronLeft } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import Pagination from '$lib/components/Pagination.svelte';

	type Comment = {
		id: number;
		text: string;
		depth: number;
		parent_id: number;
		amount_of_replies: number;
	};

	let commentID = $derived(Number(page.params.commentID));

	let currentPage = $state(Number(page.url.searchParams.get('page') ?? '0'));
	let pageSize = $state(Number(page.url.searchParams.get('pageSize') ?? '10'));
	let maxDepth = $state(Number(page.url.searchParams.get('maxDepth') ?? '3'));

	let rootComment = $state<Comment | null>(null);
	let replies = $state<Comment[]>([]);
	let totalPages = $state(1);
	let loading = $state(false);
	let error = $state<string | null>(null);

	let replyingTo = $state<number | null>(null);
	let replyText = $state('');
	let sending = $state(false);
	let sendError = $state<string | null>(null);

	async function fetchComments(p: number) {
		loading = true;
		error = null;
		try {
			const res = await fetch(
				`http://localhost:8080/comment/${commentID}?page=${p}&pageSize=${pageSize}&maxDepth=${maxDepth}`
			);
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			const data = await res.json();
			const all: Comment[] = data.comments ?? [];
			rootComment = all.length > 0 ? { ...all[0], depth: 0 } : null;
			replies = sortIntoTreeOrder(all.slice(1), rootComment!.id);
			totalPages = Math.max(1, data.totalPages ?? 1);
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	function sortIntoTreeOrder(comments: Comment[], rootId: number): Comment[] {
		const map = new Map<number, Comment[]>();
		for (const c of comments) {
			const list = map.get(c.parent_id) ?? [];
			list.push(c);
			map.set(c.parent_id, list);
		}
		const result: Comment[] = [];
		function dfs(parentId: number) {
			for (const child of map.get(parentId) ?? []) {
				result.push(child);
				dfs(child.id);
			}
		}
		dfs(rootId);
		return result;
	}

	async function goToPage(p: number) {
		if (p < 0 || p >= totalPages) return;
		currentPage = p;
		await goto(`?page=${p}&pageSize=${pageSize}&maxDepth=${maxDepth}`, {
			replaceState: true,
			noScroll: true,
			keepFocus: true
		});
		await fetchComments(p);
	}

	async function applyControls() {
		currentPage = 0;
		await goto(`?page=0&pageSize=${pageSize}&maxDepth=${maxDepth}`, {
			replaceState: true,
			noScroll: true,
			keepFocus: true
		});
		await fetchComments(0);
	}

	async function sendReply(parentID: number) {
		if (!replyText.trim()) return;
		sending = true;
		sendError = null;
		try {
			const res = await fetch('http://localhost:8080/comment', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ text: replyText.trim(), parent_id: parentID })
			});
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			replyText = '';
			replyingTo = null;
			await fetchComments(currentPage);
		} catch (e: any) {
			sendError = e.message;
		} finally {
			sending = false;
		}
	}

	function toggleReply(id: number) {
		replyingTo = replyingTo === id ? null : id;
		replyText = '';
		sendError = null;
	}

	$effect(() => {
		currentPage = 0;
		fetchComments(0);
	});
</script>

<svelte:head><title>Topic #{commentID}</title></svelte:head>

<main class="min-h-screen px-4 py-12">
	<div class="mx-auto max-w-2xl">
		<a
			href="/"
			class="mb-8 flex items-center gap-1.5 text-sm text-muted-foreground no-underline transition-colors hover:text-foreground"
		>
			<ChevronLeft size={14} />Back to topics
		</a>

		{#if loading && !rootComment}
			<div class="flex flex-col items-center gap-3 py-16 text-muted-foreground">
				<LoaderCircle size={32} class="animate-spin" />
				<p class="text-sm">Loading…</p>
			</div>
		{:else if error}
			<div class="flex flex-col items-center gap-3 py-16 text-destructive">
				<div class="flex items-center gap-1.5 text-sm"><CircleAlert size={14} />{error}</div>
				<Button
					variant="outline"
					class="border-destructive text-destructive hover:bg-destructive/10"
					onclick={() => fetchComments(currentPage)}
				>
					Retry
				</Button>
			</div>
		{:else if rootComment}
			<div
				class="mb-6 rounded-2xl border border-border bg-card p-6 shadow-[0_0_40px_rgba(99,102,241,0.06)]"
			>
				<span class="mb-3 block text-xs font-semibold tracking-widest text-primary uppercase">
					Topic #{rootComment.id}
				</span>
				<p class="mb-4 text-lg leading-relaxed break-words text-foreground">{rootComment.text}</p>
				<div class="flex items-center justify-between">
					<span class="flex items-center gap-1.5 text-sm text-muted-foreground">
						<MessageSquare size={15} />{rootComment.amount_of_replies} replies
					</span>
					<Button
						variant="outline"
						size="sm"
						class="rounded-full hover:border-primary hover:text-primary"
						onclick={() => toggleReply(rootComment!.id)}
					>
						{#if replyingTo === rootComment.id}
							<X size={13} />Cancel
						{:else}
							<Plus size={13} />Reply
						{/if}
					</Button>
				</div>

				{#if replyingTo === rootComment.id}
					<div class="mt-4 flex flex-col gap-2 border-t border-border pt-4">
						<Textarea
							placeholder="Write a reply…"
							rows={3}
							bind:value={replyText}
							disabled={sending}
							class="resize-none"
							onkeydown={(e: KeyboardEvent) => {
								if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) sendReply(rootComment!.id);
							}}
						/>
						{#if sendError}
							<div class="flex items-center gap-1.5 text-xs text-destructive">
								<CircleAlert size={12} />{sendError}
							</div>
						{/if}
						<div class="flex justify-end">
							<Button
								size="sm"
								class="rounded-full"
								disabled={sending || !replyText.trim()}
								onclick={() => sendReply(rootComment!.id)}
							>
								{#if sending}<LoaderCircle size={13} class="animate-spin" />{/if}
								{sending ? 'Sending…' : 'Send'}
							</Button>
						</div>
					</div>
				{/if}
			</div>

			<div class="mb-4 flex flex-wrap items-center gap-3">
				<span class="text-sm text-muted-foreground">Page size</span>
				<Select.Root
					type="single"
					value={String(pageSize)}
					onValueChange={(v: any) => (pageSize = Number(v))}
				>
					<Select.Trigger class="h-8 w-24 text-xs">{pageSize}</Select.Trigger>
					<Select.Content>
						{#each [5, 10, 20, 50] as s}
							<Select.Item value={String(s)} class="text-xs">{s}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>

				<span class="text-sm text-muted-foreground">Max depth</span>
				<Select.Root
					type="single"
					value={String(maxDepth)}
					onValueChange={(v: any) => (maxDepth = Number(v))}
				>
					<Select.Trigger class="h-8 w-28 text-xs"
						>{maxDepth === -1 ? 'Unlimited' : maxDepth}</Select.Trigger
					>
					<Select.Content>
						<Select.Item value="-1" class="text-xs">Unlimited</Select.Item>
						{#each [1, 2, 3, 5, 10] as d}
							<Select.Item value={String(d)} class="text-xs">{d}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>

				<Button size="sm" onclick={applyControls}>Apply</Button>
			</div>

			<Separator class="mb-6" />

			{#if loading}
				<div class="flex justify-center py-8 text-muted-foreground">
					<LoaderCircle size={24} class="animate-spin" />
				</div>
			{:else if replies.length === 0}
				<div class="flex justify-center py-16 text-sm text-muted-foreground">No replies yet.</div>
			{:else}
				<ul class="flex flex-col gap-2">
					{#each replies as comment (comment.id)}
						<li class="flex items-center gap-0">
							<div class="flex shrink-0 items-center">
								{#each { length: comment.depth - 1 } as _}
									<span class="flex size-3 items-center justify-center">
										<span class="size-1.5 rounded-full bg-primary opacity-50"></span>
									</span>
								{/each}
							</div>

							<div
								class="ml-2 flex-1 rounded-xl border border-border bg-card p-3 transition-colors hover:border-ring"
							>
								<div class="mb-1 flex items-center gap-2">
									<a
										href="/view/comment/{comment.id}"
										class="text-xs font-semibold text-muted-foreground no-underline transition-colors hover:text-primary"
									>
										#{comment.id}
									</a>
								</div>
								<p class="text-sm leading-relaxed break-words text-muted-foreground">
									{comment.text}
								</p>
								<div class="mt-2 flex items-center justify-between">
									<span class="flex items-center gap-1 text-xs text-muted-foreground">
										<MessageSquare size={12} />{comment.amount_of_replies}
									</span>
									<Button
										variant="ghost"
										size="sm"
										class="rounded-full hover:border-primary hover:text-primary"
										onclick={() => toggleReply(comment.id)}
									>
										{#if replyingTo === comment.id}
											<X size={13} />Cancel
										{:else}
											<Plus size={13} />Reply
										{/if}
									</Button>
								</div>

								{#if replyingTo === comment.id}
									<div class="mt-3 flex flex-col gap-2 border-t border-border pt-3">
										<Textarea
											placeholder="Write a reply…"
											rows={2}
											bind:value={replyText}
											disabled={sending}
											class="resize-none text-sm"
											onkeydown={(e: KeyboardEvent) => {
												if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
													e.preventDefault();
													sendReply(comment.id);
												}
											}}
										/>
										{#if sendError}
											<div class="flex items-center gap-1 text-xs text-destructive">
												<CircleAlert size={11} />{sendError}
											</div>
										{/if}
										<div class="flex justify-end">
											<Button
												size="sm"
												class="rounded-full"
												disabled={sending || !replyText.trim()}
												onclick={() => sendReply(comment.id)}
											>
												{#if sending}<LoaderCircle size={12} class="animate-spin" />{/if}
												{sending ? 'Sending…' : 'Send'}
											</Button>
										</div>
									</div>
								{/if}
							</div>
						</li>
					{/each}
				</ul>
			{/if}

			<Pagination {currentPage} {totalPages} onpage={goToPage} />
		{/if}
	</div>
</main>
