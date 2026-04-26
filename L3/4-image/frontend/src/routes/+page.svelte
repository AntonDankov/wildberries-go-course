<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { ImageStatus } from '$lib/types';
	import ImageCard from '$lib/components/ImageCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { RefreshCw, Upload, Check, CircleAlert, CircleCheck, LoaderCircle } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	import { Separator } from '$lib/components/ui/separator/index.js';
	import { Toggle } from '$lib/components/ui/toggle/index.js';

	const pageSizeOptions = [3, 5, 10, 20];
	let pageSize = $state(10);

	let images = $state<ImageStatus[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let currentPage = $state(0);
	let totalPages = $state(1);

	let fileInput = $state<HTMLInputElement | null>(null);
	let selectedFile = $state<File | null>(null);
	let sizeOp = $state<'miniature' | 'resize' | ''>('');
	let useWatermark = $state(false);
	let resizeWidth = $state('');
	let resizeHeight = $state('');
	let uploading = $state(false);
	let uploadError = $state<string | null>(null);
	let uploadSuccess = $state<string | null>(null);
	let dragOver = $state(false);
	let lastUpdated = $state<Date | null>(null);

	const refreshOptions = [
		{ label: '5s', ms: 5_000 },
		{ label: '15s', ms: 15_000 },
		{ label: '30s', ms: 30_000 },
		{ label: '1min', ms: 60_000 },
		{ label: '5min', ms: 300_000 }
	];
	let refreshInterval = $state(5_000);
	let timer: ReturnType<typeof setInterval> | null = null;
	const toggleClass =
		'rounded-full px-4 hover:bg-muted hover:text-muted-foreground data-[state=on]:bg-primary data-[state=on]:text-primary-foreground data-[state=on]:border-primary';

	function startTimer() {
		if (timer) clearInterval(timer);
		timer = setInterval(() => fetchImages(currentPage), refreshInterval);
	}

	function changeInterval(ms: number) {
		refreshInterval = ms;
		startTimer();
	}

	async function fetchImages(p: number) {
		loading = true;
		error = null;
		try {
			const res = await fetch(`http://localhost:8080/image?page=${p}&pageSize=${pageSize}`);
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			const data = await res.json();
			images = data.comments ?? [];
			totalPages = Math.max(1, data.totalPages ?? 1);
			lastUpdated = new Date();
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function goToPage(p: number) {
		currentPage = p;
		await fetchImages(p);
	}

	function handleFileDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
		const f = e.dataTransfer?.files[0];
		if (f) selectedFile = f;
	}

	function handleFileSelect(e: Event) {
		const f = (e.target as HTMLInputElement).files?.[0];
		if (f) selectedFile = f;
	}

	async function uploadImage() {
		if (!selectedFile) return;
		if (sizeOp === 'resize' && (!resizeWidth || !resizeHeight)) {
			uploadError = 'Enter resize dimensions';
			return;
		}
		uploading = true;
		uploadError = null;
		uploadSuccess = null;
		try {
			const form = new FormData();
			form.append('image', selectedFile);
			if (sizeOp === 'miniature') form.append('miniature', 'true');
			else if (sizeOp === 'resize') {
				form.append('resize', 'true');
				form.append('resize_width', resizeWidth);
				form.append('resize_height', resizeHeight);
			}
			if (useWatermark) form.append('watermark', 'true');

			const res = await fetch('http://localhost:8080/upload', { method: 'POST', body: form });
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}
			const data = await res.json();
			uploadSuccess = `Uploaded! ID: ${data.id}`;
			selectedFile = null;
			sizeOp = '';
			useWatermark = false;
			resizeWidth = '';
			resizeHeight = '';
			if (fileInput) {
				fileInput.value = '';
				fileInput.dispatchEvent(new Event('change'));
			}
			await fetchImages(currentPage);
		} catch (e: any) {
			uploadError = e.message;
		} finally {
			uploading = false;
		}
	}

	async function deleteImage(id: string) {
		try {
			const res = await fetch(`http://localhost:8080/image/${id}`, { method: 'DELETE' });
			if (!res.ok) throw new Error(`Server error: ${res.status}`);
			await fetchImages(currentPage);
		} catch (e: any) {
			error = e.message;
		}
	}

	onMount(() => {
		fetchImages(0);
		startTimer();
	});
	onDestroy(() => {
		if (timer) clearInterval(timer);
	});
</script>

<svelte:head><title>Image Processor</title></svelte:head>

<main class="min-h-screen px-4 py-12">
	<div class="mx-auto max-w-2xl">
		<h1 class="mb-8 text-3xl font-bold tracking-tight">Image Processor</h1>

		<section class="bg-card border-border mb-8 flex flex-col gap-4 rounded-xl border p-6">
			<label
				class="border-border text-muted-foreground hover:border-primary hover:bg-primary/5 flex cursor-pointer flex-col items-center gap-2 rounded-lg border-2 border-dashed p-8 transition-colors select-none"
				class:border-emerald-500={!!selectedFile}
				class:border-solid={!!selectedFile}
				class:border-primary={dragOver}
				ondragover={(e) => {
					e.preventDefault();
					dragOver = true;
				}}
				ondragleave={() => (dragOver = false)}
				ondrop={handleFileDrop}
			>
				<input
					type="file"
					accept="image/*"
					class="hidden"
					bind:this={fileInput}
					onchange={handleFileSelect}
				/>

				{#if selectedFile}
					<Check size={22} class="text-emerald-500" />
					<span class="text-foreground text-sm font-medium">{selectedFile.name}</span>
					<span class="text-xs">{(selectedFile.size / 1024).toFixed(1)} KB — click to change</span>
				{:else}
					<Upload size={22} />
					<span class="text-sm font-medium">Drop image here or click to select</span>
					<span class="text-xs">PNG, JPG…</span>
				{/if}
			</label>

			<div class="flex flex-wrap items-center gap-3">
				<span class="text-muted-foreground shrink-0 text-sm">Size</span>
				{#each [['miniature', 'Miniature (150×150)'], ['resize', 'Custom resize']] as [val, label]}
					<Toggle
						variant="outline"
						class={toggleClass}
						pressed={sizeOp === val}
						onPressedChange={() => (sizeOp = sizeOp === val ? '' : (val as typeof sizeOp))}
						>{label}</Toggle
					>
				{/each}
			</div>

			<div class="flex flex-wrap items-center gap-3">
				<span class="text-muted-foreground shrink-0 text-sm">Effects</span>

				<Toggle
					variant="outline"
					class={toggleClass}
					pressed={useWatermark}
					onPressedChange={(v: boolean) => (useWatermark = v)}>Watermark</Toggle
				>
			</div>

			{#if sizeOp === 'resize'}
				<div class="flex items-center gap-2">
					<Input
						type="number"
						placeholder="Width px"
						bind:value={resizeWidth}
						min="1"
						class="w-28 [appearance:textfield]"
					/>
					<span class="text-muted-foreground">*</span>
					<Input
						type="number"
						placeholder="Height px"
						bind:value={resizeHeight}
						min="1"
						class="w-28 [appearance:textfield]"
					/>
				</div>
			{/if}

			{#if uploadError}
				<div class="text-destructive flex items-center gap-1.5 text-sm">
					<CircleAlert size={14} />
					{uploadError}
				</div>
			{/if}
			{#if uploadSuccess}
				<div class="flex items-center gap-1.5 text-sm text-emerald-500">
					<CircleCheck size={14} />
					{uploadSuccess}
				</div>
			{/if}

			<Button
				class="self-end rounded-full"
				disabled={!selectedFile || uploading}
				onclick={uploadImage}
			>
				{#if uploading}<LoaderCircle size={14} class="animate-spin" />{/if}
				{uploading ? 'Uploading…' : 'Upload & Process'}
			</Button>
		</section>

		<Separator class="mb-6" />

		<div class="mb-4 flex flex-wrap items-center justify-between gap-4">
			<h2 class="text-lg font-semibold">Processing Queue</h2>
			<div class="flex flex-wrap items-center justify-end gap-4">
				<Select.Root
					type="single"
					value={String(pageSize)}
					onValueChange={(v) => {
						pageSize = Number(v);
						fetchImages(0);
					}}
				>
					<span class="text-muted-foreground text-xs whitespace-nowrap"> Page size: </span>
					<Select.Trigger class="h-8 w-15 text-xs">
						{pageSize}
					</Select.Trigger>
					<Select.Content>
						{#each pageSizeOptions as size}
							<Select.Item value={String(size)} class="text-xs">{size}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>

				<div class="flex items-center gap-2">
					{#if lastUpdated}
						<span class="text-muted-foreground text-xs whitespace-nowrap">
							Last update time: {lastUpdated.toLocaleTimeString()}
						</span>
					{/if}

					<Select.Root
						type="single"
						value={String(refreshInterval)}
						onValueChange={(v) => changeInterval(Number(v))}
					>
						<Select.Trigger class="h-8 w-15 text-xs">
							{refreshOptions.find((o) => o.ms === refreshInterval)?.label ?? 'Auto'}
						</Select.Trigger>
						<Select.Content>
							{#each refreshOptions as opt}
								<Select.Item value={String(opt.ms)} class="text-xs">{opt.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>

					<Button
						variant="outline"
						size="icon"
						class="size-8"
						onclick={() => fetchImages(currentPage)}
					>
						<RefreshCw size={14} class={loading ? 'animate-spin' : ''} />
					</Button>
				</div>
			</div>
		</div>

		{#if loading && images.length === 0}
			<div class="text-muted-foreground flex flex-col items-center justify-center gap-3 py-16">
				<LoaderCircle size={32} class="animate-spin" />
				<p class="text-sm">Loading…</p>
			</div>
		{:else if error}
			<div class="flex flex-col items-center justify-center gap-3 py-16">
				<p class="text-destructive text-sm">
					<CircleAlert size={14} />
					{error}
				</p>
				<Button
					variant="outline"
					class="border-destructive text-destructive hover:bg-destructive/10"
					onclick={() => fetchImages(currentPage)}
				>
					Retry
				</Button>
			</div>
		{:else if images.length === 0}
			<div class="text-muted-foreground flex items-center justify-center py-16">
				<p class="text-sm">No images yet. Upload one above.</p>
			</div>
		{:else}
			<ul class="flex flex-col gap-2.5">
				{#each images as image (image.id)}
					<ImageCard {image} ondelete={deleteImage} />
				{/each}
			</ul>
		{/if}

		<Pagination {currentPage} {totalPages} onpage={goToPage} />
	</div>
</main>
