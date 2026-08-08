-- Migration: 005_create_encounters.up.sql
-- Clinical encounters (hospital stays, ER visits, wellness checks, etc.).
-- Central join table for all clinical events — most tables FK into this.
-- Depends on: patients (003), organizations (001), providers (004), payers (002)

CREATE TABLE IF NOT EXISTS encounters (
    Id                   VARCHAR(36)     NOT NULL,
    START                DATETIME        NOT NULL,
    STOP                 DATETIME,
    PATIENT              VARCHAR(36)     NOT NULL  COMMENT 'FK → patients.Id',
    ORGANIZATION         VARCHAR(36)               COMMENT 'FK → organizations.Id',
    PROVIDER             VARCHAR(36)               COMMENT 'FK → providers.Id',
    PAYER                VARCHAR(36)               COMMENT 'FK → payers.Id',
    ENCOUNTERCLASS       VARCHAR(50)     NOT NULL  COMMENT 'ambulatory | emergency | inpatient | wellness | urgentcare | outpatient',
    CODE                 VARCHAR(50)     NOT NULL  COMMENT 'SNOMED CT encounter type code',
    DESCRIPTION          VARCHAR(500),
    BASE_ENCOUNTER_COST  DECIMAL(12,4)   NOT NULL DEFAULT 0  COMMENT 'Sensitive: base cost before insurance in USD',
    TOTAL_CLAIM_COST     DECIMAL(12,4)   NOT NULL DEFAULT 0  COMMENT 'Sensitive: total billed amount in USD',
    PAYER_COVERAGE       DECIMAL(12,4)   NOT NULL DEFAULT 0  COMMENT 'Sensitive: insurance-covered amount in USD',
    REASONCODE           VARCHAR(50)               COMMENT 'SNOMED CT code for primary reason',
    REASONDESCRIPTION    VARCHAR(500)              COMMENT 'Sensitive: human-readable primary reason; may reveal diagnosis',

    PRIMARY KEY (Id),
    INDEX idx_patient                  (PATIENT),
    INDEX idx_organization             (ORGANIZATION),
    INDEX idx_provider                 (PROVIDER),
    INDEX idx_payer                    (PAYER),
    INDEX idx_start                    (START),
    INDEX idx_encounterclass           (ENCOUNTERCLASS),
    INDEX idx_patient_start            (PATIENT, START),
    INDEX idx_patient_encounterclass   (PATIENT, ENCOUNTERCLASS),
    INDEX idx_code                     (CODE)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Clinical encounters between patients and providers. Central join table for all clinical events.';
