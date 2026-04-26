<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	let commentID = $derived(Number(page.params.commentID));

	$effect(() => {
		const id = commentID;
		currentPage = 0;
		fetchComments(0);
	});

	type Comment = {
		id: number;
		text: string;
		depth: number;
		parent_id: number;
		amount_of_replies: number;
	};

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

			// root is always first (depth 0, omitted by omitempty so we use index)
			rootComment = all.length > 0 ? { ...all[0], depth: 0 } : null;
			const flat = all.slice(1);
			replies = sortIntoTreeOrder(flat, rootComment!.id);
			totalPages = Math.max(1, data.totalPages ?? 1);
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	function sortIntoTreeOrder(comments: Comment[], rootId: number): Comment[] {
		const childrenMap = new Map<number, Comment[]>();
		for (const c of comments) {
			const list = childrenMap.get(c.parent_id) ?? [];
			list.push(c);
			childrenMap.set(c.parent_id, list);
		}

		const result: Comment[] = [];
		function dfs(parentId: number) {
			for (const child of childrenMap.get(parentId) ?? []) {
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
		if (replyingTo === id) {
			replyingTo = null;
			replyText = '';
			sendError = null;
		} else {
			replyingTo = id;
			replyText = '';
			sendError = null;
		}
	}

	function buildPageWindow(cur: number, total: number): number[] {
		const delta = 3;
		const start = Math.max(0, cur - delta);
		const end = Math.min(total - 1, cur + delta);
		const pages: number[] = [];
		for (let i = start; i <= end; i++) pages.push(i);
		return pages;
	}

	let pageWindow = $derived(buildPageWindow(currentPage, totalPages));

	$effect(() => {
		fetchComments(currentPage);
	});
</script>

<svelte:head>
	<title>Topic #{commentID}</title>
</svelte:head>

<main>
	<div class="container">
		<a href="/" class="back-link">← Back to topics</a>

		{#if loading && !rootComment}
			<div class="state-box">
				<div class="spinner"></div>
				<p>Loading…</p>
			</div>
		{:else if error}
			<div class="state-box error">
				<p>⚠ {error}</p>
				<button onclick={() => fetchComments(currentPage)}>Retry</button>
			</div>
		{:else if rootComment}
			<!-- ══ ROOT COMMENT ══ -->
			<div class="root-card">
				<div class="root-label">Topic #{rootComment.id}</div>
				<p class="root-text">{rootComment.text}</p>
				<div class="root-footer">
					<span class="root-replies">
						<svg
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
						</svg>
						{rootComment.amount_of_replies} replies
					</span>
					<button class="reply-btn" onclick={() => toggleReply(rootComment!.id)}>
						<svg
							width="14"
							height="14"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2.5"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							{#if replyingTo === rootComment.id}
								<line x1="18" y1="6" x2="6" y2="18" />
								<line x1="6" y1="6" x2="18" y2="18" />
							{:else}
								<line x1="12" y1="5" x2="12" y2="19" />
								<line x1="5" y1="12" x2="19" y2="12" />
							{/if}
						</svg>
						{replyingTo === rootComment.id ? 'Cancel' : 'Reply'}
					</button>
				</div>

				{#if replyingTo === rootComment.id}
					<div class="inline-reply">
						<textarea
							class="reply-input"
							placeholder="Write a reply…"
							rows="3"
							bind:value={replyText}
							disabled={sending}
							onkeydown={(e) => {
								if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) sendReply(rootComment!.id);
							}}
						></textarea>
						{#if sendError}
							<p class="send-error">⚠ {sendError}</p>
						{/if}
						<div class="reply-actions">
							<button
								class="send-btn"
								disabled={sending || !replyText.trim()}
								onclick={() => sendReply(rootComment!.id)}
							>
								{sending ? 'Sending…' : 'Send'}
							</button>
						</div>
					</div>
				{/if}
			</div>

			<!-- ══ CONTROLS ══ -->
			<div class="controls">
				<label class="control-group">
					<span>Page size</span>
					<select bind:value={pageSize}>
						{#each [5, 10, 20, 50] as s}
							<option value={s}>{s}</option>
						{/each}
					</select>
				</label>
				<label class="control-group">
					<span>Max depth</span>
					<select bind:value={maxDepth}>
						<option value={-1}>Unlimited</option>
						{#each [1, 2, 3, 5, 10] as d}
							<option value={d}>{d}</option>
						{/each}
					</select>
				</label>
				<button class="apply-btn" onclick={applyControls}>Apply</button>
			</div>

			<div class="divider"></div>

			<!-- ══ REPLIES ══ -->
			{#if loading}
				<div class="state-box small">
					<div class="spinner"></div>
				</div>
			{:else if replies.length === 0}
				<div class="state-box"><p>No replies yet.</p></div>
			{:else}
				<ul class="reply-list">
					{#each replies as comment (comment.id)}
						{@const circles = comment.depth - 1}
						{@const indent = (comment.depth - 1) * 20}
						<li class="reply-item">
							<div class="reply-row">
								<div class="depth-dots">
									{#each { length: circles } as _}
										<span class="depth-dot"></span>
									{/each}
								</div>

								<a href="/view/comment/{comment.id}" class="reply-card">
									<div class="reply-meta">
										<span class="reply-id">#{comment.id}</span>
										{#if comment.depth > 1}
											<span class="reply-depth-badge">depth {comment.depth}</span>
										{/if}
									</div>
									<p class="reply-text">{comment.text}</p>
									<div class="reply-footer">
										<span class="action">
											<svg
												width="14"
												height="14"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												stroke-width="2"
												stroke-linecap="round"
												stroke-linejoin="round"
											>
												<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
											</svg>
											{comment.amount_of_replies}
										</span>
										<button
											class="reply-btn small"
											onclick={(e) => {
												e.preventDefault();
												toggleReply(comment.id);
											}}
										>
											<svg
												width="13"
												height="13"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												stroke-width="2.5"
												stroke-linecap="round"
												stroke-linejoin="round"
											>
												{#if replyingTo === comment.id}
													<line x1="18" y1="6" x2="6" y2="18" />
													<line x1="6" y1="6" x2="18" y2="18" />
												{:else}
													<line x1="12" y1="5" x2="12" y2="19" />
													<line x1="5" y1="12" x2="19" y2="12" />
												{/if}
											</svg>
										</button>
									</div>

									{#if replyingTo === comment.id}
										<div class="inline-reply" onclick={(e) => e.preventDefault()}>
											<textarea
												class="reply-input"
												placeholder="Write a reply…"
												rows="2"
												bind:value={replyText}
												disabled={sending}
												onkeydown={(e) => {
													if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
														e.preventDefault();
														sendReply(comment.id);
													}
												}}
											></textarea>
											{#if sendError}
												<p class="send-error">⚠ {sendError}</p>
											{/if}
											<div class="reply-actions">
												<button
													class="send-btn"
													disabled={sending || !replyText.trim()}
													onclick={(e) => {
														e.preventDefault();
														sendReply(comment.id);
													}}
												>
													{sending ? 'Sending…' : 'Send'}
												</button>
											</div>
										</div>
									{/if}
								</a>
							</div>
						</li>
					{/each}
				</ul>
			{/if}

			<!-- ══ PAGINATION ══ -->
			<nav class="pagination">
				<button
					class="page-btn arrow"
					disabled={currentPage === 0}
					onclick={() => goToPage(currentPage - 1)}>←</button
				>

				{#if pageWindow[0] > 0}
					<button class="page-btn" onclick={() => goToPage(0)}>1</button>
					{#if pageWindow[0] > 1}<span class="ellipsis">…</span>{/if}
				{/if}

				{#each pageWindow as p (p)}
					<button class="page-btn" class:active={p === currentPage} onclick={() => goToPage(p)}
						>{p + 1}</button
					>
				{/each}

				{#if pageWindow[pageWindow.length - 1] < totalPages - 1}
					{#if pageWindow[pageWindow.length - 1] < totalPages - 2}
						<span class="ellipsis">…</span>
					{/if}
					<button class="page-btn" onclick={() => goToPage(totalPages - 1)}>{totalPages}</button>
				{/if}

				<button
					class="page-btn arrow"
					disabled={currentPage >= totalPages - 1}
					onclick={() => goToPage(currentPage + 1)}>→</button
				>
			</nav>
		{/if}
	</div>
</main>

<style>
	main {
		min-height: 100vh;
		background: #0f0f13;
		color: #e8e8f0;
		font-family: 'Inter', system-ui, sans-serif;
		padding: 3rem 1rem;
	}

	.container {
		max-width: 680px;
		margin: 0 auto;
	}

	.back-link {
		display: inline-block;
		font-size: 0.875rem;
		color: #6b7280;
		text-decoration: none;
		margin-bottom: 2rem;
		transition: color 0.15s;
	}
	.back-link:hover {
		color: #fff;
	}

	/* ── Root comment ── */
	.root-card {
		background: linear-gradient(135deg, #1a1a2e, #16162a);
		border: 1px solid #3b3b58;
		border-radius: 16px;
		padding: 1.75rem;
		margin-bottom: 1.5rem;
		box-shadow: 0 0 40px rgba(99, 102, 241, 0.08);
	}

	.root-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: #6366f1;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		margin-bottom: 0.75rem;
	}

	.root-text {
		font-size: 1.15rem;
		line-height: 1.7;
		color: #f0f0fa;
		margin: 0 0 1.25rem;
		word-break: break-word;
	}

	.root-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.root-replies {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		font-size: 0.85rem;
		color: #6b7280;
	}

	/* ── Controls ── */
	.controls {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
		margin-bottom: 1.5rem;
	}

	.control-group {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.85rem;
		color: #6b7280;
	}

	.control-group select {
		background: #1a1a24;
		border: 1px solid #2a2a38;
		border-radius: 8px;
		color: #e8e8f0;
		padding: 0.3rem 0.6rem;
		font-size: 0.85rem;
		cursor: pointer;
		outline: none;
	}

	.apply-btn {
		padding: 0.35rem 1rem;
		background: #6366f1;
		border: none;
		border-radius: 8px;
		color: #fff;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.15s;
	}
	.apply-btn:hover {
		background: #4f46e5;
	}

	.divider {
		height: 1px;
		background: #2a2a38;
		margin-bottom: 1.5rem;
	}

	/* ── Reply list ── */
	.reply-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.reply-row {
		display: flex;
		align-items: center;
		gap: 0;
	}

	.depth-dots {
		display: flex;
		flex-direction: row;
		align-items: center;
		gap: 0;
		margin-right: 4px;
		flex-shrink: 0;
	}

	.depth-dot {
		width: 16px;
		height: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.depth-dot::after {
		content: '';
		width: 5px;
		height: 5px;
		border-radius: 50%;
		background: #6366f1;
		opacity: 0.5;
	}
	/* ── Reply card ── */
	.reply-card {
		flex: 1;
		background: #1a1a24;
		border: 1px solid #2a2a38;
		border-radius: 10px;
		padding: 0.85rem 1rem;
		transition: border-color 0.15s;
	}

	.reply-card:hover {
		border-color: #3b3b58;
	}

	.reply-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.4rem;
	}

	.reply-id {
		font-size: 0.8rem;
		font-weight: 600;
		color: #9ca3af;
		text-decoration: none;
		transition: color 0.15s;
	}

	.reply-id:hover {
		color: #6366f1;
	}

	.reply-depth-badge {
		font-size: 0.7rem;
		color: #6366f1;
		background: rgba(99, 102, 241, 0.1);
		padding: 0.1rem 0.45rem;
		border-radius: 999px;
	}

	.reply-text {
		margin: 0 0 0.6rem;
		font-size: 0.93rem;
		line-height: 1.55;
		color: #d1d5db;
		word-break: break-word;
	}

	.reply-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.action {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		font-size: 0.8rem;
		color: #6b7280;
	}

	/* ── Reply button ── */
	.reply-btn {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		background: transparent;
		border: 1px solid #2a2a38;
		border-radius: 999px;
		color: #6b7280;
		font-size: 0.82rem;
		padding: 0.3rem 0.75rem;
		cursor: pointer;
		transition: all 0.15s;
	}

	.reply-btn:hover {
		border-color: #6366f1;
		color: #6366f1;
	}

	.reply-btn.small {
		padding: 0.25rem 0.5rem;
	}

	/* ── Inline reply form ── */
	.inline-reply {
		margin-top: 0.75rem;
		border-top: 1px solid #2a2a38;
		padding-top: 0.75rem;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.reply-input {
		width: 100%;
		background: #0f0f13;
		border: 1px solid #2a2a38;
		border-radius: 8px;
		padding: 0.6rem 0.8rem;
		color: #e8e8f0;
		font-size: 0.9rem;
		font-family: inherit;
		line-height: 1.5;
		resize: none;
		outline: none;
		transition: border-color 0.15s;
		box-sizing: border-box;
	}

	.reply-input:focus {
		border-color: #6366f1;
	}
	.reply-input::placeholder {
		color: #4b5563;
	}
	.reply-input:disabled {
		opacity: 0.5;
	}

	.reply-actions {
		display: flex;
		justify-content: flex-end;
	}

	.send-btn {
		padding: 0.4rem 1.1rem;
		background: #6366f1;
		border: none;
		border-radius: 999px;
		color: #fff;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		transition:
			background 0.15s,
			opacity 0.15s;
	}

	.send-btn:hover:not(:disabled) {
		background: #4f46e5;
	}
	.send-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.send-error {
		font-size: 0.82rem;
		color: #f87171;
		margin: 0;
	}

	/* ── States ── */
	.state-box {
		text-align: center;
		padding: 4rem 1rem;
		color: #6b7280;
	}

	.state-box.small {
		padding: 2rem;
	}
	.state-box.error {
		color: #f87171;
	}

	.state-box button {
		margin-top: 1rem;
		padding: 0.5rem 1.2rem;
		border: 1px solid #f87171;
		border-radius: 8px;
		background: transparent;
		color: #f87171;
		cursor: pointer;
	}

	/* ── Spinner ── */
	.spinner {
		width: 32px;
		height: 32px;
		border: 3px solid #2a2a38;
		border-top-color: #6366f1;
		border-radius: 50%;
		animation: spin 0.7s linear infinite;
		margin: 0 auto 1rem;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	/* ── Pagination ── */
	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 0.35rem;
		margin-top: 2.5rem;
		flex-wrap: wrap;
	}

	.page-btn {
		min-width: 38px;
		height: 38px;
		padding: 0 0.6rem;
		border: 1px solid #2a2a38;
		border-radius: 8px;
		background: #1a1a24;
		color: #9ca3af;
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.15s;
	}

	.page-btn:hover:not(:disabled) {
		border-color: #6366f1;
		color: #fff;
	}

	.page-btn.active {
		background: #6366f1;
		border-color: #6366f1;
		color: #fff;
		font-weight: 600;
	}

	.page-btn:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}
	.page-btn.arrow {
		font-size: 1rem;
	}

	.ellipsis {
		color: #6b7280;
		padding: 0 0.2rem;
		line-height: 38px;
	}
</style>
