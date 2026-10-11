# Hybrid-Grid Website

This directory contains the Hybrid-Grid landing page and documentation site. It is built with Astro 7, Starlight, and Tailwind CSS, independently from the Go binaries and the React operations dashboard.

## Requirements

- Node.js 22.12 or newer
- npm

## Development

Install the locked dependencies:

```bash
cd site
npm ci
```

Start the local development server:

```bash
npm run dev
```

The site is available at `http://localhost:4321` by default. When using the project’s Claude development workflow, start Astro in background mode as described in `CLAUDE.md`.

## Build and preview

Create the production output in `site/dist`:

```bash
npm run build
```

Preview that output locally:

```bash
npm run preview
```

## Project structure

- `src/pages/index.astro` — landing page entry point.
- `src/components/` — landing-page sections and shared Astro components.
- `src/content/docs/` — Starlight documentation pages.
- `src/styles/` — landing-page and Starlight theme styles.
- `src/assets/` and `public/` — bundled and static assets.
- `astro.config.mjs` — Starlight navigation, branding, social links, and Vite integrations.
- `src/content.config.ts` — Starlight content collection configuration.

When adding or moving documentation pages, update the sidebar in `astro.config.mjs` so navigation and content slugs remain aligned. Run `npm run build` before opening a pull request to catch broken links, invalid frontmatter, and type errors.
