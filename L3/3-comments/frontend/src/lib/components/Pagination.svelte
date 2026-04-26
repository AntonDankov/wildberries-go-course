<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination/index.js';

	type Props = { currentPage: number; totalPages: number; onpage: (p: number) => void };
	let { currentPage, totalPages, onpage }: Props = $props();
</script>

<Pagination.Root
	class="mt-8"
	count={totalPages}
	perPage={1}
	page={currentPage + 1}
	onPageChange={(p: number) => onpage(p - 1)}
>
	{#snippet children({ pages, currentPage: activePage }: any)}
		<Pagination.Content>
			<Pagination.Item>
				<Pagination.Previous />
			</Pagination.Item>

			{#each pages as page (page.key)}
				<Pagination.Item>
					{#if page.type === 'ellipsis'}
						<Pagination.Ellipsis />
					{:else}
						<Pagination.Link {page} isActive={activePage === page.value}>
							{page.value}
						</Pagination.Link>
					{/if}
				</Pagination.Item>
			{/each}

			<Pagination.Item>
				<Pagination.Next />
			</Pagination.Item>
		</Pagination.Content>
	{/snippet}
</Pagination.Root>
