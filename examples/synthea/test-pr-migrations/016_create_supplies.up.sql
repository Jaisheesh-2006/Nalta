-- Migration: 016_create_supplies.up.sql
-- Medical supplies consumed during patient encounters (e.g., bandages, gloves, syringes).
-- NOTE: This table is empty in the small sample dataset but populated in the 10k COVID dataset.
-- Uses SNOMED CT codes.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS supplies (
    DATE        DATE           NOT NULL  COMMENT 'Date the supply was used',
    PATIENT     VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER   VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id',
    CODE        VARCHAR(50)    NOT NULL  COMMENT 'SNOMED CT supply item code',
    DESCRIPTION VARCHAR(500)             COMMENT 'Human-readable name of the supply item',
    QUANTITY    INT            NOT NULL DEFAULT 1 COMMENT 'Number of units of this supply consumed',

    INDEX idx_patient          (PATIENT),
    INDEX idx_encounter        (ENCOUNTER),
    INDEX idx_code             (CODE),
    INDEX idx_date             (DATE),
    INDEX idx_patient_date     (PATIENT, DATE),
    INDEX idx_patient_code     (PATIENT, CODE)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Medical supplies consumed during encounters (bandages, syringes, PPE, etc.). Populated in COVID-19 datasets.';
