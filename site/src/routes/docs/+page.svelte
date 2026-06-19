<script lang="ts">
	import { base } from '$app/paths';
	import LanguageToggle from '$lib/LanguageToggle.svelte';
	import { language } from '$lib/language';
	import mark from '$lib/assets/logo-icon.png';

	const copy = {
		ja: {
			title: 'specter docs | はじめに、Quick Start、config.yml、Contributing',
			description:
				'specter のドキュメント。はじめに、Quick Start、config.yml リファレンス、CLI、コントリビュート方法をまとめています。',
			toc: [
				{ href: '#introduction', label: 'はじめに' },
				{ href: '#quick-start', label: 'Quick Start' },
				{ href: '#config', label: 'config.yml' },
				{ href: '#examples', label: 'Examples' },
				{ href: '#comparison', label: '比較' },
				{ href: '#recipes', label: 'レシピ' },
				{ href: '#cli', label: 'CLI' },
				{ href: '#contributing', label: 'Contributing' }
			],
			kicker: 'Documentation',
			heroTitle: '1 つの YAML から信頼できるモック API を作る。',
			heroBody:
				'specter はフロントエンド開発、デモ、自動テスト、API 契約の確認に使える軽量なモック API サーバーです。config.yml にルートを書き、単一バイナリを起動するだけで、アプリを作り直さずに挙動を調整できます。',
			startQuickly: 'すぐ始める',
			configReference: '設定リファレンス',
			copySnippet: 'コピー',
			copiedSnippet: 'コピー済み',
			copySnippetLabel: 'コードをコピー',
			quickStartTitle: '空のフォルダからモック API へ',
			quickStart: [
				{
					title: 'スターターを作成',
					body: '現在のディレクトリに config.yml を作成します。'
				},
				{
					title: 'サーバーを起動',
					body: 'API は 8080、コントロール UI は 4444 で起動します。'
				},
				{
					title: 'ルートを呼び出す',
					body: 'YAML を編集して保存すると、specter が自動で再読み込みします。'
				}
			],
			configTitle: '完全リファレンス',
			configBody:
				'config は 1 つのルートだけの小さなファイルにも、状態管理、条件分岐、fixtures、OpenAPI 検証、proxy、stores、callbacks、delays、streams を含む本格的なシナリオにもできます。',
			examplesTitle: 'Examples gallery',
			examplesBody:
				'よくある mock pattern ごとに、すぐ生成できるサンプル config と使いどころをまとめました。CLI では specter examples <name> で作成できます。',
			examplesCta: '詳しい gallery を読む',
			exampleCards: [
				['auth', 'login、protected endpoint、state、vars、401 response の流れを試せます。'],
				['crud', 'in-memory store に REST endpoint を接続し、一覧・詳細・作成・更新・削除を再現します。'],
				['pagination', 'filtering、sorting、limit、offset を使う list endpoint の starting point です。'],
				['graphql', 'operationName と variables に応じて /graphql の response を分岐します。'],
				['webhooks', 'response 後の asynchronous callback を local listener に送信します。'],
				['sse', 'Server-Sent Events の stream と繰り返し event を再現します。'],
				['openapi', 'OpenAPI spec による request / response validation を試せます。'],
				['polling', 'long-running job の queued / running / complete を sequential responses で表現します。'],
				['errors', 'rate limit、flaky 503、latency、400/404 response で error UI を鍛えます。']
			],
			comparisonTitle: 'どの mock-server tool を選ぶか',
			comparisonBody:
				'Specter は YAML-first の local mock server です。json-server、Prism、WireMock と重なる部分はありますが、state、stores、scenarios、timelines、request assertions、Web UI を 1 つの local workflow にまとめることを重視しています。',
			comparisonCta: '比較ガイドを読む',
			comparisonCards: [
				['json-server', 'JSON database から CRUD REST API をすばやく作る用途に向いています。Specter は同じ path で条件分岐・state・scenario・assertion が必要なときに向いています。'],
				['Prism', 'OpenAPI / Postman contract を source of truth にした mock と validation proxy に強い tool です。Specter は OpenAPI validation に加えて hand-written behavior を重ねたいときに向いています。'],
				['WireMock', 'rich matching、verification、record/playback、JVM integration、service virtualization に強い mature tool です。Specter は小さく読める YAML と built-in UI で local dev / E2E を軽く回したいときに向いています。']
			],
			sections: {
				topLevel: {
					title: 'トップレベルのフィールド',
					body: 'YAML ファイルのルートに 1 回だけ書くフィールドです。'
				},
				route: {
					title: 'ルートのフィールド',
					body: 'routes の各要素は、モック、proxy、store 操作、redirect、stream のいずれかを表します。'
				},
				match: {
					title: 'match のフィールド',
					body: '同じ method と path で、リクエスト内容に応じてレスポンスを分岐したいときに使います。1 つの match 内の条件はすべて満たす必要があります。'
				},
				response: {
					title: 'responses の要素',
					body: 'responses と mode: sequential / random を使うと、retry、polling、不安定な API、変化するデータを再現できます。'
				}
			},
			tableHead: ['Field', 'Type', 'Description'],
			recipesTitle: 'よく使う config パターン',
			recipes: {
				conditional: {
					title: '条件付きレスポンス',
					body: 'request body、query、headers、cookies、form data、GraphQL の値で分岐できます。'
				},
				state: {
					title: '状態を持つフロー',
					body: 'login flow やシナリオの分岐には state、set_state、vars、set_vars を使います。'
				},
				store: {
					title: 'インメモリ CRUD',
					body: 'REST endpoint を名前付き store に直接接続できます。store のデータはサーバー再起動時にリセットされます。'
				},
				advanced: {
					title: 'OpenAPI、proxy、chaos、SSE',
					body: '実サービスと mock を混ぜ、request 検証、jitter、failure injection、event stream を使えます。'
				},
				template: {
					title: 'Templates と faker',
					body: 'template は .body、.query、.params、.headers、.method、.path を参照できます。'
				},
				fixture: {
					title: 'Fixtures、redirects、cookies、webhooks',
					body: '大きな JSON/YAML/text fixture には file、HTTP redirect には redirect、認証 simulation には set_cookies、非同期 callback には webhook を使います。'
				}
			},
			cliTitle: 'よく使うコマンド',
			cliBody:
				'flags は environment variables より優先されます。API のデフォルト port は 8080、組み込み dashboard は 4444 です。無効化するには --ui-port 0 を指定します。',
			contributingTitle: 'specter を一緒に良くする',
			contributingBody:
				'docs、examples、bug fixes、CLI behavior、validation、UI improvements、新しい mock-server features への contribution を歓迎します。変更は focused にし、挙動変更には tests を追加し、config や CLI behavior が変わる場合は docs も更新してください。',
			contributingSteps: [
				{
					title: '1. Fork and branch',
					body: '目的が分かる branch を作り、小さく意味のある変更にします。'
				},
				{
					title: '2. Validate locally',
					body: '関連する tests を実行し、docs examples には specter validate -c config.yml を使います。'
				},
				{
					title: '3. Open a pull request',
					body: 'scenario、変更内容、既存 config への compatibility notes を書いて pull request を作成します。'
				}
			],
			projectKicker: 'Project',
			projectTitle: 'Security、License、Roadmap',
			projectBody:
				'specter は local development tool です。public internet に公開する場合は追加の security measures を用意してください。Security policy、license、releases、roadmap は GitHub で確認できます。',
			projectLinks: [
				{ label: 'Security Policy', href: 'https://github.com/Saku0512/specter/security/policy' },
				{ label: 'License', href: 'https://github.com/Saku0512/specter/blob/main/LICENSE' },
				{ label: 'Releases / Changelog', href: 'https://github.com/Saku0512/specter/releases' },
				{ label: 'Roadmap / Issues', href: 'https://github.com/Saku0512/specter/issues' },
				{ label: 'GitHub', href: 'https://github.com/Saku0512/specter' },
				{ label: 'README', href: 'https://github.com/Saku0512/specter#readme' }
			]
		},
		en: {
			title: 'specter docs | Introduction, Quick Start, config.yml, Contributing',
			description:
				'Documentation for specter, including introduction, quick start, complete config.yml reference, CLI usage, and contribution guide.',
			toc: [
				{ href: '#introduction', label: 'Introduction' },
				{ href: '#quick-start', label: 'Quick Start' },
				{ href: '#config', label: 'config.yml' },
				{ href: '#examples', label: 'Examples' },
				{ href: '#comparison', label: 'Comparison' },
				{ href: '#recipes', label: 'Recipes' },
				{ href: '#cli', label: 'CLI' },
				{ href: '#contributing', label: 'Contributing' }
			],
			kicker: 'Documentation',
			heroTitle: 'Build reliable mock APIs from one YAML file.',
			heroBody:
				'specter is a lightweight mock API server for frontend development, demos, automated tests, and API contract work. Define routes in config.yml, run a single binary, and adjust behavior without rebuilding your application.',
			startQuickly: 'Start Quickly',
			configReference: 'Config Reference',
			copySnippet: 'Copy',
			copiedSnippet: 'Copied',
			copySnippetLabel: 'Copy code',
			quickStartTitle: 'From blank folder to mock API',
			quickStart: [
				{ title: 'Generate a starter file', body: 'Creates config.yml in the current directory.' },
				{
					title: 'Run the server',
					body: 'The API listens on port 8080 and the control UI opens on port 4444.'
				},
				{ title: 'Call a route', body: 'Edit the YAML and save. specter reloads the config automatically.' }
			],
			configTitle: 'Complete reference',
			configBody:
				'A config can stay tiny with one route, or grow into a full scenario with state, conditional matching, fixtures, OpenAPI validation, proxying, stores, callbacks, delays, and streams.',
			examplesTitle: 'Examples gallery',
			examplesBody:
				'Browse common mock patterns with generated sample configs and guidance on when to use each one. From the CLI, create a starter with specter examples <name>.',
			examplesCta: 'Read the full gallery',
			exampleCards: [
				['auth', 'Try login, protected endpoints, state, vars, and 401 responses.'],
				['crud', 'Connect REST endpoints to an in-memory store for list, detail, create, update, and delete flows.'],
				['pagination', 'Start list endpoints with filtering, sorting, limit, and offset query params.'],
				['graphql', 'Branch /graphql responses by operationName and variables.'],
				['webhooks', 'Send an asynchronous callback to a local listener after responding.'],
				['sse', 'Model Server-Sent Events streams and repeated event sequences.'],
				['openapi', 'Validate mock requests and responses against an OpenAPI spec.'],
				['polling', 'Represent long-running queued, running, and complete jobs with sequential responses.'],
				['errors', 'Exercise error UI with rate limits, flaky 503s, latency, and 400/404 responses.']
			],
			comparisonTitle: 'Choosing a mock-server tool',
			comparisonBody:
				'Specter is a YAML-first local mock server. It overlaps with json-server, Prism, and WireMock, but focuses on combining state, stores, scenarios, timelines, request assertions, and a Web UI into one local workflow.',
			comparisonCta: 'Read the comparison guide',
			comparisonCards: [
				['json-server', 'Best when you want a CRUD REST API from a JSON database. Specter fits better when one path needs matching, state, scenarios, or assertions.'],
				['Prism', 'Best when an OpenAPI or Postman contract is the source of truth for mocks and validation proxy behavior. Specter fits better when you want OpenAPI validation plus hand-written behavior.'],
				['WireMock', 'Best for rich matching, verification, record/playback, JVM integration, and broad service virtualization. Specter fits better when local dev and E2E need a small readable YAML workflow with a built-in UI.']
			],
			sections: {
				topLevel: { title: 'Top-level fields', body: 'Use these fields once at the root of the YAML file.' },
				route: {
					title: 'Route fields',
					body: 'Each item in routes describes one mock, proxy, store operation, redirect, or stream.'
				},
				match: {
					title: 'Match fields',
					body: 'Use match when one method and path should branch based on request data. All conditions in one match entry must pass.'
				},
				response: {
					title: 'Response entries',
					body: 'Use responses with mode: sequential or mode: random to simulate retries, polling, flaky APIs, or changing data.'
				}
			},
			tableHead: ['Field', 'Type', 'Description'],
			recipesTitle: 'Common config patterns',
			recipes: {
				conditional: {
					title: 'Conditional responses',
					body: 'Branch by request body, query parameters, headers, cookies, form data, or GraphQL values.'
				},
				state: {
					title: 'Stateful flows',
					body: 'Use state, set_state, vars, and set_vars for login flows and scenario gates.'
				},
				store: {
					title: 'In-memory CRUD',
					body: 'Wire REST endpoints directly to a named store. Store data resets when the server restarts.'
				},
				advanced: {
					title: 'OpenAPI, proxy, chaos, and SSE',
					body: 'Mix real services with mocks, validate requests, add jitter, inject failures, and stream events.'
				},
				template: {
					title: 'Templates and faker',
					body: 'Templates can read .body, .query, .params, .headers, .method, and .path.'
				},
				fixture: {
					title: 'Fixtures, redirects, cookies, and webhooks',
					body: 'Use file for large JSON/YAML/text fixtures, redirect for HTTP redirects, set_cookies for auth simulations, and webhook for async callbacks.'
				}
			},
			cliTitle: 'Useful commands',
			cliBody:
				'Flags override environment variables. The default API port is 8080, and the built-in dashboard uses 4444. Set --ui-port 0 to disable it.',
			contributingTitle: 'Help improve specter',
			contributingBody:
				'Contributions are welcome across docs, examples, bug fixes, CLI behavior, validation, UI improvements, and new mock-server features. Keep changes focused, add tests for behavioral changes, and update documentation when config or CLI behavior changes.',
			contributingSteps: [
				{ title: '1. Fork and branch', body: 'Create a branch with a focused name, then make the smallest useful change.' },
				{ title: '2. Validate locally', body: 'Run the relevant tests and use specter validate -c config.yml for docs examples.' },
				{ title: '3. Open a pull request', body: 'Describe the scenario, what changed, and any compatibility notes for existing configs.' }
			],
			projectKicker: 'Project',
			projectTitle: 'Security, license, and roadmap',
			projectBody:
				'specter is a local development tool. If you expose it to the public internet, add security measures such as a firewall, reverse proxy, or authentication. Security policy, license, releases, and roadmap live on GitHub.',
			projectLinks: [
				{ label: 'Security Policy', href: 'https://github.com/Saku0512/specter/security/policy' },
				{ label: 'License', href: 'https://github.com/Saku0512/specter/blob/main/LICENSE' },
				{ label: 'Releases / Changelog', href: 'https://github.com/Saku0512/specter/releases' },
				{ label: 'Roadmap / Issues', href: 'https://github.com/Saku0512/specter/issues' },
				{ label: 'GitHub', href: 'https://github.com/Saku0512/specter' },
				{ label: 'README', href: 'https://github.com/Saku0512/specter#readme' }
			]
		}
	};

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

	const topLevelFields = [
		['cors', 'boolean', 'Enables CORS headers and handles OPTIONS preflight requests.', 'CORS headers を有効化し、OPTIONS preflight requests を処理します。'],
		['proxy', 'string', 'Forwards unmatched requests to a real backend.', '一致する route がない requests を実 backend に転送します。'],
		['openapi', 'string', 'Path to an OpenAPI YAML or JSON file for request and response validation.', 'request / response validation に使う OpenAPI YAML または JSON file の path です。'],
		['openapi_strict', 'boolean', 'When true, invalid requests return 400 instead of only adding a warning header.', 'true の場合、invalid requests は warning header の追加だけでなく 400 を返します。'],
		[
			'openapi_strict_response',
			'boolean',
			'When true, invalid mock responses return 500 instead of being served with a warning header.',
			'true の場合、invalid mock responses は warning header 付きで返されず 500 になります。'
		],
		['include', 'list', 'Merges routes from other YAML files. Glob patterns are supported.', '他の YAML files から routes を merge します。glob patterns も使えます。'],
		['routes', 'list', 'The route definitions served by specter.', 'specter が提供する route definitions です。']
	];

	const routeFields = [
		['path', 'string', 'Required. URL path. Supports :param path parameters.', '必須。URL path です。:param 形式の path parameters を使えます。'],
		['method', 'string', 'Required. HTTP method such as GET, POST, PUT, PATCH, or DELETE.', '必須。GET、POST、PUT、PATCH、DELETE などの HTTP method です。'],
		['status', 'int', 'Response status code. Defaults to 200.', 'response status code です。default は 200 です。'],
		['response', 'any', 'Inline JSON object, array, scalar, or string response body.', 'inline の JSON object、array、scalar、string response body です。'],
		['headers', 'map', 'Custom response headers for the route.', 'route に付与する custom response headers です。'],
		['content_type', 'string', 'Response MIME type. Defaults to application/json.', 'response MIME type です。default は application/json です。'],
		['delay', 'int', 'Fixed response delay in milliseconds.', '固定 response delay を milliseconds で指定します。'],
		['delay_min', 'int', 'Minimum random delay in milliseconds. Use with delay_max.', 'random delay の最小値を milliseconds で指定します。delay_max と一緒に使います。'],
		['delay_max', 'int', 'Maximum random delay in milliseconds. Use with delay_min.', 'random delay の最大値を milliseconds で指定します。delay_min と一緒に使います。'],
		['error_rate', 'float', 'Probability from 0.0 to 1.0 that the route returns an injected error.', 'injected error を返す確率を 0.0 から 1.0 で指定します。'],
		['error_status', 'int', 'Status code for injected errors. Defaults to 503.', 'injected errors の status code です。default は 503 です。'],
		['on_call', 'int', 'Only match this route on a specific 1-based call number.', '1 始まりの指定 call number のときだけこの route に match します。'],
		['match', 'list', 'Conditional responses based on query, headers, body, form data, cookies, or GraphQL.', 'query、headers、body、form data、cookies、GraphQL に基づく conditional responses です。'],
		['mode', 'string', 'Controls responses selection. Use sequential or random.', 'responses の選択方法です。sequential または random を使います。'],
		['responses', 'list', 'Multiple response entries for cycling, retry simulation, or random behavior.', 'cycling、retry simulation、random behavior 用の複数 response entries です。'],
		['rate_limit', 'int', 'Maximum requests before returning 429.', '429 を返すまでの最大 requests 数です。'],
		['rate_reset', 'int', 'Seconds until the rate-limit counter resets. Adds Retry-After on 429.', 'rate-limit counter が reset されるまでの秒数です。429 に Retry-After を追加します。'],
		['state', 'string', 'Only match when the server state equals this value.', 'server state がこの値と等しいときだけ match します。'],
		['set_state', 'string', 'Set the server state after responding.', 'response 後に server state を設定します。'],
		['vars', 'map', 'Only match when all named variables equal these values.', '指定した variables がすべてこの値と等しいときだけ match します。'],
		['set_vars', 'map', 'Set named variables after responding. Values can use templates.', 'response 後に named variables を設定します。値には templates を使えます。'],
		['webhook', 'object', 'Fire an outgoing callback after the response.', 'response 後に outgoing callback を送信します。'],
		['file', 'string', 'Serve a response body from a JSON, YAML, or text fixture file.', 'JSON、YAML、text fixture file から response body を返します。'],
		['script', 'string', 'Go template that generates the response body. Takes priority over file and response.', 'response body を生成する Go template です。file と response より優先されます。'],
		['proxy', 'string', 'Forward this route to a real backend. Takes priority over mock response fields.', 'この route を実 backend に転送します。mock response fields より優先されます。'],
		['store_push', 'string', 'Push the request body into an in-memory store and respond 201.', 'request body を in-memory store に追加し 201 を返します。'],
		['store_list', 'string', 'List all items in an in-memory store with filtering, sorting, and pagination.', 'in-memory store の items を filtering、sorting、pagination 付きで一覧します。'],
		['store_get', 'string', 'Get one store item by the store_key path parameter.', 'store_key path parameter で store item を 1 件取得します。'],
		['store_put', 'string', 'Replace or upsert one store item.', 'store item を replace または upsert します。'],
		['store_patch', 'string', 'Merge the request body into one store item.', 'request body を store item に merge します。'],
		['store_delete', 'string', 'Delete one store item.', 'store item を 1 件削除します。'],
		['store_clear', 'string', 'Delete every item in a named store.', 'named store 内のすべての item を削除します。'],
		['store_key', 'string', 'Path parameter used as the item ID. Defaults to id.', 'item ID として使う path parameter です。default は id です。'],
		['stream', 'boolean', 'Respond with a Server-Sent Events stream.', 'Server-Sent Events stream として response します。'],
		['events', 'list', 'Ordered SSE events for a stream route.', 'stream route 用の順序付き SSE events です。'],
		['stream_repeat', 'boolean', 'Repeat SSE events until the client disconnects.', 'client が disconnect するまで SSE events を繰り返します。'],
		['set_cookies', 'list', 'Set cookies in the response.', 'response に cookies を設定します。'],
		['redirect', 'string', 'Redirect to another path or URL.', '別の path または URL に redirect します。'],
		['redirect_status', 'int', 'Redirect status code. Use 301, 302, 303, 307, or 308.', 'redirect status code です。301、302、303、307、308 を使います。']
	];

	const matchFields = [
		['query', 'map', 'Match query parameters using Go regular expressions.', 'query parameters を Go regular expressions で match します。'],
		['headers', 'map', 'Match request headers. Header names are case-insensitive.', 'request headers を match します。header names は case-insensitive です。'],
		['body', 'map', 'Match top-level JSON request body fields.', 'top-level JSON request body fields を match します。'],
		['body_path', 'map', 'Match nested JSON fields with dot notation such as user.role.', 'user.role のような dot notation で nested JSON fields を match します。'],
		['form', 'map', 'Match application/x-www-form-urlencoded request bodies.', 'application/x-www-form-urlencoded request bodies を match します。'],
		['graphql', 'object', 'Match GraphQL operationName and variables.', 'GraphQL operationName と variables を match します。'],
		['cookies', 'map', 'Match request cookies with regex patterns.', 'request cookies を regex patterns で match します。'],
		['status', 'int', 'Status code returned when this match fires.', 'この match が発火したときに返す status code です。'],
		['response', 'any', 'Response body returned when this match fires.', 'この match が発火したときに返す response body です。'],
		['response_headers', 'map', 'Headers applied only for this match. Overrides route headers.', 'この match にだけ適用する headers です。route headers を override します。'],
		['content_type', 'string', 'Content type applied only for this match.', 'この match にだけ適用する content type です。'],
		['delay', 'int', 'Extra delay in milliseconds for this match. Added after route delay.', 'この match 用の追加 delay です。route delay の後に加算されます。'],
		['set_state', 'string', 'State transition applied only when this match fires.', 'この match が発火したときだけ適用する state transition です。'],
		['set_vars', 'map', 'Variable updates applied only when this match fires.', 'この match が発火したときだけ適用する variable updates です。'],
		['file', 'string', 'Fixture file returned only when this match fires.', 'この match が発火したときだけ返す fixture file です。'],
		['script', 'string', 'Template response returned only when this match fires.', 'この match が発火したときだけ返す template response です。']
	];

	const responseFields = [
		['on_call', 'int', 'Pin this response entry to a specific call number.', 'この response entry を特定の call number に固定します。'],
		['status', 'int', 'Status for this response entry.', 'この response entry の status です。'],
		['response', 'any', 'Inline body for this response entry.', 'この response entry の inline body です。'],
		['content_type', 'string', 'Content type for this response entry.', 'この response entry の content type です。'],
		['delay', 'int', 'Delay for this response entry.', 'この response entry の delay です。'],
		['file', 'string', 'Fixture file for this response entry.', 'この response entry の fixture file です。'],
		['script', 'string', 'Template body for this response entry.', 'この response entry の template body です。']
	];

	const fakerTypes = [
		'name',
		'first_name',
		'last_name',
		'email',
		'uuid',
		'phone',
		'url',
		'ip',
		'username',
		'password',
		'word',
		'sentence',
		'paragraph',
		'color',
		'country',
		'city',
		'zip',
		'street',
		'company',
		'job',
		'int',
		'float',
		'bool',
		'date',
		'datetime'
	];

	const basicConfig = `cors: true

routes:
  - path: /users
    method: GET
    status: 200
    response:
      - id: 1
        name: Alice

  - path: /users/:id
    method: GET
    response:
      id: ":id"
      name: Alice`;

	const matchingConfig = `routes:
  - path: /users
    method: POST
    match:
      - body:
          role: admin
        status: 201
        response:
          id: 1
          role: admin
      - query:
          preview: "^true$"
        response_headers:
          X-Preview: "true"
        response:
          id: 2
          preview: true
    status: 400
    response:
      error: no matching scenario`;

	const stateConfig = `routes:
  - path: /login
    method: POST
    set_state: logged_in
    set_vars:
      role: "{{ .body.role }}"
    response:
      token: abc123

  - path: /profile
    method: GET
    state: logged_in
    vars:
      role: admin
    response:
      name: Alice
      role: admin

  - path: /profile
    method: GET
    status: 401
    response:
      error: unauthorized`;

	const storeConfig = `routes:
  - path: /users
    method: POST
    store_push: users

  - path: /users
    method: GET
    store_list: users

  - path: /users/:id
    method: PATCH
    store_patch: users
    store_key: id

  - path: /users/:id
    method: DELETE
    store_delete: users`;

	const advancedConfig = `include:
  - routes/*.yml

openapi: ./openapi.yaml
openapi_strict: true
openapi_strict_response: false
proxy: https://api.example.com

routes:
  - path: /flaky
    method: GET
    delay_min: 150
    delay_max: 900
    error_rate: 0.25
    error_status: 503
    response:
      ok: true

  - path: /events
    method: GET
    stream: true
    events:
      - data: { type: connected }
      - event: done
        data: { ok: true }
        delay: 500`;

	const scriptConfig = `routes:
  - path: /greet
    method: POST
    script: |
      {
        "message": "Hello, {{ .body.name | default "friend" }}",
        "id": "{{ fake "uuid" }}",
        "created_at": "{{ now }}"
      }`;

	const cliCommands = `specter init
specter validate -c config.yml
specter -c config.yml -p 8080
specter gen -i openapi.yml -o config.yml
specter record -t http://api.example.com -o config.yml`;

	const storeQuery = `GET /users?role=admin&_sort=name&_order=asc&_limit=10&_offset=0`;

	let copiedSnippet = $state('');
	let copyTimer: ReturnType<typeof setTimeout> | undefined;

	async function copyCode(code: string, key: string) {
		try {
			await navigator.clipboard.writeText(code);
		} catch {
			const textarea = document.createElement('textarea');
			textarea.value = code;
			textarea.setAttribute('readonly', '');
			textarea.style.position = 'fixed';
			textarea.style.opacity = '0';
			document.body.appendChild(textarea);
			textarea.select();
			document.execCommand('copy');
			textarea.remove();
		}

		copiedSnippet = key;

		if (copyTimer) {
			clearTimeout(copyTimer);
		}

		copyTimer = setTimeout(() => {
			copiedSnippet = '';
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

<div class="docs-page">
	<header class="docs-hero" id="introduction">
		<nav class="topbar" aria-label="Main navigation">
			<a class="brand" href={base ? `${base}/` : '/'}>
				<img src={mark} alt="" />
				<span>specter</span>
			</a>
			<div class="nav-links">
				{#each copy[$language].toc as item}
					<a href={item.href}>{item.label}</a>
				{/each}
				<LanguageToggle />
			</div>
		</nav>

		<div class="hero-grid">
			<div>
				<p class="kicker">{copy[$language].kicker}</p>
				<h1>{copy[$language].heroTitle}</h1>
				<p class="lede">{copy[$language].heroBody}</p>
				<div class="hero-actions">
					<a class="button primary" href="#quick-start">{copy[$language].startQuickly}</a>
					<a class="button ghost" href="#config">{copy[$language].configReference}</a>
				</div>
			</div>

			{@render codeBlock(basicConfig)}
		</div>
	</header>

	<main>
		<section class="section" id="quick-start">
			<div class="section-head">
				<p class="kicker">Quick Start</p>
				<h2>{copy[$language].quickStartTitle}</h2>
			</div>

			<div class="steps">
				<article>
					<span>01</span>
					<h3>{copy[$language].quickStart[0].title}</h3>
					{@render commandBlock('specter init', 'quickstart-init')}
					<p>{copy[$language].quickStart[0].body}</p>
				</article>
				<article>
					<span>02</span>
					<h3>{copy[$language].quickStart[1].title}</h3>
					{@render commandBlock('specter -c config.yml', 'quickstart-run')}
					<p>{copy[$language].quickStart[1].body}</p>
				</article>
				<article>
					<span>03</span>
					<h3>{copy[$language].quickStart[2].title}</h3>
					{@render commandBlock('curl http://localhost:8080/users', 'quickstart-curl')}
					<p>{copy[$language].quickStart[2].body}</p>
				</article>
			</div>
		</section>

		<section class="section" id="config">
			<div class="section-head wide">
				<p class="kicker">config.yml</p>
				<h2>{copy[$language].configTitle}</h2>
				<p>{copy[$language].configBody}</p>
			</div>

			<div class="reference-grid">
				<div class="reference-copy">
					<h3>{copy[$language].sections.topLevel.title}</h3>
					<p>{copy[$language].sections.topLevel.body}</p>
				</div>
				{@render table(topLevelFields)}
			</div>

			<div class="reference-grid">
				<div class="reference-copy">
					<h3>{copy[$language].sections.route.title}</h3>
					<p>{copy[$language].sections.route.body}</p>
				</div>
				{@render table(routeFields)}
			</div>

			<div class="reference-grid">
				<div class="reference-copy">
					<h3>{copy[$language].sections.match.title}</h3>
					<p>{copy[$language].sections.match.body}</p>
				</div>
				{@render table(matchFields)}
			</div>

			<div class="reference-grid">
				<div class="reference-copy">
					<h3>{copy[$language].sections.response.title}</h3>
					<p>{copy[$language].sections.response.body}</p>
				</div>
				{@render table(responseFields)}
			</div>
		</section>

		<section class="section" id="examples">
			<div class="section-head wide">
				<p class="kicker">Examples</p>
				<h2>{copy[$language].examplesTitle}</h2>
				<p>{copy[$language].examplesBody}</p>
				<a class="text-link" href="https://github.com/Saku0512/specter/blob/main/doc/examples.md">{copy[$language].examplesCta}</a>
			</div>

			<div class="example-grid">
				{#each copy[$language].exampleCards as example}
					<article>
						<strong>{example[0]}</strong>
						<p>{example[1]}</p>
						<code>{example[0] === 'polling' ? 'doc/examples.md' : `specter examples ${example[0]}`}</code>
					</article>
				{/each}
			</div>
		</section>

		<section class="section" id="comparison">
			<div class="section-head wide">
				<p class="kicker">Comparison</p>
				<h2>{copy[$language].comparisonTitle}</h2>
				<p>{copy[$language].comparisonBody}</p>
				<a class="text-link" href="https://github.com/Saku0512/specter/blob/main/doc/comparison.md">{copy[$language].comparisonCta}</a>
			</div>

			<div class="comparison-grid">
				{#each copy[$language].comparisonCards as item}
					<article>
						<strong>{item[0]}</strong>
						<p>{item[1]}</p>
					</article>
				{/each}
			</div>
		</section>

		<section class="section" id="recipes">
			<div class="section-head">
				<p class="kicker">Recipes</p>
				<h2>{copy[$language].recipesTitle}</h2>
			</div>

			<div class="recipe-grid">
				<article>
					<h3>{copy[$language].recipes.conditional.title}</h3>
					<p>{copy[$language].recipes.conditional.body}</p>
					{@render codeBlock(matchingConfig)}
				</article>

				<article>
					<h3>{copy[$language].recipes.state.title}</h3>
					<p>{copy[$language].recipes.state.body}</p>
					{@render codeBlock(stateConfig)}
				</article>

				<article>
					<h3>{copy[$language].recipes.store.title}</h3>
					<p>{copy[$language].recipes.store.body}</p>
					{@render codeBlock(storeConfig)}
					{@render codeBlock(storeQuery, 'request')}
				</article>

				<article>
					<h3>{copy[$language].recipes.advanced.title}</h3>
					<p>{copy[$language].recipes.advanced.body}</p>
					{@render codeBlock(advancedConfig)}
				</article>

				<article>
					<h3>{copy[$language].recipes.template.title}</h3>
					<p>{copy[$language].recipes.template.body}</p>
					{@render codeBlock(scriptConfig)}
					<div class="pill-list" aria-label="Faker types">
						{#each fakerTypes as type}
							<code>{type}</code>
						{/each}
					</div>
				</article>

				<article>
					<h3>{copy[$language].recipes.fixture.title}</h3>
					<p>{copy[$language].recipes.fixture.body}</p>
					{@render codeBlock(`routes:
  - path: /login
    method: POST
    set_cookies:
      - name: session
        value: sess_abc123
        http_only: true
    webhook:
      url: http://localhost:9000/events
      body: { event: logged_in }
    response: { ok: true }

  - path: /old
    method: GET
    redirect: /new
    redirect_status: 301`)}
				</article>
			</div>
		</section>

		<section class="section split" id="cli">
			<div>
				<p class="kicker">CLI</p>
				<h2>{copy[$language].cliTitle}</h2>
				<p>{copy[$language].cliBody}</p>
			</div>
			<div class="hero-panel">
				<div class="panel-title copyable-title">
					<span>commands</span>
					<button
						type="button"
						class="copy-button"
						class:copied={copiedSnippet === 'cli-commands'}
						aria-label={`${copy[$language].copySnippetLabel}: commands`}
						onclick={() => copyCode(cliCommands, 'cli-commands')}
					>
						{copiedSnippet === 'cli-commands' ? copy[$language].copiedSnippet : copy[$language].copySnippet}
					</button>
				</div>
				<pre>{cliCommands}</pre>
			</div>
		</section>

		<section class="section split" id="contributing">
			<div>
				<p class="kicker">Contributing</p>
				<h2>{copy[$language].contributingTitle}</h2>
				<p>{copy[$language].contributingBody}</p>
			</div>

			<div class="contribute-list">
				{#each copy[$language].contributingSteps as step}
					<div>
						<strong>{step.title}</strong>
						<p>{step.body}</p>
					</div>
				{/each}
			</div>
		</section>
	</main>

	<footer class="project-footer">
		<div>
			<p class="kicker">{copy[$language].projectKicker}</p>
			<h2>{copy[$language].projectTitle}</h2>
			<p>{copy[$language].projectBody}</p>
		</div>

		<div>
			<div class="footer-links">
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
		</div>
	</footer>
</div>

{#snippet table(rows: string[][])}
	<div class="table-wrap">
		<table>
			<thead>
				<tr>
					{#each copy[$language].tableHead as heading}
						<th>{heading}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each rows as row}
					<tr>
						<td><code>{row[0]}</code></td>
						<td>{row[1]}</td>
						<td>{row[$language === 'ja' ? 3 : 2]}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/snippet}

{#snippet codeBlock(code: string, label = 'config.yml')}
	<div class="code-block">
		<div class="code-head">
			<div class="code-title">
				<span></span>
				<span></span>
				<span></span>
				<strong>{label}</strong>
			</div>
			<button
				type="button"
				class="copy-button"
				class:copied={copiedSnippet === `${label}:${code}`}
				aria-label={`${copy[$language].copySnippetLabel}: ${label}`}
				onclick={() => copyCode(code, `${label}:${code}`)}
			>
				{copiedSnippet === `${label}:${code}` ? copy[$language].copiedSnippet : copy[$language].copySnippet}
			</button>
		</div>
		<pre><code>{code}</code></pre>
	</div>
{/snippet}

{#snippet commandBlock(command: string, key: string)}
	<div class="command-block">
		<pre>{command}</pre>
		<button
			type="button"
			class="copy-button"
			class:copied={copiedSnippet === key}
			aria-label={`${copy[$language].copySnippetLabel}: ${command}`}
			onclick={() => copyCode(command, key)}
		>
			{copiedSnippet === key ? copy[$language].copiedSnippet : copy[$language].copySnippet}
		</button>
	</div>
{/snippet}

<style>
	:global(body) {
		margin: 0;
		background: #09111d;
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

	:global(html) {
		scroll-behavior: smooth;
	}

	:global(a) {
		color: inherit;
		text-decoration: none;
	}

	.docs-page {
		min-height: 100svh;
		background:
			linear-gradient(180deg, rgba(15, 32, 55, 0.96) 0%, rgba(8, 17, 29, 0.98) 32rem),
			#09111d;
	}

	.docs-hero,
	main,
	.project-footer {
		width: min(100%, 1180px);
		margin: 0 auto;
		padding-inline: clamp(1rem, 2vw, 2rem);
	}

	.docs-hero {
		padding-top: 1.25rem;
		padding-bottom: 3rem;
	}

	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: clamp(2rem, 6vw, 5rem);
	}

	.brand,
	.nav-links,
	.hero-actions {
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

	.nav-links {
		flex-wrap: wrap;
		justify-content: flex-end;
		color: #c7d7ec;
		font-size: 0.92rem;
	}

	.nav-links a {
		padding: 0.45rem 0.2rem;
	}

	.hero-grid,
	.split {
		display: grid;
		grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr);
		gap: clamp(1.2rem, 4vw, 3rem);
		align-items: start;
	}

	.kicker {
		margin: 0 0 0.8rem;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		font-size: 0.78rem;
		color: #8adcee;
		font-weight: 700;
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
		letter-spacing: 0;
	}

	h1 {
		font-size: clamp(3rem, 7vw, 5.8rem);
		line-height: 0.98;
		max-width: 10ch;
	}

	h2 {
		font-size: clamp(2rem, 4vw, 3.15rem);
		line-height: 1.05;
	}

	h3 {
		font-size: 1.35rem;
		line-height: 1.2;
	}

	p {
		color: #c7d7ec;
		line-height: 1.72;
	}

	code {
		font-family:
			'SFMono-Regular',
			'JetBrains Mono',
			'IBM Plex Mono',
			Consolas,
			monospace;
	}

	.lede {
		max-width: 41rem;
		margin: 1.4rem 0 0;
		font-size: clamp(1.05rem, 1.5vw, 1.26rem);
	}

	.hero-actions {
		flex-wrap: wrap;
		margin-top: 1.6rem;
	}

	.button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-height: 2.9rem;
		padding: 0.75rem 1.05rem;
		border-radius: 999px;
		border: 1px solid rgba(255, 255, 255, 0.12);
		font-weight: 700;
	}

	.button.primary {
		background: linear-gradient(135deg, #8af1ff 0%, #b9ffd2 100%);
		color: #07111f;
		border-color: transparent;
	}

	.button.ghost {
		background: rgba(255, 255, 255, 0.04);
	}

	.hero-panel,
	.code-block,
	.steps article,
	.example-grid article,
	.comparison-grid article,
	.recipe-grid article,
	.contribute-list div {
		border: 1px solid rgba(145, 184, 220, 0.16);
		border-radius: 8px;
		background: rgba(13, 25, 43, 0.84);
		box-shadow:
			0 24px 60px rgba(0, 0, 0, 0.22),
			inset 0 1px 0 rgba(255, 255, 255, 0.03);
	}

	.hero-panel {
		overflow: hidden;
	}

	.code-block {
		overflow: hidden;
		background: #07111f;
	}

	.panel-title {
		padding: 0.85rem 1rem;
		border-bottom: 1px solid rgba(145, 184, 220, 0.12);
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: #9fb5d4;
		font-size: 0.78rem;
		font-weight: 800;
	}

	.code-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.48rem;
		padding: 0.78rem 0.95rem;
		border-bottom: 1px solid rgba(145, 184, 220, 0.12);
		background: rgba(255, 255, 255, 0.035);
		color: #9fb5d4;
	}

	.code-title,
	.copyable-title {
		display: flex;
		align-items: center;
		gap: 0.48rem;
		min-width: 0;
	}

	.copyable-title {
		justify-content: space-between;
	}

	.code-title span {
		width: 0.58rem;
		height: 0.58rem;
		flex: 0 0 auto;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.18);
	}

	.code-title strong {
		margin-left: 0.25rem;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		font-size: 0.74rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.copy-button {
		flex: 0 0 auto;
		min-height: 1.85rem;
		padding: 0.35rem 0.62rem;
		border: 1px solid rgba(138, 241, 255, 0.24);
		border-radius: 8px;
		background: rgba(138, 241, 255, 0.08);
		color: #dff9ff;
		font: inherit;
		font-size: 0.74rem;
		font-weight: 800;
		letter-spacing: 0;
		text-transform: none;
		cursor: pointer;
	}

	.copy-button:hover,
	.copy-button:focus-visible {
		border-color: rgba(138, 241, 255, 0.48);
		background: rgba(138, 241, 255, 0.14);
		outline: none;
	}

	.copy-button.copied {
		background: rgba(185, 255, 210, 0.14);
		border-color: rgba(185, 255, 210, 0.44);
		color: #dfffe9;
	}

	pre {
		margin: 0;
		padding: 1rem;
		overflow-x: auto;
		color: #dce9f8;
		font-family:
			'SFMono-Regular',
			'JetBrains Mono',
			'IBM Plex Mono',
			Consolas,
			monospace;
		font-size: 0.86rem;
		line-height: 1.62;
		white-space: pre;
	}

	.code-block pre {
		background:
			linear-gradient(90deg, rgba(138, 241, 255, 0.045) 0 1px, transparent 1px 100%),
			linear-gradient(180deg, rgba(8, 17, 29, 0.98) 0%, rgba(5, 11, 20, 0.98) 100%);
		background-size: 3.4rem 100%, auto;
	}

	.code-block code {
		display: block;
		color: #e4f1ff;
	}

	.section {
		padding-block: clamp(2rem, 6vw, 4.5rem);
		border-top: 1px solid rgba(145, 184, 220, 0.1);
		scroll-margin-top: 1rem;
	}

	.section-head {
		display: grid;
		gap: 0.7rem;
		margin-bottom: 1.4rem;
		max-width: 45rem;
	}

	.section-head.wide {
		max-width: 58rem;
	}

	.steps,
	.example-grid,
	.comparison-grid,
	.recipe-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
	}

	.steps article,
	.example-grid article,
	.comparison-grid article,
	.recipe-grid article {
		padding: 1.1rem;
	}

	.command-block {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: stretch;
		overflow: hidden;
		margin-top: 0.9rem;
		border: 1px solid rgba(145, 184, 220, 0.12);
		border-radius: 8px;
		background: #07111f;
	}

	.command-block pre {
		min-width: 0;
		padding: 0.75rem 0.85rem;
	}

	.command-block .copy-button {
		margin: 0.45rem 0.45rem 0.45rem 0;
	}

	.example-grid strong,
	.comparison-grid strong {
		display: block;
		color: #e4f1ff;
		font-size: 1.08rem;
	}

	.example-grid code {
		display: inline-block;
		margin-top: 0.2rem;
		padding: 0.35rem 0.5rem;
		border-radius: 8px;
		background: rgba(138, 241, 255, 0.08);
		color: #dff9ff;
		font-size: 0.78rem;
	}

	.text-link {
		display: inline-flex;
		width: fit-content;
		color: #dff9ff;
		font-weight: 800;
		text-decoration: underline;
		text-underline-offset: 0.3rem;
	}

	.steps span {
		display: block;
		margin-bottom: 0.9rem;
		color: #8adcee;
		font-size: 0.78rem;
		font-weight: 800;
		letter-spacing: 0.12em;
	}

	.reference-grid {
		display: grid;
		grid-template-columns: minmax(13rem, 0.36fr) minmax(0, 1fr);
		gap: 1.1rem;
		align-items: start;
		padding-block: 1.25rem;
		border-top: 1px solid rgba(145, 184, 220, 0.1);
	}

	.reference-copy {
		position: sticky;
		top: 1rem;
	}

	.table-wrap {
		overflow-x: auto;
		border: 1px solid rgba(145, 184, 220, 0.14);
		border-radius: 8px;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		min-width: 43rem;
		background: rgba(9, 18, 31, 0.76);
	}

	th,
	td {
		padding: 0.82rem 0.9rem;
		border-bottom: 1px solid rgba(145, 184, 220, 0.1);
		text-align: left;
		vertical-align: top;
		line-height: 1.55;
	}

	th {
		color: #8adcee;
		font-size: 0.78rem;
		text-transform: uppercase;
		letter-spacing: 0.12em;
	}

	td {
		color: #cfddf0;
	}

	td:first-child,
	td:nth-child(2) {
		white-space: nowrap;
	}

	.recipe-grid {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.recipe-grid article:nth-child(5),
	.recipe-grid article:nth-child(6) {
		grid-column: span 1;
	}

	.pill-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.45rem;
		margin-top: 0.8rem;
	}

	.pill-list code {
		padding: 0.35rem 0.5rem;
		border-radius: 999px;
		background: rgba(138, 241, 255, 0.08);
		color: #dff9ff;
		font-size: 0.78rem;
	}

	.contribute-list {
		display: grid;
		gap: 0.85rem;
	}

	.contribute-list div {
		padding: 1rem;
	}

	.project-footer {
		display: grid;
		grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
		gap: 1.25rem;
		padding-block: clamp(2rem, 6vw, 4rem);
		border-top: 1px solid rgba(145, 184, 220, 0.1);
	}

	.project-footer h2 {
		margin-bottom: 0.8rem;
	}

	.project-footer p {
		max-width: 42rem;
	}

	.footer-links {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.65rem;
	}

	.footer-links a {
		padding: 0.78rem 0.85rem;
		border: 1px solid rgba(145, 184, 220, 0.12);
		border-radius: 8px;
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

	.contribute-list strong {
		display: block;
		margin-bottom: 0.35rem;
	}

	@media (max-width: 980px) {
		.topbar,
		.hero-grid,
		.split,
		.project-footer,
		.reference-grid,
		.steps,
		.example-grid,
		.comparison-grid,
		.recipe-grid {
			grid-template-columns: 1fr;
		}

		.topbar {
			align-items: flex-start;
			flex-direction: column;
		}

		.nav-links {
			justify-content: flex-start;
		}

		.reference-copy {
			position: static;
		}

		.footer-links {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 640px) {
		h1 {
			font-size: clamp(2.7rem, 16vw, 4rem);
		}

		.nav-links {
			display: none;
		}

		pre {
			font-size: 0.78rem;
		}
	}
</style>
