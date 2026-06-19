<script lang="ts">
	import { base } from '$app/paths';
	import LanguageToggle from '$lib/LanguageToggle.svelte';
	import { language } from '$lib/language';
	import mark from '$lib/assets/logo-icon.png';

	const copy = {
		ja: {
			title: 'specter | 手軽に使えるモック API サーバー',
			description:
				'specter は hot reload、OpenAPI、状態管理、組み込み UI を備えた軽量なモック API サーバーです。',
			navDocs: 'Docs',
			navRepo: 'GitHub',
			eyebrow: '軽量モック API サーバー',
			pronounced: 'pronounced',
			lede:
				'手間の少ないモック API。YAML でルートを定義し、単一バイナリを起動するだけで、実バックエンドの完成を待たずにフロントエンド、テスト、デモを進められます。',
			downloadLatest: '最新版をダウンロード',
			readDocs: 'ドキュメントを読む',
			viewRepository: 'リポジトリを見る',
			downloadKicker: 'Download',
			downloadTitle: '自分のワークフローに合うインストール方法を選ぶ',
			allBinaries: 'すべてのバイナリ',
			copyCommand: 'コピー',
			copiedCommand: 'コピー済み',
			copyCommandLabel: 'インストールコマンドをコピー',
			howKicker: 'How It Works',
			howTitle: '空のフォルダから実用的なモック API までの 3 ステップ',
			operateKicker: 'Operate',
			operateTitle: 'アプリを作り直さずにモックを操作する',
			operateBody:
				'specter は、フロントエンドがバックエンドより先行している時期、決まった失敗を再現したいデモ、実バックエンドなしで API 形状を使いたいテストのために作られています。',
			operateBody2:
				'まずは 1 つのルートから始め、必要に応じて状態遷移、OpenAPI 検証、フィルタ、ストア、組み込み UI を重ねて現実的なシナリオにできます。',
			useCasesKicker: 'Use Cases',
			useCasesTitle: 'バックエンド待ちの時間を、開発できる時間に変える',
			useCases: [
				{
					title: 'フロントエンド開発',
					body: 'API の実装前でも、画面、フォーム、error handling、empty state を作り込めます。'
				},
				{
					title: 'E2E テスト',
					body: '状態、失敗、rate limit、遅延を固定し、再現性のあるテスト環境を作れます。'
				},
				{
					title: 'デモとプロトタイプ',
					body: '実 backend の準備を待たず、決まったデータや失敗パターンで demo flow を確認できます。'
				},
				{
					title: 'OpenAPI mock',
					body: 'OpenAPI から config を生成し、request / response validation で API contract のずれを見つけられます。'
				},
				{
					title: 'Retry / failure simulation',
					body: 'sequential responses、on_call、error_rate、delay を使って retry や timeout を試せます。'
				},
				{
					title: 'Local integration',
					body: '一部だけ real API に proxy し、それ以外を mock する hybrid な開発環境を作れます。'
				}
			],
			securityKicker: 'Security',
			securityTitle: 'ローカル開発向けのツールです',
			securityBody:
				'specter は信頼できるネットワークで使う local development tool です。追加の firewall、reverse proxy、authentication なしに public internet へ公開しないでください。',
			securityLink: 'Security Policy を読む',
			projectKicker: 'Project',
			projectTitle: 'プロジェクト情報',
			projectBody:
				'Security、License、Releases、Roadmap は GitHub で管理しています。変更提案や不具合報告も歓迎です。',
			projectLinks: [
				{ label: 'Documentation', href: `${base}/docs/` },
				{ label: 'Security Policy', href: 'https://github.com/Saku0512/specter/security/policy' },
				{ label: 'License', href: 'https://github.com/Saku0512/specter/blob/main/LICENSE' },
				{ label: 'Releases / Changelog', href: 'https://github.com/Saku0512/specter/releases' },
				{ label: 'Roadmap / Issues', href: 'https://github.com/Saku0512/specter/issues' },
				{ label: 'GitHub', href: 'https://github.com/Saku0512/specter' }
			],
			panel: [
				{
					title: 'YAML を編集',
					body: '保存するとルートが再読み込みされるため、レスポンス変更をすぐ確認できます。'
				},
				{
					title: 'コントロール UI を使う',
					body: 'リクエスト確認、状態変更、ストア初期化、シナリオ操作をブラウザから行えます。'
				},
				{
					title: 'フロントエンド開発を止めない',
					body: '本物の API が固まり切っていない間も、開発やデモを前に進められます。'
				}
			],
			features: [
				'YAML で定義するルートと hot reload',
				'状態管理、vars、stores、rate limit、delay、fault',
				'OpenAPI 生成とバリデーション',
				'モックを確認・操作できる組み込み Web UI'
			],
			installMethods: [
				{
					name: 'Homebrew',
					copy: 'brew tap Saku0512/specter https://github.com/Saku0512/specter\nbrew install specter',
					note: 'macOS / Linux で Homebrew を使っているなら一番手早い方法です。'
				},
				{
					name: 'curl',
					copy: 'curl -fsSL https://raw.githubusercontent.com/Saku0512/specter/main/install.sh | bash',
					note: 'macOS / Linux で使える 1 コマンドのインストール方法です。'
				},
				{
					name: 'PowerShell',
					copy: 'irm https://raw.githubusercontent.com/Saku0512/specter/main/install.ps1 | iex',
					note: 'Windows 向けのインストーラーです。'
				},
				{
					name: 'Docker',
					copy: 'docker run -v $(pwd)/config.yml:/config.yml ghcr.io/saku0512/specter -c /config.yml',
					note: '手元の環境を汚さず、一時的なモックサーバーとして使いたいときに便利です。'
				}
			],
			quickstart: [
				{
					title: '設定ファイルを生成',
					command: 'specter init',
					body: '大きなバックエンドを作る前に、小さな YAML ファイルから始められます。'
				},
				{
					title: 'モック API を起動',
					command: 'specter -c config.yml',
					body: 'サーバーをすぐ起動し、ファイル保存ごとに hot reload しながら調整できます。'
				},
				{
					title: '現実的なシナリオを作る',
					command: 'curl http://localhost:8080/users',
					body: '状態、vars、matching、delay、fault、組み込み UI を使って必要なフローを再現できます。'
				}
			],
			uiLog: ['Requests', 'Routes', 'State & Vars', 'Stores', 'Auto Refresh On'],
			terminalExample: `$ specter init
config.yml を作成しました

$ specter -c config.yml
registered 2 route(s):
  GET      /users
  POST     /login

UI running on http://localhost:4444`
		},
		en: {
			title: 'specter | Mock APIs without the ceremony',
			description:
				'specter is a lightweight mock API server with hot reload, OpenAPI support, stateful flows, and a built-in control room UI.',
			navDocs: 'Docs',
			navRepo: 'GitHub',
			eyebrow: 'lightweight mock API server',
			pronounced: 'pronounced',
			lede:
				'Mock APIs without the ceremony. Define routes in YAML, run a single binary, and keep your frontend, tests, and demos moving while the real backend catches up.',
			downloadLatest: 'Download Latest',
			readDocs: 'Read Docs',
			viewRepository: 'View Repository',
			downloadKicker: 'Download',
			downloadTitle: 'Pick the install path that matches your workflow',
			allBinaries: 'All binaries',
			copyCommand: 'Copy',
			copiedCommand: 'Copied',
			copyCommandLabel: 'Copy install command',
			howKicker: 'How It Works',
			howTitle: 'Three moves from blank folder to useful mock API',
			operateKicker: 'Operate',
			operateTitle: 'Control the mock without rebuilding your app around it',
			operateBody:
				'specter is built for the awkward middle of product development: frontend ahead of backend, demos that need deterministic failures, and tests that want a backend shape without a backend team on standby.',
			operateBody2:
				'You can start tiny with one route, then layer in state transitions, OpenAPI validation, filters, stores, or the built-in UI as the scenario gets more realistic.',
			useCasesKicker: 'Use Cases',
			useCasesTitle: 'Turn backend waiting time into build time',
			useCases: [
				{
					title: 'Frontend development',
					body: 'Build screens, forms, error handling, and empty states before the real API is ready.'
				},
				{
					title: 'E2E testing',
					body: 'Pin state, failures, rate limits, and delays for repeatable test environments.'
				},
				{
					title: 'Demos and prototypes',
					body: 'Run a stable demo flow with deterministic data and failure cases before backend work lands.'
				},
				{
					title: 'OpenAPI mocks',
					body: 'Generate config from OpenAPI and catch contract drift with request and response validation.'
				},
				{
					title: 'Retry / failure simulation',
					body: 'Use sequential responses, on_call, error_rate, and delay to exercise retries and timeouts.'
				},
				{
					title: 'Local integration',
					body: 'Proxy part of your traffic to a real API while mocking the rest locally.'
				}
			],
			securityKicker: 'Security',
			securityTitle: 'Designed for local development',
			securityBody:
				'specter is a local development tool intended for trusted networks. Do not expose it to the public internet without additional protections such as a firewall, reverse proxy, or authentication.',
			securityLink: 'Read Security Policy',
			projectKicker: 'Project',
			projectTitle: 'Project information',
			projectBody:
				'Security, license, releases, and roadmap live on GitHub. Issues and focused contributions are welcome.',
			projectLinks: [
				{ label: 'Documentation', href: `${base}/docs/` },
				{ label: 'Security Policy', href: 'https://github.com/Saku0512/specter/security/policy' },
				{ label: 'License', href: 'https://github.com/Saku0512/specter/blob/main/LICENSE' },
				{ label: 'Releases / Changelog', href: 'https://github.com/Saku0512/specter/releases' },
				{ label: 'Roadmap / Issues', href: 'https://github.com/Saku0512/specter/issues' },
				{ label: 'GitHub', href: 'https://github.com/Saku0512/specter' }
			],
			panel: [
				{
					title: 'Edit YAML',
					body: 'Routes reload on save, so changing a response is a quick feedback loop.'
				},
				{
					title: 'Use the control room UI',
					body: 'Inspect requests, tweak state, reset stores, and steer scenarios live.'
				},
				{
					title: 'Ship your frontend anyway',
					body: 'Keep development and demos moving while the real API is still settling down.'
				}
			],
			features: [
				'YAML-defined routes with hot reload',
				'Stateful mocks, vars, stores, rate limits, delays, and faults',
				'OpenAPI generation and validation',
				'Built-in web UI for inspecting and controlling mocks'
			],
			installMethods: [
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
			],
			quickstart: [
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
			],
			uiLog: ['Requests', 'Routes', 'State & Vars', 'Stores', 'Auto Refresh On'],
			terminalExample: `$ specter init
created config.yml

$ specter -c config.yml
registered 2 route(s):
  GET      /users
  POST     /login

UI running on http://localhost:4444`
		}
	};

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

	let copiedCommand = $state('');
	let copyTimer: ReturnType<typeof setTimeout> | undefined;
	const badges = [
		{
			alt: 'CI status',
			src: 'https://github.com/Saku0512/specter/actions/workflows/test.yml/badge.svg',
			href: 'https://github.com/Saku0512/specter/actions/workflows/test.yml'
		},
		{
			alt: 'Go Report Card',
			src: 'https://goreportcard.com/badge/github.com/Saku0512/specter',
			href: 'https://goreportcard.com/report/github.com/Saku0512/specter'
		},
		{
			alt: 'Latest release',
			src: 'https://img.shields.io/github/v/release/Saku0512/specter',
			href: 'https://github.com/Saku0512/specter/releases/latest'
		},
		{
			alt: 'MIT License',
			src: 'https://img.shields.io/badge/License-MIT-yellow.svg',
			href: 'https://github.com/Saku0512/specter/blob/main/LICENSE'
		}
	];

	async function copyInstallCommand(command: string, name: string) {
		try {
			await navigator.clipboard.writeText(command);
		} catch {
			const textarea = document.createElement('textarea');
			textarea.value = command;
			textarea.setAttribute('readonly', '');
			textarea.style.position = 'fixed';
			textarea.style.opacity = '0';
			document.body.appendChild(textarea);
			textarea.select();
			document.execCommand('copy');
			textarea.remove();
		}

		copiedCommand = name;

		if (copyTimer) {
			clearTimeout(copyTimer);
		}

		copyTimer = setTimeout(() => {
			copiedCommand = '';
		}, 1600);
	}
</script>

<svelte:head>
	<title>{copy[$language].title}</title>
	<meta
		name="description"
		content={copy[$language].description}
	/>
</svelte:head>

<div class="page">
	<section class="hero">
		<header class="site-header">
			<a class="brand" href={base ? `${base}/` : '/'}>
				<img src={mark} alt="" />
				<span>specter</span>
			</a>
			<div class="header-actions">
				<a href={`${base}/docs/`}>{copy[$language].navDocs}</a>
				<a href="https://github.com/Saku0512/specter">{copy[$language].navRepo}</a>
				<LanguageToggle />
			</div>
		</header>

		<div class="hero-inner">
			<div class="hero-copy">
				<div class="eyebrow">
					<img src={mark} alt="specter mark" />
					<span>{copy[$language].eyebrow}</span>
				</div>

				<h1>specter</h1>
				<p class="pronunciation" aria-label="specter pronunciation: SPEK-ter">
					<span>{copy[$language].pronounced}</span>
					<strong>SPEK-ter</strong>
					<code>/ˈspɛk.tɚ/</code>
					<small>スペクター</small>
				</p>
				<p class="lede">{copy[$language].lede}</p>

				<div class="cta-row">
					<a class="button primary" href="https://github.com/Saku0512/specter/releases/latest"
						>{copy[$language].downloadLatest}</a
					>
					<a class="button ghost" href={`${base}/docs/`}>{copy[$language].readDocs}</a>
					<a class="button ghost" href="https://github.com/Saku0512/specter"
						>{copy[$language].viewRepository}</a
					>
				</div>

				<ul class="signal-list">
					{#each copy[$language].features as feature}
						<li>{feature}</li>
					{/each}
				</ul>
			</div>

			<div class="hero-visual">
				<div class="scene">
					<div class="terminal">
						<div class="window-bar">
							<div class="window-title">
								<span></span><span></span><span></span>
								<strong>terminal</strong>
							</div>
							<button
								type="button"
								class="mini-copy"
								class:copied={copiedCommand === 'terminal-example'}
								aria-label={`${copy[$language].copyCommandLabel}: terminal`}
								onclick={() => copyInstallCommand(copy[$language].terminalExample, 'terminal-example')}
							>
								{copiedCommand === 'terminal-example' ? copy[$language].copiedCommand : copy[$language].copyCommand}
							</button>
						</div>
						<pre>{copy[$language].terminalExample}</pre>
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
							{#each copy[$language].uiLog as item}
								<div>{item}</div>
							{/each}
						</div>
					</div>

					<div class="config-card">
						<div class="window-bar">
							<strong>config.yml</strong>
							<button
								type="button"
								class="mini-copy"
								class:copied={copiedCommand === 'route-example'}
								aria-label={`${copy[$language].copyCommandLabel}: config.yml`}
								onclick={() => copyInstallCommand(routeExample, 'route-example')}
							>
								{copiedCommand === 'route-example' ? copy[$language].copiedCommand : copy[$language].copyCommand}
							</button>
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
				<p class="kicker">{copy[$language].downloadKicker}</p>
				<h2>{copy[$language].downloadTitle}</h2>
			</div>
			<a class="button ghost" href="https://github.com/Saku0512/specter/releases/latest"
				>{copy[$language].allBinaries}</a
			>
		</div>

		<div class="install-grid">
			{#each copy[$language].installMethods as method}
				<article class="install-card">
					<div class="card-top">
						<h3>{method.name}</h3>
						<p>{method.note}</p>
					</div>
					<button
						type="button"
						class="command-copy"
						aria-label={`${copy[$language].copyCommandLabel}: ${method.name}`}
						onclick={() => copyInstallCommand(method.copy, method.name)}
					>
						<code>{method.copy}</code>
						<span class="copy-badge" class:copied={copiedCommand === method.name}>
							<span class="copy-icon" aria-hidden="true"></span>
							{copiedCommand === method.name
								? copy[$language].copiedCommand
								: copy[$language].copyCommand}
						</span>
					</button>
				</article>
			{/each}
		</div>
	</section>

	<section class="band alt" id="how-it-works">
		<div class="band-head">
			<div>
				<p class="kicker">{copy[$language].howKicker}</p>
				<h2>{copy[$language].howTitle}</h2>
			</div>
		</div>

		<div class="steps">
			{#each copy[$language].quickstart as step, index}
				<article class="step">
					<div class="step-number">0{index + 1}</div>
					<h3>{step.title}</h3>
					<div class="step-command">
						<pre>{step.command}</pre>
						<button
							type="button"
							class="mini-copy"
							class:copied={copiedCommand === `quickstart-${index}`}
							aria-label={`${copy[$language].copyCommandLabel}: ${step.command}`}
							onclick={() => copyInstallCommand(step.command, `quickstart-${index}`)}
						>
							{copiedCommand === `quickstart-${index}` ? copy[$language].copiedCommand : copy[$language].copyCommand}
						</button>
					</div>
					<p>{step.body}</p>
				</article>
			{/each}
		</div>
	</section>

	<section class="band">
		<div class="explain">
			<div class="explain-copy">
				<p class="kicker">{copy[$language].operateKicker}</p>
				<h2>{copy[$language].operateTitle}</h2>
				<p>{copy[$language].operateBody}</p>
				<p>{copy[$language].operateBody2}</p>
			</div>

			<div class="explain-panel">
				{#each copy[$language].panel as item, index}
					<div class="panel-line">
						<span>{index + 1}</span>
						<div>
							<strong>{item.title}</strong>
							<p>{item.body}</p>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</section>

	<section class="band" id="use-cases">
		<div class="band-head">
			<div>
				<p class="kicker">{copy[$language].useCasesKicker}</p>
				<h2>{copy[$language].useCasesTitle}</h2>
			</div>
		</div>

		<div class="use-case-grid">
			{#each copy[$language].useCases as useCase}
				<article class="use-case">
					<h3>{useCase.title}</h3>
					<p>{useCase.body}</p>
				</article>
			{/each}
		</div>
	</section>

	<section class="band project-band">
		<div class="project-grid">
			<article class="security-note">
				<p class="kicker">{copy[$language].securityKicker}</p>
				<h2>{copy[$language].securityTitle}</h2>
				<p>{copy[$language].securityBody}</p>
				<a class="button ghost" href="https://github.com/Saku0512/specter/security/policy"
					>{copy[$language].securityLink}</a
				>
			</article>

			<article class="project-links">
				<p class="kicker">{copy[$language].projectKicker}</p>
				<h2>{copy[$language].projectTitle}</h2>
				<p>{copy[$language].projectBody}</p>
				<div class="link-grid">
					{#each copy[$language].projectLinks as link}
						<a href={link.href}>{link.label}</a>
					{/each}
				</div>
				<div class="badge-row" aria-label="Project badges">
					{#each badges as badge}
						<a href={badge.href}>
							<img src={badge.src} alt={badge.alt} />
						</a>
					{/each}
				</div>
			</article>
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

	.site-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		max-width: 1180px;
		margin: 0 auto;
	}

	.brand,
	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.85rem;
	}

	.brand {
		font-weight: 800;
	}

	.brand img {
		width: 2rem;
		height: 2rem;
	}

	.header-actions {
		flex-wrap: wrap;
		justify-content: flex-end;
		color: #c7d7ec;
		font-size: 0.92rem;
		font-weight: 700;
	}

	.header-actions a {
		padding: 0.45rem 0.2rem;
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
	.use-case,
	.security-note,
	.project-links,
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
		justify-content: space-between;
		gap: 0.45rem;
		padding: 0.8rem 0.95rem;
		border-bottom: 1px solid rgba(145, 184, 220, 0.11);
		background: rgba(255, 255, 255, 0.03);
		font-size: 0.78rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: #98accc;
	}

	.window-title {
		display: flex;
		align-items: center;
		gap: 0.45rem;
		min-width: 0;
	}

	.window-title strong,
	.window-bar > strong {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.window-title span {
		width: 0.58rem;
		height: 0.58rem;
		flex: 0 0 auto;
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

	.step-command {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: stretch;
		overflow: hidden;
		border: 1px solid rgba(145, 184, 220, 0.12);
		border-radius: 10px;
		background: rgba(5, 12, 22, 0.5);
	}

	.step-command pre {
		min-width: 0;
		padding: 0.85rem 0.9rem;
	}

	.mini-copy {
		flex: 0 0 auto;
		min-height: 1.85rem;
		padding: 0.35rem 0.62rem;
		border: 1px solid rgba(137, 216, 235, 0.24);
		border-radius: 8px;
		background: rgba(137, 216, 235, 0.08);
		color: #e1f6ff;
		font: inherit;
		font-size: 0.72rem;
		font-weight: 800;
		letter-spacing: 0;
		text-transform: none;
		cursor: pointer;
	}

	.mini-copy:hover,
	.mini-copy:focus-visible {
		border-color: rgba(137, 216, 235, 0.48);
		background: rgba(137, 216, 235, 0.14);
		outline: none;
	}

	.mini-copy.copied {
		border-color: rgba(186, 255, 207, 0.44);
		background: rgba(186, 255, 207, 0.14);
		color: #e3ffe9;
	}

	.step-command .mini-copy {
		margin: 0.48rem 0.48rem 0.48rem 0;
	}

	.command-copy {
		position: relative;
		display: block;
		width: 100%;
		margin: 0;
		padding: 0;
		border: 1px solid rgba(145, 184, 220, 0.12);
		border-radius: 10px;
		background: rgba(5, 12, 22, 0.5);
		color: #d8e5f7;
		cursor: pointer;
		text-align: left;
		overflow: hidden;
		transition:
			border-color 160ms ease,
			background 160ms ease,
			transform 160ms ease;
	}

	.command-copy:hover,
	.command-copy:focus-visible {
		border-color: rgba(138, 220, 238, 0.42);
		background: rgba(7, 18, 32, 0.78);
		transform: translateY(-1px);
		outline: none;
	}

	.command-copy code {
		display: block;
		min-height: 4rem;
		padding: 1rem 5.7rem 1.1rem 1.05rem;
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
	}

	.copy-badge {
		position: absolute;
		top: 0.75rem;
		right: 0.75rem;
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.34rem 0.52rem;
		border-radius: 999px;
		background: rgba(138, 241, 255, 0.08);
		color: #dff9ff;
		font-size: 0.72rem;
		font-weight: 800;
		line-height: 1;
	}

	.copy-badge.copied {
		background: rgba(185, 255, 210, 0.16);
		color: #baffcf;
	}

	.copy-icon {
		position: relative;
		width: 0.8rem;
		height: 0.8rem;
	}

	.copy-icon::before,
	.copy-icon::after {
		content: '';
		position: absolute;
		width: 0.52rem;
		height: 0.62rem;
		border: 1.5px solid currentColor;
		border-radius: 2px;
	}

	.copy-icon::before {
		left: 0.08rem;
		top: 0.18rem;
		opacity: 0.55;
	}

	.copy-icon::after {
		left: 0.22rem;
		top: 0.02rem;
		background: rgba(7, 18, 32, 0.9);
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

	.use-case-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
	}

	.use-case,
	.security-note,
	.project-links {
		padding: 1.15rem;
	}

	.use-case h3 {
		font-size: 1.28rem;
		margin-bottom: 0.45rem;
	}

	.use-case p,
	.security-note p,
	.project-links p {
		color: #c6d5eb;
		line-height: 1.7;
	}

	.project-band {
		padding-bottom: 3rem;
	}

	.project-grid {
		display: grid;
		grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr);
		gap: 1rem;
	}

	.security-note h2,
	.project-links h2 {
		font-size: clamp(1.7rem, 3vw, 2.45rem);
		margin: 0.25rem 0 0.85rem;
	}

	.security-note .button {
		margin-top: 0.55rem;
	}

	.link-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.65rem;
		margin-top: 1rem;
	}

	.link-grid a {
		padding: 0.78rem 0.85rem;
		border: 1px solid rgba(145, 184, 220, 0.12);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.035);
		color: #dff9ff;
		font-weight: 700;
	}

	.badge-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.55rem;
		margin-top: 1rem;
	}

	.badge-row a {
		display: inline-flex;
		align-items: center;
	}

	.badge-row img {
		height: 1.25rem;
		max-width: 12rem;
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
		.use-case-grid,
		.project-grid,
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

		.site-header {
			align-items: flex-start;
			flex-direction: column;
		}

		.header-actions {
			justify-content: flex-start;
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

		.link-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.eyebrow img {
			animation: none;
		}
	}
</style>
