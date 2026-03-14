CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
    CREATE TYPE user_role AS ENUM ('student', 'admin');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
    CREATE TYPE user_status AS ENUM ('active', 'pending', 'disabled');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'content_visibility') THEN
    CREATE TYPE content_visibility AS ENUM ('private', 'unlisted', 'public');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'content_status') THEN
    CREATE TYPE content_status AS ENUM ('draft', 'published', 'archived');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'study_result') THEN
    CREATE TYPE study_result AS ENUM ('correct', 'incorrect', 'skipped');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(50) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  full_name VARCHAR(120),
  role user_role NOT NULL DEFAULT 'student',
  status user_status NOT NULL DEFAULT 'active',
  avatar_url TEXT,
  last_login_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token_hash TEXT NOT NULL,
  user_agent TEXT,
  ip_address INET,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);

CREATE TABLE IF NOT EXISTS courses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug VARCHAR(120) NOT NULL UNIQUE,
  title VARCHAR(160) NOT NULL,
  category VARCHAR(80) NOT NULL,
  description TEXT,
  summary TEXT,
  thumbnail_url TEXT,
  is_published BOOLEAN NOT NULL DEFAULT TRUE,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_courses_category ON courses(category);
CREATE INDEX IF NOT EXISTS idx_courses_is_published ON courses(is_published);

CREATE TABLE IF NOT EXISTS flashcard_sets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
  slug VARCHAR(160) NOT NULL UNIQUE,
  title VARCHAR(160) NOT NULL,
  description TEXT,
  visibility content_visibility NOT NULL DEFAULT 'private',
  status content_status NOT NULL DEFAULT 'draft',
  language_code VARCHAR(12) NOT NULL DEFAULT 'id',
  card_count INTEGER NOT NULL DEFAULT 0,
  estimated_minutes INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_flashcard_sets_owner_id ON flashcard_sets(owner_id);
CREATE INDEX IF NOT EXISTS idx_flashcard_sets_course_id ON flashcard_sets(course_id);
CREATE INDEX IF NOT EXISTS idx_flashcard_sets_visibility ON flashcard_sets(visibility);
CREATE INDEX IF NOT EXISTS idx_flashcard_sets_status ON flashcard_sets(status);

CREATE TABLE IF NOT EXISTS flashcards (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  flashcard_set_id UUID NOT NULL REFERENCES flashcard_sets(id) ON DELETE CASCADE,
  position INTEGER NOT NULL,
  question TEXT NOT NULL,
  answer TEXT NOT NULL,
  explanation TEXT,
  hint TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_flashcards_set_position UNIQUE (flashcard_set_id, position)
);

CREATE INDEX IF NOT EXISTS idx_flashcards_set_id ON flashcards(flashcard_set_id);

CREATE TABLE IF NOT EXISTS flashcard_set_tags (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  flashcard_set_id UUID NOT NULL REFERENCES flashcard_sets(id) ON DELETE CASCADE,
  tag VARCHAR(50) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_flashcard_set_tag UNIQUE (flashcard_set_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_flashcard_set_tags_tag ON flashcard_set_tags(tag);

CREATE TABLE IF NOT EXISTS user_flashcard_set_progress (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  flashcard_set_id UUID NOT NULL REFERENCES flashcard_sets(id) ON DELETE CASCADE,
  mastery_level SMALLINT NOT NULL DEFAULT 0 CHECK (mastery_level BETWEEN 0 AND 100),
  cards_completed INTEGER NOT NULL DEFAULT 0,
  times_studied INTEGER NOT NULL DEFAULT 0,
  last_studied_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_user_flashcard_set_progress UNIQUE (user_id, flashcard_set_id)
);

CREATE INDEX IF NOT EXISTS idx_user_set_progress_user_id ON user_flashcard_set_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_user_set_progress_set_id ON user_flashcard_set_progress(flashcard_set_id);

CREATE TABLE IF NOT EXISTS user_flashcard_progress (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  flashcard_id UUID NOT NULL REFERENCES flashcards(id) ON DELETE CASCADE,
  correct_count INTEGER NOT NULL DEFAULT 0,
  incorrect_count INTEGER NOT NULL DEFAULT 0,
  last_result study_result,
  confidence_score SMALLINT NOT NULL DEFAULT 0 CHECK (confidence_score BETWEEN 0 AND 100),
  last_reviewed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_user_flashcard_progress UNIQUE (user_id, flashcard_id)
);

CREATE INDEX IF NOT EXISTS idx_user_card_progress_user_id ON user_flashcard_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_user_card_progress_flashcard_id ON user_flashcard_progress(flashcard_id);
