# ELS — English Learning Studio

Practice a language with **films**, **interactive quests**, **vocabulary books**, and an **AI assistant** that can read your current lesson or scene.

Built around English by default, but the same studio can be adapted to any language.

## Screenshots

### Interactive quest
![Quest](docs/screenshots/01-quest.png)

### Film + on-screen subtitles
![Film](docs/screenshots/02-film.png)

### Grammar unit (theory, pictures, matching & highlighting)
![Grammar](docs/screenshots/03-grammar.png)

### Assistant grounded in the open film scene
![Assistant](docs/screenshots/04-assistant.png)

## Quick start

```bash
# backend
cd backend && cp .env.example .env && make up

# frontend (other terminal)
cd frontend && pnpm install && pnpm --filter @els/main-app dev
```

Open http://localhost:5173 — default admin is in `backend/.env.example` (`BOOTSTRAP_ADMIN_*`).
