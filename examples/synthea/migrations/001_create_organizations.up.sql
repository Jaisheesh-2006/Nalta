-- Migration: 001_create_organizations.up.sql
-- Healthcare organizations and facilities (hospitals, clinics, etc.)
-- Must be created FIRST — referenced by providers and encounters.

CREATE TABLE IF NOT EXISTS organizations (
    Id          VARCHAR(36)     NOT NULL,
    NAME        VARCHAR(255)    NOT NULL,
    ADDRESS     VARCHAR(255)    NOT NULL,
    CITY        VARCHAR(100)    NOT NULL,
    STATE       VARCHAR(50)     NOT NULL,
    ZIP         VARCHAR(20)     NOT NULL,
    LAT         DECIMAL(18,15)  NOT NULL,
    LON         DECIMAL(18,15)  NOT NULL,
    PHONE       VARCHAR(30),
    REVENUE     DECIMAL(15,4),
    UTILIZATION INT             NOT NULL DEFAULT 0,

    PRIMARY KEY (Id),
    INDEX idx_state      (STATE),
    INDEX idx_city_state (CITY, STATE),
    INDEX idx_zip        (ZIP)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Healthcare organizations and facilities (hospitals, clinics, urgent care, etc.)';
