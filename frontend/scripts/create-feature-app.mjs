#!/usr/bin/env node
/* eslint-disable no-console */
/**
 * Scaffolds a new feature-app under `apps/<name>-app/` and wires it
 * into main-app + the root workspace.
 *
 * Usage (from `frontend/`):
 *
 *   pnpm new-app <name> [--port=NNNN] [--no-install]
 *
 * `<name>` must be kebab-case (`reports`, `tasks`, `audit-log`).
 * The `-app` suffix is appended automatically — package name becomes
 * `@els/<name>-app`, PascalCase routes export becomes
 * `<Pascal>AppRoutes`.
 *
 * What it touches:
 *   apps/<name>-app/                       — generated from template
 *   apps/main-app/package.json             — adds workspace dep
 *   apps/main-app/src/App.tsx              — lazy() import + <Route>
 *   apps/main-app/src/index.css            — @source line
 *   package.json                           — dev:<name> script
 *
 * Idempotent-ish: refuses to overwrite an existing apps/<name>-app/.
 * Patches to main-app are best-effort; if an anchor is missing, the
 * step prints a warning and continues — feature still works in
 * isolated dev (`pnpm dev:<name>`).
 */

import { existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { execSync } from 'node:child_process'

const __dirname = dirname(fileURLToPath(import.meta.url))
const FRONTEND_ROOT = resolve(__dirname, '..')
const APPS_DIR = join(FRONTEND_ROOT, 'apps')
const MAIN_APP = join(APPS_DIR, 'main-app')

// ---------- args ----------

const args = process.argv.slice(2)
let nameArg = null
let portArg = null
let runInstall = true

for (const arg of args) {
  if (arg === '--help' || arg === '-h') {
    console.log(usage())
    process.exit(0)
  } else if (arg === '--no-install') runInstall = false
  else if (arg.startsWith('--port=')) portArg = Number(arg.slice('--port='.length))
  else if (arg.startsWith('--')) die(`unknown flag: ${arg}`)
  else if (!nameArg) nameArg = arg
  else die(`unexpected positional arg: ${arg}`)
}

if (!nameArg) {
  console.log(usage())
  process.exit(1)
}

if (!/^[a-z][a-z0-9-]*$/.test(nameArg)) {
  die(`invalid name "${nameArg}". Use kebab-case: lowercase letters, digits, hyphens. Example: "reports", "audit-log".`)
}

if (nameArg.endsWith('-app')) {
  die(`don't append "-app" yourself — use "${nameArg.replace(/-app$/, '')}", we add the suffix.`)
}

// ---------- naming ----------

const slug = nameArg
const pkgDir = `${slug}-app`
const pkgName = `@els/${slug}-app`
const pascal = toPascalCase(slug)
const routesExport = `${pascal}AppRoutes`
const pageName = `${pascal}AppPage`
const titleCase = toTitleCase(slug)

const targetDir = join(APPS_DIR, pkgDir)
if (existsSync(targetDir)) {
  die(`apps/${pkgDir} already exists. Pick a different name or delete the directory.`)
}

// ---------- port ----------

const port = portArg ?? pickFreePort()
if (!Number.isInteger(port) || port < 1024 || port > 65535) {
  die(`invalid port ${port}`)
}

console.log(`▸ creating ${pkgName}`)
console.log(`  dir   : ${relative(FRONTEND_ROOT, targetDir)}`)
console.log(`  port  : ${port}`)
console.log(`  routes: ${routesExport}`)

// ---------- write files ----------

const files = templateFiles({ slug, pkgDir, pkgName, pascal, routesExport, pageName, titleCase, port })
for (const [rel, contents] of Object.entries(files)) {
  const abs = join(targetDir, rel)
  mkdirSync(dirname(abs), { recursive: true })
  writeFileSync(abs, contents, 'utf8')
}
console.log(`✓ scaffolded ${Object.keys(files).length} files in apps/${pkgDir}/`)

// ---------- patch main-app + root ----------

patchMainAppPkgJson()
patchMainAppCss()
patchMainAppRoutes()
patchRootPkgJson()

// ---------- finish ----------

if (runInstall) {
  console.log('▸ running pnpm install')
  try {
    execSync('pnpm install', { cwd: FRONTEND_ROOT, stdio: 'inherit' })
  } catch {
    console.warn('⚠ pnpm install failed — run it yourself when ready')
  }
} else {
  console.log('▸ skipping pnpm install (--no-install). Run it yourself before `pnpm dev`.')
}

console.log('')
console.log('Next steps:')
console.log(`  pnpm dev:${slug}            # isolated dev on http://localhost:${port}`)
console.log(`  pnpm dev                    # full main-app, feature mounted at /v1/${slug}`)
console.log('')
console.log('Note: sidebar entries come from the backend (account.apps endpoint),')
console.log('not from the frontend. Add a row there if you want a sidebar tile.')

// ============================================================
//                          helpers
// ============================================================

function usage() {
  return `Usage: pnpm new-app <name> [--port=NNNN] [--no-install]

Scaffolds apps/<name>-app/ and wires it into main-app.

Arguments:
  <name>          kebab-case feature name without "-app" suffix
                  (e.g. "reports", "tasks", "audit-log")

Options:
  --port=NNNN     dev server port (default: max(existing) + 1)
  --no-install    skip "pnpm install" at the end
`
}

function die(msg) {
  console.error(`✗ ${msg}`)
  process.exit(1)
}

function toPascalCase(kebab) {
  return kebab
    .split('-')
    .filter(Boolean)
    .map((p) => p[0].toUpperCase() + p.slice(1))
    .join('')
}

function toTitleCase(kebab) {
  return kebab
    .split('-')
    .filter(Boolean)
    .map((p) => p[0].toUpperCase() + p.slice(1))
    .join(' ')
}

function pickFreePort() {
  // Baseline = main-app's typical port. If no apps had `port: NNNN` in
  // their vite.config we still return something sane (5174).
  let max = 5173
  for (const dir of safeReadDir(APPS_DIR)) {
    const cfg = join(APPS_DIR, dir, 'vite.config.ts')
    if (!existsSync(cfg)) continue
    for (const m of readFileSync(cfg, 'utf8').matchAll(/port:\s*(\d+)/g)) {
      max = Math.max(max, Number(m[1]))
    }
  }
  return max + 1
}

function safeReadDir(p) {
  try {
    return readdirSync(p, { withFileTypes: true }).filter((d) => d.isDirectory()).map((d) => d.name)
  } catch {
    return []
  }
}

function patchMainAppPkgJson() {
  const path = join(MAIN_APP, 'package.json')
  const json = JSON.parse(readFileSync(path, 'utf8'))
  json.dependencies ??= {}
  if (json.dependencies[pkgName]) {
    console.log('  · main-app/package.json already has the dep, skipping')
    return
  }
  json.dependencies[pkgName] = 'workspace:*'
  json.dependencies = sortObject(json.dependencies)
  writeFileSync(path, JSON.stringify(json, null, 2) + '\n', 'utf8')
  console.log('✓ patched main-app/package.json')
}

function sortObject(obj) {
  return Object.fromEntries(Object.entries(obj).sort(([a], [b]) => a.localeCompare(b)))
}

function patchMainAppCss() {
  const path = join(MAIN_APP, 'src/index.css')
  const css = readFileSync(path, 'utf8')
  const marker = `@source "../../${pkgDir}/src/**/*.{ts,tsx}";`
  if (css.includes(marker)) {
    console.log('  · main-app/src/index.css already has @source, skipping')
    return
  }
  // Insert after the last @source "../../<feature>/..." line.
  const lines = css.split('\n')
  let lastIdx = -1
  for (let i = 0; i < lines.length; i++) {
    if (/^@source\s+"\.\.\/\.\.\/[^/]+\/src/.test(lines[i])) lastIdx = i
  }
  if (lastIdx === -1) {
    console.warn('⚠ main-app/src/index.css: no existing @source line found, append manually:')
    console.warn(`    ${marker}`)
    return
  }
  lines.splice(lastIdx + 1, 0, marker)
  writeFileSync(path, lines.join('\n'), 'utf8')
  console.log('✓ patched main-app/src/index.css')
}

function patchMainAppRoutes() {
  const path = join(MAIN_APP, 'src/App.tsx')
  let src = readFileSync(path, 'utf8')

  if (src.includes(`from '${pkgName}'`)) {
    console.log('  · main-app/src/App.tsx already imports the feature, skipping')
    return
  }

  // 1. Insert lazy() import after the last "const XxxRoutes = lazy(() =>" block.
  const lazyBlock = `\nconst ${routesExport} = lazy(() =>\n  import('${pkgName}').then((m) => ({ default: m.${routesExport} })),\n)`
  const lazyAnchor = /const\s+\w+Routes\s*=\s*lazy\(\(\)\s*=>\s*\n\s*import\([^)]+\)\.then\(\(m\)\s*=>\s*\(\{\s*default:\s*m\.\w+Routes\s*\}\)\),\s*\n\)/g
  let match
  let lastEnd = -1
  while ((match = lazyAnchor.exec(src)) !== null) lastEnd = match.index + match[0].length
  if (lastEnd === -1) {
    console.warn('⚠ main-app/src/App.tsx: no existing lazy() import found, add manually:')
    console.warn(`    ${lazyBlock.trim()}`)
  } else {
    src = src.slice(0, lastEnd) + '\n' + lazyBlock + src.slice(lastEnd)
  }

  // 2. Insert <Route path="v1/<slug>/*"> before the trailing wildcard.
  const routeBlock = `            <Route\n              path="v1/${slug}/*"\n              element={\n                <Suspense fallback={<AppLoader />}>\n                  <${routesExport} />\n                </Suspense>\n              }\n            />`
  const wildcardRe = /( *)<Route\s+path="\*"\s+element=\{<Navigate[^}]*\/>\}\s*\/>/
  if (wildcardRe.test(src)) {
    src = src.replace(wildcardRe, (m, indent) => `${routeBlock}\n${indent}${m.trimStart()}`)
  } else {
    console.warn('⚠ main-app/src/App.tsx: wildcard <Route path="*"> not found, add manually:')
    console.warn(routeBlock)
  }

  writeFileSync(path, src, 'utf8')
  console.log('✓ patched main-app/src/App.tsx')
}

function patchRootPkgJson() {
  const path = join(FRONTEND_ROOT, 'package.json')
  const json = JSON.parse(readFileSync(path, 'utf8'))
  json.scripts ??= {}
  const key = `dev:${slug}`
  if (json.scripts[key]) {
    console.log(`  · root package.json already has scripts.${key}, skipping`)
    return
  }
  json.scripts[key] = `pnpm --filter ${pkgName} dev`
  json.scripts = sortDevScripts(json.scripts)
  writeFileSync(path, JSON.stringify(json, null, 2) + '\n', 'utf8')
  console.log('✓ patched package.json (dev:' + slug + ')')
}

function sortDevScripts(scripts) {
  // Keep "dev", "dev:*" together (sorted by key), preserve order of others.
  const dev = []
  const rest = []
  for (const [k, v] of Object.entries(scripts)) {
    if (k === 'dev' || k.startsWith('dev:')) dev.push([k, v])
    else rest.push([k, v])
  }
  dev.sort(([a], [b]) => {
    if (a === 'dev') return -1
    if (b === 'dev') return 1
    return a.localeCompare(b)
  })
  return Object.fromEntries([...dev, ...rest])
}

// ============================================================
//                       file templates
// ============================================================

function templateFiles(ctx) {
  return {
    'package.json': pkgJson(ctx),
    'vite.config.ts': viteConfig(ctx),
    'tsconfig.json': tsconfigRoot(),
    'tsconfig.app.json': tsconfigApp(),
    'tsconfig.node.json': tsconfigNode(),
    'eslint.config.js': eslintConfig(),
    'index.html': indexHtml(ctx),
    'src/index.ts': srcIndex(ctx),
    'src/routes.tsx': srcRoutes(ctx),
    [`src/${ctx.pageName}.tsx`]: srcPage(ctx),
    'src/dev.tsx': srcDev(ctx),
    'src/index.css': srcCss(),
    'src/vite-env.d.ts': '/// <reference types="vite/client" />\n',
    'src/lib/api.ts': srcApi(),
  }
}

function pkgJson({ pkgName }) {
  const j = {
    name: pkgName,
    private: true,
    version: '0.0.0',
    type: 'module',
    exports: { '.': './src/index.ts' },
    scripts: {
      dev: 'vite',
      build: 'tsc -b && vite build',
      preview: 'vite preview',
      lint: 'eslint .',
      typecheck: 'tsc -b',
    },
    dependencies: {
      '@els/api-client': 'workspace:*',
      '@els/dev-harness': 'workspace:*',
      '@els/ui': 'workspace:*',
      '@tanstack/react-query': 'catalog:',
      'lucide-react': 'catalog:',
      react: 'catalog:',
      'react-dom': 'catalog:',
      'react-router-dom': 'catalog:',
    },
    devDependencies: {
      '@eslint/js': 'catalog:',
      '@tailwindcss/vite': 'catalog:',
      '@types/node': 'catalog:',
      '@types/react': 'catalog:',
      '@types/react-dom': 'catalog:',
      '@vitejs/plugin-react': 'catalog:',
      eslint: 'catalog:',
      'eslint-plugin-react-hooks': 'catalog:',
      'eslint-plugin-react-refresh': 'catalog:',
      globals: 'catalog:',
      tailwindcss: 'catalog:',
      typescript: 'catalog:',
      'typescript-eslint': 'catalog:',
      vite: 'catalog:',
    },
  }
  return JSON.stringify(j, null, 2) + '\n'
}

function viteConfig({ port }) {
  return `import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: ${port},
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
`
}

function tsconfigRoot() {
  return `{
  "files": [],
  "references": [
    { "path": "./tsconfig.app.json" },
    { "path": "./tsconfig.node.json" }
  ]
}
`
}

function tsconfigApp() {
  return `{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "composite": true,
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.app.tsbuildinfo",
    "rootDir": "./src",
    "paths": {
      "@/*": ["./src/*"]
    },
    "types": ["vite/client"]
  },
  "include": ["src"]
}
`
}

function tsconfigNode() {
  return `{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "composite": true,
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.node.tsbuildinfo",
    "types": ["node"]
  },
  "include": ["vite.config.ts"]
}
`
}

function eslintConfig() {
  return `import js from '@eslint/js'
import globals from 'globals'
import tseslint from 'typescript-eslint'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'

export default tseslint.config(
  { ignores: ['dist', 'node_modules'] },
  {
    files: ['**/*.{ts,tsx}'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
    },
  },
)
`
}

function indexHtml({ titleCase }) {
  return `<!doctype html>
<html lang="ru">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Els · ${titleCase} (dev)</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/dev.tsx"></script>
    <div id="portal"></div>
  </body>
</html>
`
}

function srcIndex({ routesExport }) {
  return `export { ${routesExport} } from './routes.tsx'\n`
}

function srcRoutes({ routesExport, pageName }) {
  return `import { Route, Routes } from 'react-router-dom'
import { ${pageName} } from './${pageName}.tsx'

export function ${routesExport}() {
  return (
    <Routes>
      <Route index element={<${pageName} />} />
      <Route path="*" element={<${pageName} />} />
    </Routes>
  )
}
`
}

function srcPage({ pageName, titleCase, slug }) {
  return `import { Card, CardContent, CardHeader, CardTitle } from '@els/ui'

export function ${pageName}() {
  return (
    <div className="h-full w-full overflow-auto bg-neutral-50 p-6">
      <div className="mx-auto max-w-3xl">
        <Card>
          <CardHeader>
            <CardTitle>${titleCase}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-neutral-600">
              Hello from <code className="bg-neutral-100 px-1 rounded">@els/${slug}-app</code>.
              Edit <code className="bg-neutral-100 px-1 rounded">src/${pageName}.tsx</code> to start.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
`
}

function srcDev({ routesExport, slug }) {
  return `import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { DevShell } from '@els/dev-harness'
import { ${routesExport} } from './routes.tsx'
import './index.css'

const rootEl = document.getElementById('root')
if (!rootEl) throw new Error('#root not found')

createRoot(rootEl).render(
  <StrictMode>
    <DevShell title="${slug}-app" initialPath="/">
      <${routesExport} />
    </DevShell>
  </StrictMode>,
)
`
}

function srcCss() {
  return `@import "tailwindcss";

@source "../../../packages/ui/src/**/*.{ts,tsx}";
@source "../../../packages/dev-harness/src/**/*.{ts,tsx}";

@theme {
  --color-brand-50:  #ecfdf5;
  --color-brand-100: #d1fae5;
  --color-brand-200: #a7f3d0;
  --color-brand-300: #6ee7b7;
  --color-brand-400: #34d399;
  --color-brand-500: #10b981;
  --color-brand-600: #059669;
  --color-brand-700: #047857;
  --color-brand-800: #065f46;
  --color-brand-900: #064e3b;
}
`
}

function srcApi() {
  return `import { createApi } from '@els/api-client'

const TOKEN_KEY = 'els.auth.token'

/**
 * Module-level api-client singleton used by this feature.
 *
 * Same instance is used in isolated dev (\`pnpm dev:<feature>\`,
 * proxied through Vite to the backend) and in production
 * (main-app loads the feature via lazy-import; the client makes
 * relative \`/api/...\` requests against whatever host the SPA was
 * served from).
 *
 * Token is read from localStorage; the dev-harness banner has a
 * paste-token panel that writes there for isolated dev.
 */
export const api = createApi({
  baseUrl: '',
  getToken: () => {
    try {
      return localStorage.getItem(TOKEN_KEY)
    } catch {
      return null
    }
  },
})
`
}
