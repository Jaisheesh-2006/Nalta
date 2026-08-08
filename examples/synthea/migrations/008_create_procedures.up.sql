-- Migration: 008_create_procedures.up.sql
-- Medical procedures performed on patients using SNOMED CT codes.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS procedures (
    DATE              DATE           NOT NULL  COMMENT 'Date the procedure was performed',
    PATIENT           VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER         VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id',
    CODE              VARCHAR(50)    NOT NULL  COMMENT 'SNOMED CT procedure code',
    DESCRIPTION       VARCHAR(500)             COMMENT 'Sensitive: procedure name; reveals clinical history',
    BASE_COST         DECIMAL(12,4)            COMMENT 'Sensitive: base cost before insurance in USD',
    REASONCODE        VARCHAR(50)              COMMENT 'SNOMED CT code for why this procedure was performed',
    REASONDESCRIPTION VARCHAR(500)             COMMENT 'Sensitive: human-readable reason; may reveal diagnosis',

    INDEX idx_patient              (PATIENT),
    INDEX idx_encounter            (ENCOUNTER),
    INDEX idx_code                 (CODE),
    INDEX idx_patient_code         (PATIENT, CODE),
    INDEX idx_date                 (DATE),
    INDEX idx_patient_date         (PATIENT, DATE),
    INDEX idx_reasoncode           (REASONCODE)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Medical procedures performed during encounters. Uses SNOMED CT codes. Includes cost and clinical reason.';
