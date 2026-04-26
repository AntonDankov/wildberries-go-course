<script lang="ts">
	import { goto } from '$app/navigation';
	import { onDestroy, onMount } from 'svelte';
	import {
		CircleAlert,
		CircleCheck,
		LoaderCircle,
		Warehouse,
		Eye,
		EyeOff,
		LogIn,
		UserPlus
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	const roleOptions = [
		{ label: 'Viewer', value: '1' },
		{ label: 'Owner', value: '2' },
		{ label: 'Manager', value: '4' },
		{ label: 'Admin', value: '8' }
	];

	let loginName = $state('');
	let loginPassword = $state('');
	let loginLoading = $state(false);
	let loginError = $state<string | null>(null);
	let showLoginPassword = $state(false);

	let regName = $state('');
	let regPassword = $state('');
	let regRole = $state('1');
	let regLoading = $state(false);
	let regError = $state<string | null>(null);
	let regSuccess = $state<string | null>(null);
	let showRegPassword = $state(false);

	let countdown = $state<number | null>(null);
	let countdownTimer: ReturnType<typeof setInterval> | null = null;

	onMount(() => {
		const token = localStorage.getItem('token');
		if (token) goto('/items');
	});

	onDestroy(() => {
		if (countdownTimer) clearInterval(countdownTimer);
	});

	async function login() {
		if (!loginName || !loginPassword) {
			loginError = 'Name and password are required';
			return;
		}

		loginLoading = true;
		loginError = null;

		try {
			const res = await fetch('http://localhost:8080/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: loginName, password: loginPassword })
			});

			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}

			const data = await res.json();
			localStorage.setItem('token', data.token);
			localStorage.setItem('user_id', String(data.user_id));
			localStorage.setItem('name', data.name);
			localStorage.setItem('role', String(data.role));
			await goto('/items');
		} catch (e: any) {
			loginError = e.message;
		} finally {
			loginLoading = false;
		}
	}

	async function register() {
		if (!regName || !regPassword) {
			regError = 'Name and password are required';
			return;
		}
		if (regName.length < 3) {
			regError = 'Name must be at least 3 characters';
			return;
		}
		if (regPassword.length < 6) {
			regError = 'Password must be at least 6 characters';
			return;
		}

		regLoading = true;
		regError = null;
		regSuccess = null;

		try {
			const res = await fetch('http://localhost:8080/user', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: regName,
					password: regPassword,
					role: Number(regRole)
				})
			});

			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error ?? `Server error: ${res.status}`);
			}

			regSuccess = 'Account created! Signing in…';
			const createdName = regName;
			const createdPassword = regPassword;
			regName = '';
			regPassword = '';
			regRole = '1';

			countdown = 2;
			countdownTimer = setInterval(async () => {
				countdown = (countdown ?? 1) - 1;
				if (countdown <= 0) {
					clearInterval(countdownTimer!);
					countdownTimer = null;
					countdown = null;
					loginName = createdName;
					loginPassword = createdPassword;
					await login();
				}
			}, 1_000);
		} catch (e: any) {
			regError = e.message;
		} finally {
			regLoading = false;
		}
	}
</script>

<svelte:head><title>Warehouse - Auth</title></svelte:head>

<main class="flex min-h-screen items-center justify-center px-4 py-12">
	<div class="w-full max-w-sm">
		<div class="mb-8 flex flex-col items-center gap-2">
			<div class="bg-card border-border flex size-12 items-center justify-center rounded-xl border">
				<Warehouse size={22} />
			</div>
			<h1 class="text-xl font-bold tracking-tight">Warehouse Control</h1>
		</div>

		<Tabs.Root value="login">
			<Tabs.List class="mb-6 w-full">
				<Tabs.Trigger value="login" class="flex-1">
					<LogIn size={13} />
					Sign in
				</Tabs.Trigger>
				<Tabs.Trigger value="register" class="flex-1">
					<UserPlus size={13} />
					Register
				</Tabs.Trigger>
			</Tabs.List>

			<Tabs.Content value="login">
				<section class="bg-card border-border flex flex-col gap-4 rounded-xl border p-6">
					<div class="flex flex-col gap-3">
						<Input placeholder="Username" bind:value={loginName} autocomplete="username" />
						<div class="relative">
							<Input
								type={showLoginPassword ? 'text' : 'password'}
								placeholder="Password"
								bind:value={loginPassword}
								autocomplete="current-password"
								class="pr-9"
							/>
							<button
								type="button"
								class="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2 transition-colors"
								onclick={() => (showLoginPassword = !showLoginPassword)}
							>
								{#if showLoginPassword}
									<EyeOff size={14} />
								{:else}
									<Eye size={14} />
								{/if}
							</button>
						</div>
					</div>

					{#if loginError}
						<div class="text-destructive flex items-center gap-1.5 text-sm">
							<CircleAlert size={14} />
							{loginError}
						</div>
					{/if}

					<Button
						class="w-full rounded-full"
						disabled={!loginName || !loginPassword || loginLoading}
						onclick={login}
					>
						{#if loginLoading}
							<LoaderCircle size={14} class="animate-spin" />
							Signing in…
						{:else}
							<LogIn size={14} />
							Sign in
						{/if}
					</Button>
				</section>
			</Tabs.Content>

			<Tabs.Content value="register">
				<section class="bg-card border-border flex flex-col gap-4 rounded-xl border p-6">
					<div class="flex flex-col gap-3">
						<Input
							placeholder="Username (min 3 chars)"
							bind:value={regName}
							autocomplete="username"
						/>
						<div class="relative">
							<Input
								type={showRegPassword ? 'text' : 'password'}
								placeholder="Password (min 6 chars)"
								bind:value={regPassword}
								autocomplete="new-password"
								class="pr-9"
							/>
							<button
								type="button"
								class="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2 transition-colors"
								onclick={() => (showRegPassword = !showRegPassword)}
							>
								{#if showRegPassword}
									<EyeOff size={14} />
								{:else}
									<Eye size={14} />
								{/if}
							</button>
						</div>
						<Select.Root type="single" value={regRole} onValueChange={(v) => (regRole = v)}>
							<Select.Trigger class="text-sm">
								{roleOptions.find((r) => r.value === regRole)?.label ?? 'Select role'}
							</Select.Trigger>
							<Select.Content>
								{#each roleOptions as opt}
									<Select.Item value={opt.value} class="text-sm">{opt.label}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>

					{#if regError}
						<div class="text-destructive flex items-center gap-1.5 text-sm">
							<CircleAlert size={14} />
							{regError}
						</div>
					{/if}

					{#if regSuccess}
						<div class="flex items-center gap-1.5 text-sm text-emerald-500">
							<CircleCheck size={14} />
							{regSuccess}
							{#if countdown !== null}
								<span class="text-muted-foreground ml-auto font-mono">{countdown}s</span>
							{/if}
						</div>
					{/if}

					<Button
						class="w-full rounded-full"
						disabled={!regName || !regPassword || regLoading}
						onclick={register}
					>
						{#if regLoading}
							<LoaderCircle size={14} class="animate-spin" />
							Creating account…
						{:else}
							<UserPlus size={14} />
							Create account
						{/if}
					</Button>
				</section>
			</Tabs.Content>
		</Tabs.Root>
	</div>
</main>
