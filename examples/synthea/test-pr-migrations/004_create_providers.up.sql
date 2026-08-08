-- Migration: 004_create_providers.up.sql
-- Individual clinicians and care providers attached to organizations.
-- Depends on: organizations (001)

CREATE TABLE IF NOT EXISTS providers (
    Id           VARCHAR(36)     NOT NULL,
    ORGANIZATION VARCHAR(36)     NOT NULL COMMENT 'FK → organizations.Id',
    NAME         VARCHAR(255)    NOT NULL COMMENT 'PII: Full name of the provider',
    GENDER       VARCHAR(1)      COMMENT 'M or F',
    SPECIALITY   VARCHAR(255)    NOT NULL,
    ADDRESS      VARCHAR(255),
    CITY         VARCHAR(100),
    STATE        VARCHAR(50),
    ZIP          VARCHAR(20),
    LAT          DECIMAL(18,15),
    LON          DECIMAL(18,15),
    UTILIZATION  INT             NOT NULL DEFAULT 0 COMMENT 'Total encounters conducted by this provider',

    PRIMARY KEY (Id),
    INDEX idx_organization       (ORGANIZATION),
    INDEX idx_speciality         (SPECIALITY),
    INDEX idx_state_speciality   (STATE, SPECIALITY),
    INDEX idx_name               (NAME)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Individual healthcare providers and clinicians attached to organizations.';
