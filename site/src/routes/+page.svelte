<script lang="ts">
	import { base } from '$app/paths';
	import mark from '$lib/assets/logo-icon.png';

	const installMethods = [
		{
			name: 'Homebrew',
			copy: 'brew tap Saku0512/specter https://github.com/Saku0512/specter\nbrew install specter',
			note: 'Fastest path on macOS and Linux if you already live in Homebrew.'
		},
		{
			name: 'curl',
			copy: 'curl -fsSL https://raw.githubusercontent.com/Saku0512/specter/main/install.sh | bash',
			note: 'Single-command install for macOS and Linux.'
		},
		{
			name: 'PowerShell',
			copy: 'irm https://raw.githubusercontent.com/Saku0512/specter/main/install.ps1 | iex',
			note: 'Windows install with a colored installer flow.'
		},
		{
			name: 'Docker',
			copy: 'docker run -v $(pwd)/config.yml:/config.yml ghcr.io/saku0512/specter -c /config.yml',
			note: 'Nice when you want a throwaway mock server without touching your machine.'
		}
	];

	const quickstart = [
		{
			title: 'Generate a config',
			command: 'specter init',
			body: 'Start from a tiny YAML file instead of scaffolding a full backend.'
		},
		{
			title: 'Run the mock API',
			command: 'specter -c config.yml',
			body: 'Boot the server instantly and iterate with hot reload while you edit the file.'
		},
		{
			title: 'Shape real scenarios',
			command: 'curl http://localhost:8080/users',
			body: 'Use state, vars, matching, delays, faults, and the built-in UI to simulate the flow you need.'
		}
	];

	const features = [
		'YAML-defined routes with hot reload',
		'Stateful mocks, vars, stores, rate limits, delays, and faults',
		'OpenAPI generation and validation',
		'Built-in web UI for inspecting and controlling mocks'
	];

	const routeExample = `routes:
  - path: /users
    method: GET
    response:
      - id: 1
        name: Alice

  - path: /login
    method: POST
    set_state: logged_in
    response:
      token: abc123`;

	const terminalExample = `$ specter init
created config.yml

$ specter -c config.yml
registered 2 route(s):
  GET      /users
  POST     /login

UI running on http://localhost:4444`;

	const uiLog = [
		'Requests',
		'Routes',
		'State & Vars',
		'Stores',
		'Auto Refresh On'
	];
</script>

<svelte:head>
	<title>specter | Mock APIs without the ceremony</title>
	<meta
		name="description"
		content="specter is a lightweight mock API server with hot reload, OpenAPI support, stateful flows, and a built-in control room UI."
	/>
</svelte:head>

<div class="page">
	<section class="hero">
		<div class="hero-inner">
			<div class="hero-copy">
				<div class="eyebrow">
					<img src={mark} alt="specter mark" />
					<span>lightweight mock API server</span>
				</div>

				<h1>specter</h1>
				<p class="pronunciation" aria-label="specter pronunciation: SPEK-ter">
					<span>pronounced</span>
					<strong>SPEK-ter</strong>
					<code>/ˈspɛk.tɚ/</code>
					<small>スペクター</small>
				</p>
				<p class="lede">
					Mock APIs without the ceremony. Define routes in YAML, run a single binary, and keep your
					frontend, tests, and demos moving while the real backend catches up.
				</p>

				<div class="cta-row">
					<a class="button primary" href="https://github.com/Saku0512/specter/releases/latest"
						>Download Latest</a
					>
					<a class="button ghost" href={`${base}/docs/`}>Read Docs</a>
					<a class="button ghost" href="https://github.com/Saku0512/specter">View Repository</a>
				</div>

				<ul class="signal-list">
					{#each features as feature}
						<li>{feature}</li>
					{/each}
				</ul>
			</div>

			<div class="hero-visual">
				<div class="scene">
					<div class="terminal">
						<div class="window-bar">
							<span></span><span></span><span></span>
							<strong>terminal</strong>
						</div>
						<pre>{terminalExample}</pre>
					</div>

					<div class="dashboard">
						<div class="window-bar alt">
							<strong>specter control room</strong>
							<span class="status">live</span>
						</div>
						<div class="dashboard-grid">
							<div class="stat">
								<span class="stat-label">Requests</span>
								<strong>24</strong>
							</div>
							<div class="stat">
								<span class="stat-label">Routes</span>
								<strong>12</strong>
							</div>
							<div class="stat">
								<span class="stat-label">Stores</span>
								<strong>3</strong>
							</div>
						</div>
						<div class="ui-list">
							{#each uiLog as item}
								<div>{item}</div>
							{/each}
						</div>
					</div>

					<div class="config-card">
						<div class="window-bar">
							<strong>config.yml</strong>
						</div>
						<pre>{routeExample}</pre>
					</div>
				</div>
			</div>
		</div>
	</section>

	<section class="band" id="download">
		<div class="band-head">
			<div>
				<p class="kicker">Download</p>
				<h2>Pick the install path that matches your workflow</h2>
			</div>
			<a class="button ghost" href="https://github.com/Saku0512/specter/releases/latest"
				>All binaries</a
			>
		</div>

		<div class="install-grid">
			{#each installMethods as method}
				<article class="install-card">
					<div class="card-top">
						<h3>{method.name}</h3>
						<p>{method.note}</p>
					</div>
					<pre>{method.copy}</pre>
				</article>
			{/each}
		</div>
	</section>

	<section class="band alt" id="how-it-works">
		<div class="band-head">
			<div>
				<p class="kicker">How It Works</p>
				<h2>Three moves from blank folder to useful mock API</h2>
			</div>
		</div>

		<div class="steps">
			{#each quickstart as step, index}
				<article class="step">
					<div class="step-number">0{index + 1}</div>
					<h3>{step.title}</h3>
					<pre>{step.command}</pre>
					<p>{step.body}</p>
				</article>
			{/each}
		</div>
	</section>

	<section class="band">
		<div class="explain">
			<div class="explain-copy">
				<p class="kicker">Operate</p>
				<h2>Control the mock without rebuilding your app around it</h2>
				<p>
					specter is built for the awkward middle of product development: frontend ahead of backend,
					demos that need deterministic failures, and tests that want a backend shape without a backend
					team on standby.
				</p>
				<p>
					You can start tiny with one route, then layer in state transitions, OpenAPI validation,
					filters, stores, or the built-in UI as the scenario gets more realistic.
				</p>
			</div>

			<div class="explain-panel">
				<div class="panel-line">
					<span>1</span>
					<div>
						<strong>Edit YAML</strong>
						<p>Routes reload on save, so changing a response is a quick feedback loop.</p>
					</div>
				</div>
				<div class="panel-line">
					<span>2</span>
					<div>
						<strong>Use the control room UI</strong>
						<p>Inspect requests, tweak state, reset stores, and steer scenarios live.</p>
					</div>
				</div>
				<div class="panel-line">
					<span>3</span>
					<div>
						<strong>Ship your frontend anyway</strong>
						<p>Keep development and demos moving while the real API is still settling down.</p>
					</div>
				</div>
			</div>
		</div>
	</section>
</div>

<style>
	:global(body) {
		margin: 0;
		background:
			radial-gradient(circle at 18% 18%, rgba(77, 208, 225, 0.18), transparent 24%),
			radial-gradient(circle at 88% 12%, rgba(120, 255, 180, 0.11), transparent 18%),
			linear-gradient(180deg, #07111f 0%, #0c1523 48%, #0a1220 100%);
		color: #e9f3ff;
		font-family:
			'Avenir Next',
			'Segoe UI',
			'Helvetica Neue',
			Arial,
			sans-serif;
	}

	:global(*) {
		box-sizing: border-box;
	}

	:global(a) {
		color: inherit;
		text-decoration: none;
	}

	.page {
		width: 100%;
		overflow-x: clip;
	}

	.hero {
		min-height: 94svh;
		padding: 2rem clamp(1rem, 2vw, 2rem) 1.25rem;
	}

	.hero-inner,
	.band,
	.explain {
		max-width: 1180px;
		margin: 0 auto;
	}

	.hero-inner {
		display: grid;
		grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr);
		gap: 2rem;
		align-items: center;
		min-height: calc(94svh - 3rem);
	}

	.hero-copy {
		padding: 2rem 0;
	}

	.eyebrow,
	.kicker {
		display: inline-flex;
		align-items: center;
		gap: 0.8rem;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		font-size: 0.77rem;
		color: #8adcee;
	}

	.eyebrow {
		gap: 1rem;
		font-size: 0.9rem;
		font-weight: 700;
	}

	.eyebrow img {
		width: clamp(3.2rem, 6vw, 4.8rem);
		height: clamp(3.2rem, 6vw, 4.8rem);
		animation: logo-float 3.8s ease-in-out infinite;
		filter: drop-shadow(0 18px 22px rgba(77, 208, 225, 0.2));
		will-change: transform;
	}

	@keyframes logo-float {
		0%,
		100% {
			transform: translateY(0);
		}

		50% {
			transform: translateY(-0.55rem);
		}
	}

	h1,
	h2,
	h3 {
		margin: 0;
		font-family:
			'Iowan Old Style',
			'Palatino Linotype',
			'Book Antiqua',
			serif;
		font-weight: 700;
		letter-spacing: 0;
	}

	h1 {
		font-size: clamp(4rem, 10vw, 7rem);
		line-height: 0.92;
		margin-top: 0.85rem;
	}

	.pronunciation {
		display: inline-flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 0.55rem;
		margin: 1rem 0 0;
		padding: 0.58rem 0.72rem;
		border: 1px solid rgba(138, 220, 238, 0.22);
		border-radius: 999px;
		background: rgba(9, 17, 31, 0.32);
		color: #c8d7ee;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
		backdrop-filter: blur(10px);
	}

	.pronunciation span,
	.pronunciation small {
		text-transform: uppercase;
		letter-spacing: 0.1em;
		font-size: 0.72rem;
		color: #8adcee;
	}

	.pronunciation strong {
		color: #f0fbff;
		font-size: clamp(1rem, 1.35vw, 1.16rem);
		letter-spacing: 0.02em;
	}

	.pronunciation code {
		padding: 0.2rem 0.45rem;
		border-radius: 999px;
		background: rgba(138, 241, 255, 0.08);
		color: #dff9ff;
		font-family:
			'SFMono-Regular',
			'JetBrains Mono',
			'IBM Plex Mono',
			Consolas,
			monospace;
		font-size: 0.86rem;
	}

	.lede {
		font-size: clamp(1.05rem, 1.5vw, 1.35rem);
		line-height: 1.7;
		max-width: 36rem;
		color: #c8d7ee;
		margin: 1.3rem 0 0;
	}

	.cta-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.9rem;
		margin-top: 1.7rem;
	}

	.button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-height: 3rem;
		padding: 0.8rem 1.15rem;
		border-radius: 999px;
		border: 1px solid rgba(255, 255, 255, 0.12);
		font-weight: 600;
		transition:
			transform 160ms ease,
			background 160ms ease,
			border-color 160ms ease;
	}

	.button:hover {
		transform: translateY(-1px);
	}

	.button.primary {
		background: linear-gradient(135deg, #8af1ff 0%, #b9ffd2 100%);
		color: #08111f;
		border-color: transparent;
	}

	.button.ghost {
		background: rgba(9, 17, 31, 0.2);
		backdrop-filter: blur(10px);
	}

	.signal-list {
		list-style: none;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.75rem 1rem;
		padding: 0;
		margin: 1.8rem 0 0;
		max-width: 40rem;
	}

	.signal-list li {
		position: relative;
		padding-left: 1.1rem;
		color: #dbe7f6;
	}

	.signal-list li::before {
		content: '';
		position: absolute;
		left: 0;
		top: 0.62rem;
		width: 0.45rem;
		height: 0.45rem;
		border-radius: 999px;
		background: linear-gradient(135deg, #7dd3fc 0%, #a7f3d0 100%);
	}

	.hero-visual {
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 1rem 0 2rem;
	}

	.scene {
		position: relative;
		width: min(100%, 43rem);
		aspect-ratio: 1 / 1.04;
	}

	.terminal,
	.dashboard,
	.config-card,
	.install-card,
	.step,
	.explain-panel {
		border: 1px solid rgba(145, 184, 220, 0.16);
		border-radius: 12px;
		background:
			linear-gradient(180deg, rgba(14, 25, 43, 0.96) 0%, rgba(8, 16, 28, 0.96) 100%);
		box-shadow:
			0 28px 70px rgba(0, 0, 0, 0.28),
			inset 0 1px 0 rgba(255, 255, 255, 0.03);
	}

	.terminal,
	.dashboard,
	.config-card {
		position: absolute;
		overflow: hidden;
	}

	.terminal {
		left: 0;
		top: 0;
		width: 64%;
		transform: rotate(-2deg);
	}

	.dashboard {
		right: 0;
		top: 12%;
		width: 60%;
		transform: rotate(3deg);
	}

	.config-card {
		left: 14%;
		bottom: 0;
		width: 72%;
		transform: rotate(-1deg);
	}

	.window-bar {
		display: flex;
		align-items: center;
		gap: 0.45rem;
		padding: 0.8rem 0.95rem;
		border-bottom: 1px solid rgba(145, 184, 220, 0.11);
		background: rgba(255, 255, 255, 0.03);
		font-size: 0.78rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: #98accc;
	}

	.window-bar span {
		width: 0.58rem;
		height: 0.58rem;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.16);
	}

	.window-bar.alt {
		justify-content: space-between;
	}

	.status {
		color: #baffcf;
	}

	pre {
		margin: 0;
		padding: 1rem 1.05rem 1.1rem;
		font-family:
			'SFMono-Regular',
			'JetBrains Mono',
			'IBM Plex Mono',
			Consolas,
			monospace;
		font-size: 0.84rem;
		line-height: 1.58;
		white-space: pre-wrap;
		word-break: break-word;
		color: #d8e5f7;
	}

	.dashboard-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.8rem;
		padding: 1rem;
	}

	.stat {
		border-radius: 10px;
		padding: 0.9rem;
		background: rgba(125, 211, 252, 0.06);
	}

	.stat-label {
		display: block;
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.11em;
		color: #88a8cb;
	}

	.stat strong {
		display: block;
		font-size: 1.8rem;
		margin-top: 0.35rem;
	}

	.ui-list {
		display: grid;
		gap: 0.55rem;
		padding: 0 1rem 1rem;
	}

	.ui-list div {
		padding: 0.75rem 0.85rem;
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.04);
		color: #d9e8fa;
	}

	.band {
		padding: 1.25rem clamp(1rem, 2vw, 2rem) 1.75rem;
	}

	.band.alt {
		padding-top: 0.25rem;
	}

	.band-head {
		display: flex;
		align-items: end;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.3rem;
	}

	.band-head h2 {
		font-size: clamp(2rem, 4vw, 3.1rem);
		max-width: 14ch;
	}

	.install-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1rem;
	}

	.install-card,
	.step {
		padding: 1.15rem;
	}

	.card-top {
		min-height: 5.9rem;
	}

	.install-card h3,
	.step h3 {
		font-size: 1.45rem;
		margin-bottom: 0.45rem;
	}

	.install-card p,
	.step p,
	.explain-copy p,
	.panel-line p {
		color: #c6d5eb;
		line-height: 1.7;
	}

	.steps {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
	}

	.step-number {
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: #89d8eb;
		margin-bottom: 0.9rem;
	}

	.explain {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 0.94fr);
		gap: 1.25rem;
		align-items: start;
	}

	.explain-copy h2 {
		font-size: clamp(2rem, 4vw, 3.1rem);
		max-width: 12ch;
		margin: 0.3rem 0 1rem;
	}

	.explain-panel {
		padding: 1rem;
		display: grid;
		gap: 0.9rem;
	}

	.panel-line {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 1rem;
		align-items: start;
		padding: 0.95rem;
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.03);
	}

	.panel-line span {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		border-radius: 999px;
		background: linear-gradient(135deg, #8af1ff 0%, #baffcf 100%);
		color: #07111f;
		font-weight: 700;
	}

	.panel-line strong {
		display: block;
		font-size: 1rem;
		margin-bottom: 0.3rem;
	}

	@media (max-width: 980px) {
		.hero-inner,
		.explain,
		.steps,
		.install-grid {
			grid-template-columns: 1fr;
		}

		.hero {
			padding-top: 1rem;
		}

		.hero-inner {
			min-height: auto;
		}

		.scene {
			width: min(100%, 34rem);
			aspect-ratio: 1 / 1.18;
			margin: 0 auto;
		}
	}

	@media (max-width: 720px) {
		.hero {
			min-height: auto;
		}

		.eyebrow {
			align-items: flex-start;
			flex-direction: column;
			gap: 0.75rem;
		}

		.pronunciation {
			border-radius: 14px;
		}

		.signal-list {
			grid-template-columns: 1fr;
		}

		.scene {
			aspect-ratio: 1 / 1.28;
		}

		.terminal {
			width: 84%;
		}

		.dashboard {
			width: 76%;
			top: 23%;
		}

		.config-card {
			left: 6%;
			width: 88%;
		}

		.band-head {
			align-items: start;
			flex-direction: column;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.eyebrow img {
			animation: none;
		}
	}
</style>
