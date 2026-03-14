# Initial API Outline

## Health

- `GET /health`

## Auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`

Implemented now:
- register
- login
- me

Planned next:
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`

## Courses

- `GET /api/v1/courses`
- `GET /api/v1/courses/:slug`

## Flashcard Sets

- `GET /api/v1/flashcard-sets`
- `POST /api/v1/flashcard-sets`
- `GET /api/v1/flashcard-sets/:slug`
- `PATCH /api/v1/flashcard-sets/:slug`
- `DELETE /api/v1/flashcard-sets/:slug`

Implemented now:
- list public flashcard sets
- get flashcard set by slug
- create flashcard set for authenticated user
- update owned flashcard set
- delete owned flashcard set

## Flashcards

- `GET /api/v1/flashcard-sets/:slug/cards`
- `POST /api/v1/flashcard-sets/:slug/cards`
- `PATCH /api/v1/flashcards/:id`
- `DELETE /api/v1/flashcards/:id`

Implemented now:
- list public flashcards for a public published set
- create flashcard inside an owned flashcard set
- update owned flashcard
- delete owned flashcard

## Progress

- `GET /api/v1/me/progress`
- `POST /api/v1/me/flashcards/:id/review`
