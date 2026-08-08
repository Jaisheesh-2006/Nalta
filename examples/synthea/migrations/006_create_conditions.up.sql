-- Migration: 006_create_conditions.up.sql
-- Medical conditions/diagnoses using SNOMED CT codes.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS conditions (
    START        DATE           NOT NULL  COMMENT 'Date first diagnosed',
    STOP         DATE                     COMMENT 'Date resolved; NULL = ongoing',
    PATIENT      VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER    VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id',
    CODE         VARCHAR(50)    NOT NULL  COMMENT 'SNOMED CT condition code',
    DESCRIPTION  VARCHAR(500)             COMMENT 'Sensitive: human-readable diagnosis name',

    INDEX idx_patient             (PATIENT),
    INDEX idx_encounter           (ENCOUNTER),
    INDEX idx_code                (CODE),
    INDEX idx_patient_code        (PATIENT, CODE),
    INDEX idx_patient_active      (PATIENT, STOP),   -- STOP IS NULL = active condition
    INDEX idx_start               (START),
    INDEX idx_code_start          (CODE, START)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Patient medical conditions and diagnoses. Uses SNOMED CT codes. A patient may have multiple active or resolved conditions.';
