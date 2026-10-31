import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import tseslint from 'typescript-eslint';
import globals from 'globals';

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...svelte.configs.recommended,
  {
    languageOptions: {
      globals: { ...globals.browser, ...globals.node },
      parserOptions: { extraFileExtensions: ['.svelte'] }
    }
  },
  {
    files: ['**/*.svelte'],
    languageOptions: { parserOptions: { parser: tseslint.parser } }
  },
  {
    // Runes modules: .svelte.ts is TypeScript, not Svelte markup.
    files: ['**/*.svelte.ts'],
    languageOptions: { parser: tseslint.parser }
  },
  {
    ignores: ['build/', '.svelte-kit/', 'node_modules/', 'test-results/', 'playwright-report/']
  }
);
