-- Migration: 009_create_observations.up.sql
-- Clinical observations, lab results, and vital signs. Uses LOINC codes.
-- Largest table in the dataset — 300k+ rows in the sample, millions in 10k dataset.
-- VALUE column is the raw clinical result — always treat as sensitive.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS observations (
    DATE        DATETIME       NOT NULL  COMMENT 'Datetime the observation was recorded',
    PATIENT     VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER   VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id',
    CODE        VARCHAR(50)    NOT NULL  COMMENT 'LOINC code identifying the observation type (e.g., 8302-2 = Body Height)',
    DESCRIPTION VARCHAR(500)             COMMENT 'Human-readable name of what was measured (e.g., Body Height, Blood Glucose)',
    VALUE       VARCHAR(255)             COMMENT 'Sensitive: raw measured result (e.g., 193.3, Negative, Positive)',
    UNITS       VARCHAR(50)              COMMENT 'Unit of measurement (e.g., cm, mmHg, mg/dL)',
    TYPE        VARCHAR(50)              COMMENT 'Data type of VALUE: numeric | text | date',

    INDEX idx_patient              (PATIENT),
    INDEX idx_encounter            (ENCOUNTER),
    INDEX idx_code                 (CODE),
    INDEX idx_date                 (DATE),
    INDEX idx_patient_code         (PATIENT, CODE),
    INDEX idx_patient_date         (PATIENT, DATE),
    INDEX idx_type                 (TYPE),
    INDEX idx_code_date            (CODE, DATE)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Clinical observations, lab results, and vital signs. Uses LOINC codes. Most granular and highest-volume clinical table.';
