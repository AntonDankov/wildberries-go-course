<script lang="ts">
	import type { ImageStatus } from '$lib/types';
	import { Trash2, Download, ImageIcon } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';

	type StatusConfig = { label: string; color: string; bg: string };
	const STATUS: Record<number, StatusConfig> = {
		0: { label: 'Waiting', color: '#f59e0b', bg: 'rgba(245,158,11,0.1)' },
		1: { label: 'Processed', color: '#10b981', bg: 'rgba(16,185,129,0.1)' },
		2: { label: 'Deleted', color: '#6b7280', bg: 'rgba(107,114,128,0.1)' },
		3: { label: 'Failed', color: '#f87171', bg: 'rgba(248,113,113,0.1)' }
	};

	let { image, ondelete }: { image: ImageStatus; ondelete: (id: string) => void } = $props();

	const status = $derived(STATUS[image.process_type] ?? STATUS[0]);
	const isProcessed = $derived(image.process_type === 1);
	const isWaiting = $derived(image.process_type === 0);
	const url = $derived(`http://localhost:8080/image/${image.id}`);
</script>

<li
	class="border-border bg-card hover:bg-accent/5 hover:border-ring flex items-center gap-4 rounded-lg border p-3 transition-colors"
>
	<div
		class="border-border bg-background flex size-14 shrink-0 items-center justify-center overflow-hidden rounded-md border"
	>
		{#if isProcessed}
			<img src={url} alt={image.id} class="size-full object-cover" />
		{:else}
			<ImageIcon size={24} class="text-muted-foreground/30" />
		{/if}
	</div>

	<div class="flex min-w-0 flex-1 flex-col gap-1">
		<div class="flex items-center gap-2">
			<span class="text-muted-foreground truncate text-sm font-semibold" title={image.id}>
				{image.id.slice(0, 12)}…
			</span>
			<Badge variant="outline" class="text-primary border-primary/30 bg-primary/10 shrink-0">
				.{image.extension}
			</Badge>
		</div>

		<Badge
			variant="outline"
			class="w-fit gap-1.5"
			style="color:{status.color}; background:{status.bg}; border-color:{status.color}33"
		>
			{#if isWaiting}
				<span class="size-1.5 animate-pulse rounded-full" style:background={status.color}></span>
			{/if}
			{status.label}
		</Badge>
	</div>

	<div class="flex shrink-0 items-center gap-1.5">
		<Button
			variant="outline"
			size="sm"
			href={isProcessed ? url : undefined}
			target="_blank"
			disabled={!isProcessed}
			title={isProcessed ? 'Download' : 'Not processed yet'}
		>
			<Download size={13} />Download
		</Button>
		<Button
			variant="outline"
			size="icon"
			onclick={() => ondelete(image.id)}
			class="hover:border-destructive hover:text-destructive size-8"
		>
			<Trash2 size={13} />
		</Button>
	</div>
</li>
