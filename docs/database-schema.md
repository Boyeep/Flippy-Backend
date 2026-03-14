# Flippy Database Schema

## Core Design Goals

- support user accounts and authentication
- allow flashcard sets owned by users
- let flashcard sets optionally belong to courses
- track publishing state and soft ownership metadata
- track user learning progress over time

## Tables

### `users`

Application accounts.

Key fields:
- `id`
- `username`
- `email`
- `password_hash`
- `role`
- `status`

### `user_sessions`

Refresh-token/session tracking for sign-in management.

Key fields:
- `id`
- `user_id`
- `refresh_token_hash`
- `expires_at`
- `revoked_at`

### `courses`

High-level learning collections such as Math or Science.

Key fields:
- `id`
- `slug`
- `title`
- `category`
- `description`
- `is_published`

### `flashcard_sets`

User-created study sets.

Key fields:
- `id`
- `owner_id`
- `course_id`
- `slug`
- `title`
- `description`
- `visibility`
- `status`
- `card_count`

Notes:
- a set belongs to one owner
- a set can optionally be linked to one course

### `flashcards`

Cards that belong to a set.

Key fields:
- `id`
- `flashcard_set_id`
- `position`
- `question`
- `answer`
- `explanation`

### `flashcard_set_tags`

Simple tag system for discovery and filtering.

### `user_flashcard_set_progress`

Aggregated progress per user per flashcard set.

Key fields:
- `mastery_level`
- `last_studied_at`
- `times_studied`
- `cards_completed`

### `user_flashcard_progress`

Per-card user performance.

Key fields:
- `correct_count`
- `incorrect_count`
- `last_result`
- `confidence_score`

## Relationship Summary

- one `user` has many `flashcard_sets`
- one `course` has many `flashcard_sets`
- one `flashcard_set` has many `flashcards`
- one `user` has many `user_sessions`
- one `user` has many set progress rows
- one `user` has many card progress rows

## First API Modules Suggested

1. `auth`
2. `courses`
3. `flashcard_sets`
4. `flashcards`
5. `me/progress`
