import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export type Language = 'ja' | 'en';

const storageKey = 'specter-language';
const initialLanguage: Language =
	browser && localStorage.getItem(storageKey) === 'en' ? 'en' : 'ja';

export const language = writable<Language>(initialLanguage);

export function setLanguage(nextLanguage: Language) {
	language.set(nextLanguage);
}

if (browser) {
	language.subscribe((currentLanguage) => {
		localStorage.setItem(storageKey, currentLanguage);
		document.documentElement.lang = currentLanguage;
	});
}
