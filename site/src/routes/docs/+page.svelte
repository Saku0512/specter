<script lang="ts">
	import mark from '$lib/assets/logo-icon.png';

	const toc = [
		{ href: '#introduction', label: 'Introduction' },
		{ href: '#quick-start', label: 'Quick Start' },
		{ href: '#config', label: 'config.yml' },
		{ href: '#recipes', label: 'Recipes' },
		{ href: '#cli', label: 'CLI' },
		{ href: '#contributing', label: 'Contributing' }
	];

	const topLevelFields = [
		['cors', 'boolean', 'Enables CORS headers and handles OPTIONS preflight requests.'],
		['proxy', 'string', 'Forwards unmatched requests to a real backend.'],
		['openapi', 'string', 'Path to an OpenAPI YAML or JSON file for request and response validation.'],
		['openapi_strict', 'boolean', 'When true, invalid requests return 400 instead of only adding a warning header.'],
		[
			'openapi_strict_response',
			'boolean',
			'When true, invalid mock responses return 500 instead of being served with a warning header.'
		],
		['include', 'list', 'Merges routes from other YAML files. Glob patterns are supported.'],
		['routes', 'list', 'The route definitions served by specter.']
	];

	const routeFields = [
		['path', 'string', 'Required. URL path. Supports :param path parameters.'],
		['method', 'string', 'Required. HTTP method such as GET, POST, PUT, PATCH, or DELETE.'],
		['status', 'int', 'Response status code. Defaults to 200.'],
		['response', 'any', 'Inline JSON object, array, scalar, or string response body.'],
		['headers', 'map', 'Custom response headers for the route.'],
		['content_type', 'string', 'Response MIME type. Defaults to application/json.'],
		['delay', 'int', 'Fixed response delay in milliseconds.'],
		['delay_min', 'int', 'Minimum random delay in milliseconds. Use with delay_max.'],
		['delay_max', 'int', 'Maximum random delay in milliseconds. Use with delay_min.'],
		['error_rate', 'float', 'Probability from 0.0 to 1.0 that the route returns an injected error.'],
		['error_status', 'int', 'Status code for injected errors. Defaults to 503.'],
		['on_call', 'int', 'Only match this route on a specific 1-based call number.'],
		['match', 'list', 'Conditional responses based on query, headers, body, form data, cookies, or GraphQL.'],
		['mode', 'string', 'Controls responses selection. Use sequential or random.'],
		['responses', 'list', 'Multiple response entries for cycling, retry simulation, or random behavior.'],
		['rate_limit', 'int', 'Maximum requests before returning 429.'],
		['rate_reset', 'int', 'Seconds until the rate-limit counter resets. Adds Retry-After on 429.'],
		['state', 'string', 'Only match when the server state equals this value.'],
		['set_state', 'string', 'Set the server state after responding.'],
		['vars', 'map', 'Only match when all named variables equal these values.'],
		['set_vars', 'map', 'Set named variables after responding. Values can use templates.'],
		['webhook', 'object', 'Fire an outgoing callback after the response.'],
		['file', 'string', 'Serve a response body from a JSON, YAML, or text fixture file.'],
		['script', 'string', 'Go template that generates the response body. Takes priority over file and response.'],
		['proxy', 'string', 'Forward this route to a real backend. Takes priority over mock response fields.'],
		['store_push', 'string', 'Push the request body into an in-memory store and respond 201.'],
		['store_list', 'string', 'List all items in an in-memory store with filtering, sorting, and pagination.'],
		['store_get', 'string', 'Get one store item by the store_key path parameter.'],
		['store_put', 'string', 'Replace or upsert one store item.'],
		['store_patch', 'string', 'Merge the request body into one store item.'],
		['store_delete', 'string', 'Delete one store item.'],
		['store_clear', 'string', 'Delete every item in a named store.'],
		['store_key', 'string', 'Path parameter used as the item ID. Defaults to id.'],
		['stream', 'boolean', 'Respond with a Server-Sent Events stream.'],
		['events', 'list', 'Ordered SSE events for a stream route.'],
		['stream_repeat', 'boolean', 'Repeat SSE events until the client disconnects.'],
		['set_cookies', 'list', 'Set cookies in the response.'],
		['redirect', 'string', 'Redirect to another path or URL.'],
		['redirect_status', 'int', 'Redirect status code. Use 301, 302, 303, 307, or 308.']
	];

	const matchFields = [
		['query', 'map', 'Match query parameters using Go regular expressions.'],
		['headers', 'map', 'Match request headers. Header names are case-insensitive.'],
		['body', 'map', 'Match top-level JSON request body fields.'],
		['body_path', 'map', 'Match nested JSON fields with dot notation such as user.role.'],
		['form', 'map', 'Match application/x-www-form-urlencoded request bodies.'],
		['graphql', 'object', 'Match GraphQL operationName and variables.'],
		['cookies', 'map', 'Match request cookies with regex patterns.'],
		['status', 'int', 'Status code returned when this match fires.'],
		['response', 'any', 'Response body returned when this match fires.'],
		['response_headers', 'map', 'Headers applied only for this match. Overrides route headers.'],
		['content_type', 'string', 'Content type applied only for this match.'],
		['delay', 'int', 'Extra delay in milliseconds for this match. Added after route delay.'],
		['set_state', 'string', 'State transition applied only when this match fires.'],
		['set_vars', 'map', 'Variable updates applied only when this match fires.'],
		['file', 'string', 'Fixture file returned only when this match fires.'],
		['script', 'string', 'Template response returned only when this match fires.']
	];

	const responseFields = [
		['on_call', 'int', 'Pin this response entry to a specific call number.'],
		['status', 'int', 'Status for this response entry.'],
		['response', 'any', 'Inline body for this response entry.'],
		['content_type', 'string', 'Content type for this response entry.'],
		['delay', 'int', 'Delay for this response entry.'],
		['file', 'string', 'Fixture file for this response entry.'],
		['script', 'string', 'Template body for this response entry.']
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
</script>

<svelte:head>
	<title>specter docs | Introduction, Quick Start, config.yml, Contributing</title>
	<meta
		name="description"
		content="Documentation for specter, including introduction, quick start, complete config.yml reference, CLI usage, and contribution guide."
	/>
</svelte:head>

<div class="docs-page">
	<header class="docs-hero" id="introduction">
		<nav class="topbar" aria-label="Main navigation">
			<a class="brand" href="/">
				<img src={mark} alt="" />
				<span>specter</span>
			</a>
			<div class="nav-links">
				{#each toc as item}
					<a href={item.href}>{item.label}</a>
				{/each}
			</div>
		</nav>

		<div class="hero-grid">
			<div>
				<p class="kicker">Documentation</p>
				<h1>Build reliable mock APIs from one YAML file.</h1>
				<p class="lede">
					specter is a lightweight mock API server for frontend development, demos, automated
					tests, and API contract work. Define routes in <code>config.yml</code>, run a single
					binary, and adjust behavior without rebuilding your application.
				</p>
				<div class="hero-actions">
					<a class="button primary" href="#quick-start">Start Quickly</a>
					<a class="button ghost" href="#config">Config Reference</a>
				</div>
			</div>

			{@render codeBlock(basicConfig)}
		</div>
	</header>

	<main>
		<section class="section" id="quick-start">
			<div class="section-head">
				<p class="kicker">Quick Start</p>
				<h2>From blank folder to mock API</h2>
			</div>

			<div class="steps">
				<article>
					<span>01</span>
					<h3>Generate a starter file</h3>
					<pre>specter init</pre>
					<p>Creates <code>config.yml</code> in the current directory.</p>
				</article>
				<article>
					<span>02</span>
					<h3>Run the server</h3>
					<pre>specter -c config.yml</pre>
					<p>The API listens on port <code>8080</code> and the control UI opens on port <code>4444</code>.</p>
				</article>
				<article>
					<span>03</span>
					<h3>Call a route</h3>
					<pre>curl http://localhost:8080/users</pre>
					<p>Edit the YAML and save. specter reloads the config automatically.</p>
				</article>
			</div>
		</section>

		<section class="section" id="config">
			<div class="section-head wide">
				<p class="kicker">config.yml</p>
				<h2>Complete reference</h2>
				<p>
					A config can stay tiny with one route, or grow into a full scenario with state,
					conditional matching, fixtures, OpenAPI validation, proxying, stores, callbacks, delays,
					and streams.
				</p>
			</div>

			<div class="reference-grid">
				<div class="reference-copy">
					<h3>Top-level fields</h3>
					<p>Use these fields once at the root of the YAML file.</p>
				</div>
				{@render table(topLevelFields)}
			</div>

			<div class="reference-grid">
				<div class="reference-copy">
					<h3>Route fields</h3>
					<p>Each item in <code>routes</code> describes one mock, proxy, store operation, redirect, or stream.</p>
				</div>
				{@render table(routeFields)}
			</div>

			<div class="reference-grid">
				<div class="reference-copy">
					<h3>Match fields</h3>
					<p>
						Use <code>match</code> when one method and path should branch based on request data.
						All conditions in one match entry must pass.
					</p>
				</div>
				{@render table(matchFields)}
			</div>

			<div class="reference-grid">
				<div class="reference-copy">
					<h3>Response entries</h3>
					<p>
						Use <code>responses</code> with <code>mode: sequential</code> or <code>mode: random</code>
						to simulate retries, polling, flaky APIs, or changing data.
					</p>
				</div>
				{@render table(responseFields)}
			</div>
		</section>

		<section class="section" id="recipes">
			<div class="section-head">
				<p class="kicker">Recipes</p>
				<h2>Common config patterns</h2>
			</div>

			<div class="recipe-grid">
				<article>
					<h3>Conditional responses</h3>
					<p>Branch by request body, query parameters, headers, cookies, form data, or GraphQL values.</p>
					{@render codeBlock(matchingConfig)}
				</article>

				<article>
					<h3>Stateful flows</h3>
					<p>Use <code>state</code>, <code>set_state</code>, <code>vars</code>, and <code>set_vars</code> for login flows and scenario gates.</p>
					{@render codeBlock(stateConfig)}
				</article>

				<article>
					<h3>In-memory CRUD</h3>
					<p>Wire REST endpoints directly to a named store. Store data resets when the server restarts.</p>
					{@render codeBlock(storeConfig)}
					{@render codeBlock(storeQuery, 'request')}
				</article>

				<article>
					<h3>OpenAPI, proxy, chaos, and SSE</h3>
					<p>Mix real services with mocks, validate requests, add jitter, inject failures, and stream events.</p>
					{@render codeBlock(advancedConfig)}
				</article>

				<article>
					<h3>Templates and faker</h3>
					<p>
						Templates can read <code>.body</code>, <code>.query</code>, <code>.params</code>,
						<code>.headers</code>, <code>.method</code>, and <code>.path</code>.
					</p>
					{@render codeBlock(scriptConfig)}
					<div class="pill-list" aria-label="Faker types">
						{#each fakerTypes as type}
							<code>{type}</code>
						{/each}
					</div>
				</article>

				<article>
					<h3>Fixtures, redirects, cookies, and webhooks</h3>
					<p>
						Use <code>file</code> for large JSON/YAML/text fixtures, <code>redirect</code> for
						HTTP redirects, <code>set_cookies</code> for auth simulations, and <code>webhook</code>
						for async callbacks.
					</p>
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
				<h2>Useful commands</h2>
				<p>
					Flags override environment variables. The default API port is <code>8080</code>, and
					the built-in dashboard uses <code>4444</code>. Set <code>--ui-port 0</code> to disable it.
				</p>
			</div>
			<div class="hero-panel">
				<div class="panel-title">commands</div>
				<pre>{cliCommands}</pre>
			</div>
		</section>

		<section class="section split" id="contributing">
			<div>
				<p class="kicker">Contributing</p>
				<h2>Help improve specter</h2>
				<p>
					Contributions are welcome across docs, examples, bug fixes, CLI behavior, validation,
					UI improvements, and new mock-server features. Keep changes focused, add tests for
					behavioral changes, and update documentation when config or CLI behavior changes.
				</p>
			</div>

			<div class="contribute-list">
				<div>
					<strong>1. Fork and branch</strong>
					<p>Create a branch with a focused name, then make the smallest useful change.</p>
				</div>
				<div>
					<strong>2. Validate locally</strong>
					<p>Run the relevant tests and use <code>specter validate -c config.yml</code> for docs examples.</p>
				</div>
				<div>
					<strong>3. Open a pull request</strong>
					<p>Describe the scenario, what changed, and any compatibility notes for existing configs.</p>
				</div>
			</div>
		</section>
	</main>
</div>

{#snippet table(rows: string[][])}
	<div class="table-wrap">
		<table>
			<thead>
				<tr>
					<th>Field</th>
					<th>Type</th>
					<th>Description</th>
				</tr>
			</thead>
			<tbody>
				{#each rows as row}
					<tr>
						<td><code>{row[0]}</code></td>
						<td>{row[1]}</td>
						<td>{row[2]}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/snippet}

{#snippet codeBlock(code: string, label = 'config.yml')}
	<div class="code-block">
		<div class="code-head">
			<span></span>
			<span></span>
			<span></span>
			<strong>{label}</strong>
		</div>
		<pre><code>{code}</code></pre>
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
	main {
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
		gap: 0.48rem;
		padding: 0.78rem 0.95rem;
		border-bottom: 1px solid rgba(145, 184, 220, 0.12);
		background: rgba(255, 255, 255, 0.035);
		color: #9fb5d4;
	}

	.code-head span {
		width: 0.58rem;
		height: 0.58rem;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.18);
	}

	.code-head strong {
		margin-left: 0.25rem;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		font-size: 0.74rem;
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
	.recipe-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
	}

	.steps article,
	.recipe-grid article {
		padding: 1.1rem;
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

	.contribute-list strong {
		display: block;
		margin-bottom: 0.35rem;
	}

	@media (max-width: 980px) {
		.topbar,
		.hero-grid,
		.split,
		.reference-grid,
		.steps,
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
