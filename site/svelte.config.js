import adapter from '@sveltejs/adapter-static';

const isDev = process.argv.includes('dev');

/** @type {import('@sveltejs/kit').Config} */
const config = {
	vitePlugin: {
		dynamicCompileOptions: ({ filename, compileOptions }) => {
			if (!filename.split(/[/\\]/).includes('node_modules') && !compileOptions.runes) {
				return { runes: true };
			}
		}
	},
	kit: {
		adapter: adapter({
			fallback: '200.html'
		}),
		paths: {
			base: isDev ? '' : '/specter'
		}
	}
};

export default config;
