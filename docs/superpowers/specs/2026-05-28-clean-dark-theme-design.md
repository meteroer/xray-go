# Clean Dark Theme Design

## Overview

Redesign the xray-go web UI from a cyberpunk/geek aesthetic to a clean, elegant dark theme using a pure grayscale palette (no neon colors).

## Current State

The UI currently uses `geek-theme.css` with:
- Neon green accent (#00ff88) with glow effects
- Particle background animation (`ParticleBackground.vue`)
- Monospace font (JetBrains Mono)
- Uppercase text transforms
- Gradient backgrounds
- Excessive letter spacing

## Design

### Color System

| Purpose | Value |
|---------|-------|
| Page background | `#0a0a0a` |
| Card / panel | `#111111` |
| Hover / active | `#1a1a1a` |
| Primary text | `#e8e8e8` |
| Secondary text | `#888888` |
| Muted / disabled | `#555555` |
| Border default | `#222222` |
| Border hover | `#333333` |
| Accent (buttons, active) | `#ffffff` |

### Typography

Replace monospace font with system sans-serif stack:
```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
```

Remove all `text-transform: uppercase` and excessive `letter-spacing`.

### Elements to Remove

- Particle background animation
- Neon green accent color and glow box-shadows
- Gradient backgrounds
- Uppercase text transforms
- Exaggerated letter spacing
- Monospace font usage

### Elements to Keep

- Dark background
- Card-based layout (with subtle borders instead of glow)
- Basic hover states (brightness shift instead of glow)
- Element Plus component dark adaptation

## Files to Modify

1. `web/frontend/src/styles/geek-theme.css` — Rewrite theme variables and overrides
2. `web/frontend/src/views/LayoutView.vue` — Remove particle background, simplify background
3. `web/frontend/src/components/TopNavBar.vue` — Remove monospace font and glow effects
4. Review other components for hardcoded cyberpunk styles

## Implementation Steps

1. Rewrite `geek-theme.css` with new color variables and simplified component overrides
2. Remove `ParticleBackground` from `LayoutView.vue`
3. Clean up `TopNavBar.vue` scoped styles
4. Audit other Vue components for remaining cyberpunk styles
5. Build and verify the UI renders correctly
