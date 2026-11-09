CREATE TABLE IF NOT EXISTS users (
  id            BINARY(16)   NOT NULL PRIMARY KEY,
  email         VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  display_name  VARCHAR(120) NOT NULL,
  city          VARCHAR(120) NOT NULL DEFAULT '',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sessions (
  token      CHAR(64)   NOT NULL PRIMARY KEY,
  user_id    BINARY(16) NOT NULL,
  created_at DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME   NOT NULL,
  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_sessions_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pets (
  id         BINARY(16)   NOT NULL PRIMARY KEY,
  user_id    BINARY(16)   NOT NULL,
  name       VARCHAR(120) NOT NULL,
  species    VARCHAR(40)  NOT NULL DEFAULT 'dog',
  sex        ENUM('male','female','unknown') NOT NULL DEFAULT 'unknown',
  birth_year SMALLINT     NULL,
  breed      VARCHAR(120) NOT NULL DEFAULT '',
  traits     VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_pets_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_pets_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS walks (
  id         BINARY(16)   NOT NULL PRIMARY KEY,
  user_id    BINARY(16)   NOT NULL,
  pet_id     BINARY(16)   NULL,
  started_at DATETIME     NOT NULL,
  ended_at   DATETIME     NOT NULL,
  distance_m INT UNSIGNED NOT NULL DEFAULT 0,
  h3_cells   INT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_walks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_walks_pet FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE SET NULL,
  INDEX idx_walks_user_started (user_id, started_at),
  INDEX idx_walks_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS walk_points (
  walk_id BINARY(16)        NOT NULL,
  seq     SMALLINT UNSIGNED NOT NULL,
  lat     DECIMAL(9,6)      NOT NULL,
  lng     DECIMAL(9,6)      NOT NULL,
  at_utc  DATETIME          NOT NULL,
  PRIMARY KEY (walk_id, seq),
  CONSTRAINT fk_points_walk FOREIGN KEY (walk_id) REFERENCES walks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS meetups (
  id         BINARY(16)   NOT NULL PRIMARY KEY,
  user_id    BINARY(16)   NOT NULL,
  title      VARCHAR(160) NOT NULL,
  lat        DECIMAL(9,6) NOT NULL,
  lng        DECIMAL(9,6) NOT NULL,
  starts_at  DATETIME     NOT NULL,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_meetups_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_meetups_starts (starts_at),
  INDEX idx_meetups_geo (lat, lng)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
