import { readdirSync } from 'node:fs';
import { join } from 'node:path';
import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import importX from 'eslint-plugin-import-x';
import { createTypeScriptImportResolver } from 'eslint-import-resolver-typescript';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';

const appSrc = './graph-frontapp/src';
const features = readdirSync(join(import.meta.dirname, appSrc, 'features'), {
  withFileTypes: true,
})
  .filter((entry) => entry.isDirectory())
  .map((entry) => entry.name);

export default tseslint.config(
  { ignores: ['**/dist/**', '**/node_modules/**', '**/coverage/**', '**/*.d.ts'] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'warn',
    },
  },
  {
    files: ['graph-frontapp/src/**/*.{ts,tsx}'],
    plugins: { 'import-x': importX },
    settings: {
      'import-x/resolver-next': [
        createTypeScriptImportResolver({ project: 'graph-frontapp/tsconfig.json' }),
      ],
    },
    rules: {
      'import-x/no-restricted-paths': [
        'error',
        {
          basePath: import.meta.dirname,
          zones: [
            ...features.map((feature) => ({
              target: `${appSrc}/features/${feature}`,
              from: `${appSrc}/features`,
              except: [`./${feature}`],
              message: 'Features exchange state through URL params and the query cache.',
            })),
            {
              target: `${appSrc}/shared`,
              from: `${appSrc}/features`,
              message: 'Shared code does not depend on features.',
            },
          ],
        },
      ],
    },
  }
);
