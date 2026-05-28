# Clean Dark Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the cyberpunk/geek theme with a clean, elegant dark theme using a pure grayscale palette.

**Architecture:** Rewrite CSS variables and component overrides. Remove decorative elements (particle background, glow effects, monospace fonts). Update scoped styles in Vue components.

**Tech Stack:** Vue 3, Element Plus, CSS variables

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `web/frontend/src/styles/geek-theme.css` | Rewrite | Main theme variables and Element Plus overrides |
| `web/frontend/src/views/LayoutView.vue` | Modify | Remove particle background, simplify layout background |
| `web/frontend/src/components/TopNavBar.vue` | Modify | Remove monospace font, glow effects, geek colors |
| `web/frontend/src/views/NodesView.vue` | Modify | Remove `geek-title` class usage if hardcoded styles exist |
| `web/frontend/src/views/SubscriptionView.vue` | Modify | Remove `geek-title` class usage |
| `web/frontend/src/views/RoutingView.vue` | Modify | Remove `geek-title` class usage |
| `web/frontend/src/views/SettingsView.vue` | Modify | Remove `geek-title` class usage, geek border styles |
| `web/frontend/src/views/LoginView.vue` | Modify | Remove geek accent colors |
| `web/frontend/src/components/AddNodeDialog.vue` | Modify | Remove `geek-dialog` class |
| `web/frontend/src/components/AddSubscriptionDialog.vue` | Modify | Remove `geek-dialog` class |
| `web/frontend/src/components/NodeTable.vue` | Modify | Remove geek border/accent colors |
| `web/frontend/src/components/RouteRuleEditor.vue` | Modify | Remove geek border color |
| `web/frontend/src/components/ProxyControl.vue` | Modify | Remove geek border color |

---

## Task 1: Rewrite theme CSS

**Files:**
- Modify: `web/frontend/src/styles/geek-theme.css`

- [ ] **Step 1: Rewrite geek-theme.css with clean dark palette**

Replace the entire file content with new grayscale variables and simplified Element Plus overrides. Remove all green neon colors, glow effects, monospace fonts, and uppercase transforms.

Key changes:
- `--geek-bg: #0a0a0a`
- `--geek-bg-secondary: #111111`
- `--geek-bg-card: #111111`
- `--geek-border: #222222`
- `--geek-border-hover: #333333`
- `--geek-text: #e8e8e8`
- `--geek-text-secondary: #888888`
- `--geek-accent: #ffffff`
- `--geek-accent-secondary: #888888`
- `--geek-danger: #ff4d4f`
- `--geek-warning: #faad14`
- Font: system sans-serif stack
- Remove all `text-transform: uppercase`, excessive `letter-spacing`, glow `box-shadow`
- Simplify `.geek-title` to plain heading style

- [ ] **Step 2: Commit**

```bash
git add web/frontend/src/styles/geek-theme.css
git commit -m "feat(ui): rewrite theme to clean dark grayscale"
```

---

## Task 2: Remove particle background from layout

**Files:**
- Modify: `web/frontend/src/views/LayoutView.vue`

- [ ] **Step 1: Remove ParticleBackground and simplify layout**

Remove `<ParticleBackground />` from template. Change background from gradient to solid `#0a0a0a`.

- [ ] **Step 2: Commit**

```bash
git add web/frontend/src/views/LayoutView.vue
git commit -m "feat(ui): remove particle background, simplify layout"
```

---

## Task 3: Clean up TopNavBar

**Files:**
- Modify: `web/frontend/src/components/TopNavBar.vue`

- [ ] **Step 1: Remove geek styles from TopNavBar**

Remove:
- `font-family: 'JetBrains Mono'` from `.logo`, `.status-text`, `.lang-toggle`
- `box-shadow: 0 0 8px ...` glow effects from `.status-dot`
- `letter-spacing` properties
- `text-transform: uppercase`
- `var(--geek-accent)` color from `.logo` (change to white or default text)

- [ ] **Step 2: Commit**

```bash
git add web/frontend/src/components/TopNavBar.vue
git commit -m "feat(ui): clean up TopNavBar styles"
```

---

## Task 4: Clean up view components

**Files:**
- Modify: `web/frontend/src/views/NodesView.vue`
- Modify: `web/frontend/src/views/SubscriptionView.vue`
- Modify: `web/frontend/src/views/RoutingView.vue`
- Modify: `web/frontend/src/views/SettingsView.vue`
- Modify: `web/frontend/src/views/LoginView.vue`

- [ ] **Step 1: Remove geek-specific styles from views**

For each view:
- Keep `geek-title` class on `<h2>` elements (CSS handles the style)
- Remove hardcoded `var(--geek-*)` colors in scoped `<style>` blocks
- Remove `border: 1px solid var(--geek-border)` overrides where Element Plus default is fine

- [ ] **Step 2: Commit**

```bash
git add web/frontend/src/views/
git commit -m "feat(ui): clean up view component styles"
```

---

## Task 5: Clean up remaining components

**Files:**
- Modify: `web/frontend/src/components/AddNodeDialog.vue`
- Modify: `web/frontend/src/components/AddSubscriptionDialog.vue`
- Modify: `web/frontend/src/components/NodeTable.vue`
- Modify: `web/frontend/src/components/RouteRuleEditor.vue`
- Modify: `web/frontend/src/components/ProxyControl.vue`

- [ ] **Step 1: Remove geek classes and colors**

- Remove `class="geek-dialog"` from dialogs
- Remove hardcoded `var(--geek-border)` and `var(--geek-accent-secondary)` where Element Plus default is fine

- [ ] **Step 2: Commit**

```bash
git add web/frontend/src/components/
git commit -m "feat(ui): clean up component styles"
```

---

## Task 6: Build and verify

- [ ] **Step 1: Build frontend**

```bash
cd web/frontend
npm run build
```

Expected: Build succeeds without errors.

- [ ] **Step 2: Final commit**

```bash
git add -A
git commit -m "feat(ui): complete clean dark theme"
```

---

## Spec Coverage Check

| Spec Requirement | Task |
|---|---|
| Grayscale color system | Task 1 |
| Sans-serif font | Task 1 |
| Remove particle background | Task 2 |
| Remove glow effects | Tasks 1, 3, 4, 5 |
| Remove uppercase transforms | Task 1 |
| Remove monospace fonts | Tasks 1, 3 |

---

## Placeholder Scan

No placeholders. All steps contain exact file paths and specific changes.

---

## Type Consistency Check

- CSS variable names (`--geek-*`) remain consistent across all tasks — only values change
- Component class names (`geek-title`, `geek-dialog`) remain as handles for CSS rules
- No new types or signatures introduced
