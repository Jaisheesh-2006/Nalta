-- Migration: 007_create_medications.up.sql
-- Medications prescribed or dispensed. Uses RxNorm codes.
-- Largest cost-data table — all financial columns are sensitive.
-- Depends on: patients (003), encounters (005), payers (002)

CREATE TABLE IF NOT EXISTS medications (
    START             DATE           NOT NULL  COMMENT 'Date prescribed',
    STOP              DATE                     COMMENT 'Date stopped; NULL = still active',
    PATIENT           VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    PAYER             VARCHAR(36)              COMMENT 'FK → payers.Id',
    ENCOUNTER         VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id',
    CODE              VARCHAR(50)    NOT NULL  COMMENT 'RxNorm medication code',
    DESCRIPTION       VARCHAR(500)             COMMENT 'Sensitive: medication name; may reveal underlying conditions',
    BASE_COST         DECIMAL(12,4)            COMMENT 'Sensitive: list price before insurance in USD',
    PAYER_COVERAGE    DECIMAL(12,4)            COMMENT 'Sensitive: amount covered by insurance in USD',
    DISPENSES         INT            NOT NULL DEFAULT 1,
    TOTALCOST         DECIMAL(12,4)            COMMENT 'Sensitive: cumulative cost of all dispenses in USD',
    REASONCODE        VARCHAR(50)              COMMENT 'SNOMED CT code for prescribing reason',
    REASONDESCRIPTION VARCHAR(500)             COMMENT 'Sensitive: human-readable prescribing reason; may reveal diagnosis',

    INDEX idx_patient               (PATIENT),
    INDEX idx_encounter             (ENCOUNTER),
    INDEX idx_payer                 (PAYER),
    INDEX idx_code                  (CODE),
    INDEX idx_patient_code          (PATIENT, CODE),
    INDEX idx_patient_active        (PATIENT, STOP),   -- STOP IS NULL = still prescribed
    INDEX idx_start                 (START),
    INDEX idx_reasoncode            (REASONCODE)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Medications prescribed or dispensed to patients. Uses RxNorm codes. Includes full cost breakdown.';
