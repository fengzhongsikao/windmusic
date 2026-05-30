# AGENTS

## Project Overview

- Project type: desktop app with `Wails + Svelte + TypeScript`
- Goal: search songs from imported JS sources, fetch playable URL/lyrics/cover via Go backend, and play audio in frontend
- Runtime boundary:
  - Go (`Wails`) handles source management and data fetching
  - Svelte frontend handles UI state and actual audio playback (`HTMLAudioElement`)

## Repository Layout

- `main.go` / `app.go`: Wails app entry and exported methods
- `app_*.go`: Wails bindings split by domain (local library, user library, settings, sync)
- `internal/`: Go domain logic (models, musicsearch)
- `music/`: backend services
  - `appdata/`: config directory helpers
  - `events/`: Wails event name constants
  - `persist/`: JSON persistence (favorites, recent, playlists, settings)
  - `cache/`: in-memory caches (discover)
  - `local/`: local folder library
  - `meting/`: online source search and URLs
- `frontend/`: Svelte 5 app
  - `src/pages/`: page-level UI
  - `src/components/`: reusable UI (grouped: `layout/`, `player/`, `track/`, `playlist/`, `ui/`, `song/`)
  - `src/stores/`: global state (grouped by domain)
    - `playback/`: player queue, audio engine, lyrics, player settings
    - `library/`: local library, favorites, recent, playlists
    - `sources/`: Meting source settings
    - `ui/`: toast notifications
    - `sync/`: backend event subscription bootstrap
  - `wailsjs/`: generated Wails bindings (do not hand-edit)

## Frontend Tech Stack (from `frontend/package.json`)

- Core:
  - `svelte` `^5.55.9`
  - `typescript` `^6.0.3`
  - `vite` `^8.0.14`
- Svelte/Vite integration:
  - `@sveltejs/vite-plugin-svelte` `^7.1.2`
  - `@tsconfig/svelte` `^5.0.8`
  - `svelte-check` `^4.4.8`
  - `tslib` `^2.8.1`
  - `@types/node` `^25.9.1`
- UI:
  - `flowbite` `^4.0.2`
  - `flowbite-svelte` `^1.33.1`
  - `@lucide/svelte` `^1.3.0`
  - `tailwindcss` `^4.3.0`
  - `@tailwindcss/vite` `^4.3.0`
- Routing:
  - `svelte-spa-router` `^5.1.0`

## Development Commands

- Frontend install deps: `bun install --cwd frontend`
- Frontend dev server: `bun run --cwd frontend dev`
- Frontend type check: `bun run --cwd frontend check`
- Frontend build: `bun run --cwd frontend build`
- Wails app dev mode (recommended for full app): `wails dev`
- Wails production build: `wails build`

## Architecture Rules

- Playback URL flow:
  1. UI selects a track
  2. Frontend builds playback context (`sourceId`, `platform`, `metaJson`)
  3. Frontend calls `GetMusicURL(...)` through Wails binding
  4. Go source manager/runtime resolves final URL
  5. Frontend `<audio>` sets `src` and plays
- Keep this invariant: Go does not decode/play audio; frontend audio engine is the single playback executor.
- Keep queue/play state in stores; avoid page-local duplicated playback state.

## Coding Guidelines for Agents

- Prefer minimal, focused changes that match existing style.
- Do not edit generated files in `frontend/wailsjs/**` manually.
- When adding playback features:
  - update store API first
  - wire UI controls second
  - verify `ended`, `next/prev`, repeat/shuffle behavior
- Preserve current backend contract (`Search`, `GetMusicURL`, `GetLyric`, `GetPic`) unless explicitly changing both ends.
- Avoid introducing new dependencies unless required by feature scope.

## Validation Checklist

- Run `bun run --cwd frontend check` after meaningful UI/store changes.
- Manually verify:
  - discover/search can start playback
  - URL fetch failure shows clear error
  - play/pause/seek works
  - next/prev and repeat/shuffle behave correctly
  - lyrics still sync with current playback time
