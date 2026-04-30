# AGENTS.md

## Project Structure

Monorepo with two projects:
- **CaddyAPI** (Go/Gin) - port 8081, or set `PORT` env var
- **CaddyWeb** (Vue 3 + Vite + TypeScript) - port 8080

```
CaddyWeb/src/
├── api/                 # API layer (axios instance, API methods)
│   └── index.ts
├── components/          # Reusable Vue components
├── layout/              # Layout components
│   └── Layout.vue
├── router/              # Vue Router configuration
│   └── index.ts
├── stores/              # Pinia state management
│   └── settings.ts
├── styles/              # Global styles
├── types/               # TypeScript types and enums (CENTRALIZED)
│   └── index.ts         # All shared interfaces and type aliases
├── views/               # Page components
│   ├── Dashboard.vue
│   ├── Domains.vue
│   ├── Logs.vue
│   ├── Settings.vue
│   ├── Sites.vue
│   ├── Setup.vue
│   ├── TLS.vue
│   └── Login.vue
├── App.vue
└── main.ts
```

## Commands

### Backend
```bash
cd CaddyAPI && go run main.go    # Run server
cd CaddyAPI && go build          # Build binary
```

### Frontend
```bash
cd CaddyWeb && pnpm dev         # Dev server (--host for network access)
cd CaddyWeb && pnpm build       # Production build
cd CaddyWeb && pnpm preview     # Preview build
```

### Type-checking
```bash
npx vue-tsc --noEmit           # TypeScript check (frontend)
go build ./...                 # Go build (backend, no separate typecheck)
```

## Development Workflow

1. Start backend first: `cd CaddyAPI && go run main.go`
2. Then frontend: `cd CaddyWeb && pnpm dev`
3. Frontend proxies `/api` to `http://localhost:8081`

## Architecture & Conventions

### TypeScript Types (CENTRALIZED)

All shared TypeScript interfaces and enum types MUST be placed in `src/types/index.ts`.

**Type exports pattern:**
```typescript
// types/index.ts
export type SiteType = 'static' | 'reverse_proxy'
export type CaddyStatus = 'running' | 'stopped'
export interface Certificate { ... }
```

**Usage in components:**
```typescript
import type { Certificate, CaddyStatus } from '@/types'
import { THEMES } from '@/types'  // for const arrays
```

### API Layer

All API calls go through `src/api/index.ts`. API methods return typed promises.

**API structure:**
```typescript
// api/index.ts
export const caddyAPI = {
    getConfig: () => get<CaddyConfig>('/config'),
    // ...
}
export const domainAPI = { ... }
export const siteAPI = { ... }
export const settingsAPI = { ... }
export const logsAPI = { ... }
export const sseAPI = { ... }
```

### Vue Components

- **Views** (`src/views/`): Page-level components, import types from `@/types`
- **Components** (`src/components/`): Reusable UI components
- **Layout** (`src/layout/`): Layout wrapper components

### Form Components

Use Element Plus with `import type { FormInstance } from 'element-plus'` for type annotations.

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Interfaces | PascalCase, no prefix | `SiteFormData` |
| Type Aliases | PascalCase | `SiteType`, `CaddyStatus` |
| Const Arrays | UPPER_SNAKE_CASE | `THEMES`, `SITE_TYPES` |
| API Objects | camelCase + API suffix | `caddyAPI`, `domainAPI` |
| Vue Files | PascalCase | `Dashboard.vue`, `Domains.vue` |

### Vue 3 Composition API

- Use `<script setup lang="ts">` for all components
- Use `ref()` and `reactive()` for reactive state
- Use `computed()` for derived state
- Use `watch()` for side effects
- Use `shallowRef()` for large objects that don't need deep reactivity
- Import Element Plus types with `import type { FormInstance }`

## Notes

- No test suites configured
- No pre-commit hooks
- Frontend uses Element Plus UI + Pinia state
- Backend uses gorilla/websocket for real-time
- All enum/string union types must be centralized in `src/types/index.ts`
