---
paths: "web/**"
---

# Frontend Standards

## Component & Code Structure Rules

These rules apply to every file in `web/src`. They are non-negotiable and override default
behaviour. Any AI assistant working on this repo **must** follow them without exception.

1. **One component per file. Full stop.** No helpers, sub-components, icons, or rows
   alongside the primary component. If you need a second component, create a second file.
2. **Use props, `atom`, `atomFamily`, `atomWithStorage` where it makes sense.** Share state
   via atoms; avoid prop-drilling beyond one level.
3. **JavaScript is functional, not OO.** No classes. Pure functions, closures, hooks.
4. **No IIFEs and no `truthy && <Component>` inline conditional rendering.** Extract to
   render helper functions, named functions, or separate components instead.
5. **React 19 — drop unnecessary manual memoization.** Remove `useMemo`/`useCallback`/
   `React.memo` for pure derivations and stable child-passed callbacks — the React Compiler
   handles these. **Keep** manual memoization only for:
   - Non-pure or unstable computations touching globals/random/timers/I/O.
   - Third-party identity constraints (e.g. leaflet icons, canvas layer event handlers).
   - Expensive derived data with intentionally unstable inputs.
   - Intentional throttle/debounce wrappers (business logic, not perf micro-opts).
   - Opt-in micro-optimisations on verified hot loops in massive lists.
   Leave imperative `useRef` (DOM handles, websockets, autoscroll) untouched.
6. **Design for screen size.** Features may collapse or disappear on mobile/tablet. *Full
   responsive refactor is a planned future effort — see the backlog.*
7. **`index.ts` files re-export only. Never define a component in an index file.** A folder
   that contains a tab or view component should have a named component file
   (`FooTab/FooTab.tsx`) and a barrel (`FooTab/index.ts` → `export { FooTab } from './FooTab'`).
8. **Types → `types.ts` next to the component files that use them.** Do not declare types
   inline in `.tsx` files. Export types up the tree as needed.
9. **Interfaces → `interfaces.ts` next to the component files that use them.** Do not
   declare interfaces inline in `.tsx` files. Export up the tree as needed.
10. **Export types and interfaces as needed.** Shared types/interfaces are re-exported from
    the closest common ancestor barrel.
11. **No React shorthand imports.** Always write `import * as React from 'react';` and use
    the namespace: `React.FC`, `React.PropsWithChildren`, `React.Fragment`, `React.useState`,
    `React.useEffect`, etc. Never use `import React from 'react'`, `import { useState }`,
    or `<>...</>` fragment shorthand — write `React.Fragment` explicitly.
12. **All functions must declare concrete return types.** Components via `React.FC<Props>`.
    Generic components via `: React.ReactNode` or `: JSX.Element`. Helper functions via their
    explicit return type. Non-trivial extracted render helpers must be typed. Trivial inline
    event handlers (`onClick={() => set(v)}`) are exempt.
13. **`exactOptionalPropertyTypes` is enforced.** Respect it — check every optional prop and
    index-access for `undefined`. `strict: true` and `exactOptionalPropertyTypes: true` are
    set in `web/tsconfig.json`. Do not pass `undefined` to an optional prop unless the prop
    type explicitly includes `| undefined`.

---

## Stack

- **Framework**: React 19 + TypeScript (`strict` + `exactOptionalPropertyTypes`)
- **UI library**: HeroUI v3 (via `dune-ui/` wrappers)
- **Build**: Vite + React Compiler (`babel-plugin-react-compiler`)
- **Package manager**: `pnpm` — **never use npm or yarn in `web/`**
- **Auth**: Clerk (optional; keyed off `VITE_CLERK_PUBLISHABLE_KEY`)

## Canonical Reference Pattern

**`BasesTab.tsx` is the reference for new simple tabs.** Read it before creating a new tab.

### Minimal Tab Structure

`tabs/FooTab/FooTab.tsx` — the component file:

```tsx
import * as React from 'react';
import { api, ApiError } from '../../api/client';
import { Panel, PageHeader, DataTable } from '../../dune-ui';
import type { Column } from '../../dune-ui';
import type { FooRow } from './types';

const columns: Column<FooRow>[] = [/* … */];

export const FooTab: React.FC = (): React.ReactElement => {
  const [data, setData] = React.useState<FooRow[]>([]);
  const [loading, setLoading] = React.useState(false);

  const load = async (): Promise<void> => {
    setLoading(true);
    try {
      setData(await api.foo.list());
    } catch (e) {
      toast.danger(`Failed: ${e instanceof Error ? e.message : String(e)}`);
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => { void load(); }, []);

  return (
    <Panel>
      <PageHeader title="Foo" onRefresh={load} loading={loading} />
      <DataTable columns={columns} rows={data} />
    </Panel>
  );
};
```

`tabs/FooTab/index.ts` — the barrel (re-export only, never a component):

```ts
export { FooTab } from './FooTab';
```

### Complex Tab Structure

For complex tabs, use a directory. The root component lives in a **named file**, not
`index.tsx`. `index.ts` re-exports only:

```
tabs/FooTab/
  index.ts        — barrel: export { FooTab } from './FooTab'
  FooTab.tsx      — root component (one component only)
  types.ts        — local types (no interfaces here → interfaces.ts)
  interfaces.ts   — local interfaces
  components/     — tab-local components (one component per file)
  modals/         — modal components (one component per file)
  views/          — sub-views (if needed)
```

## API Client

All backend calls go through `api/client.ts`. Import the `api` namespace for typed wrappers:

```ts
import { api, ApiError } from '../api/client'

const rows = await api.foo.list()
const detail = await api.foo.get(id)
```

- Do not use `fetch` directly in tab/component code
- The backend URL is runtime-configurable via `localStorage('dune_admin_backend')`
- Vite dev server proxies `/api` and WebSocket `/api/v1/logs/stream` → `:8080`

## Component Library (`dune-ui/`)

Import shared components from `../dune-ui` when a wrapper exists:

```ts
import {
  ActionBar, ConfirmDialog, DataTable, Dropzone, FieldInput, FieldSelect,
  Icon, InfoCard, LoadingState, NumberInput, PageHeader, Panel,
  SectionDivider, SectionLabel, SideNav, TimeInput,
} from '../dune-ui'
import type { Column } from '../dune-ui'
```

Use `@heroui/react` directly only for primitives not wrapped in `dune-ui`
(Button, Card, Chip, Spinner, toast, etc.).

`StatusChip` was removed — use inline `<Chip size="sm" variant="soft" color={...}>` instead.

### `FieldInput`

Wraps HeroUI `Input` with `size="sm"`. Use for all text, number, password, email, and url inputs.

```tsx
<FieldInput value={val} onChange={setVal} placeholder="…" aria-label="…" />
<FieldInput type="number" value={num} onChange={setNum} className="w-32" />
<FieldInput value={path} onChange={setPath} classNames={{ input: 'font-mono' }} />
```

### `FieldSelect`

Wraps HeroUI `Select` + `ListBox` for small, fixed option sets (booleans, enums up to ~20 items).

```tsx
<FieldSelect value={val} onChange={setVal} options={['true', 'false']} />
<FieldSelect value={mode} onChange={setMode} options={['A', 'B', 'C']} className="w-40" />
```

For large option sets, `FieldSelect` (and HeroUI `Select` directly) still work — use them
for visual consistency. `TimezoneSelect` in `components/` wraps HeroUI `Select` for the
~400-entry IANA list with a host-local sentinel.

### `TimeInput`

Wraps HeroUI `TimeField` with 24-hour segmented input. Accepts and emits `"HH:MM"` strings.

```tsx
<TimeInput value={rule.time} onChange={(v) => setRuleTime(i, v)} ariaLabel="time" />
```

### Checkboxes and Toggles

Use HeroUI's `Checkbox` and `Switch` from `@heroui/react` — never native `<input type="checkbox">`:

```tsx
import { Checkbox, Switch } from '@heroui/react'

// Toggle (on/off) — use Switch. The compound children are REQUIRED:
// without Switch.Control/Switch.Thumb no visual control renders at all.
<Switch isSelected={enabled} onChange={setEnabled} size="sm">
  <Switch.Control><Switch.Thumb /></Switch.Control>
  <Switch.Content>{t('enable')}</Switch.Content>
</Switch>

// Checkbox (filter/option) — use Checkbox (no size prop)
<Checkbox isSelected={isOn} onChange={setOn}>{t('label')}</Checkbox>

// Checkbox with indeterminate state
<Checkbox isSelected={allOn} isIndeterminate={!allOn && anyOn} onChange={handleChange} />
```

## HeroUI v3 limitations

- HeroUI `Select` has no `<optgroup>` — keep native `<select>` for grouped option lists
- No equivalent for `<input list="...">` + `<datalist>` — keep native `<input>` with `bg-surface text-foreground border border-border rounded`

## Migration backlog

BattlegroupTab, StorageTab, DatabaseTab, LogsTab, BlueprintsTab still use raw HTML + inline styles. When refactoring any of these, follow the BasesTab pattern. Do not remove state/code — use `display: none` to hide features temporarily.

## Theming

All colours are CSS custom properties defined in `web/src/index.css`.
**Never use raw Tailwind colour utilities** (`bg-amber-900`, `text-zinc-400`, etc.).

Use semantic utilities:

```
bg-background       bg-surface        bg-surface-secondary
text-foreground     text-muted        text-accent
border-border
```

Inline `style={{ color: '#...' }}` overrides for colours are a sign the semantic token
approach wasn't used — fix them.

## Auth

`hasClerk = !!import.meta.env.VITE_CLERK_PUBLISHABLE_KEY`

When absent, the app renders without auth (local dev). The `isSignedIn` prop gates
destructive features in certain tabs. Do not remove this gate.

## Frontend Checklist

- [ ] Using `pnpm` (not npm/yarn)
- [ ] New tab follows `BasesTab.tsx` pattern
- [ ] All API calls go through `api/client.ts`
- [ ] `dune-ui/` wrappers used instead of direct `@heroui/react` where available
- [ ] Semantic colour tokens used (no raw Tailwind colours, no inline colour styles)
- [ ] One component per file (Rules 1, 7)
- [ ] `import * as React from 'react'` — no shorthand imports or `<>` fragments (Rule 11)
- [ ] All exported functions and components have concrete return types (Rule 12)
- [ ] No inline `interface`/`type` in `.tsx` — colocated in `types.ts`/`interfaces.ts` (Rules 8-10)
- [ ] No unnecessary `useMemo`/`useCallback` — React Compiler handles these (Rule 5)
- [ ] `exactOptionalPropertyTypes` respected — no silent `undefined` in optional props (Rule 13)
- [ ] `pnpm lint` passes (`cd web && pnpm lint`)
- [ ] `pnpm build` passes (tsc -b + vite build)
