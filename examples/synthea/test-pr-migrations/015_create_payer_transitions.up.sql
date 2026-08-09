-- Migration: 015_create_payer_transitions.up.sql
-- Records of patients changing insurance payers over time. Captures full insurance history.
-- START_YEAR/END_YEAR are integer years, not dates — Synthea models coverage in annual periods.
-- Depends on: patients (003), payers (002)

CREATE TABLE IF NOT EXISTS payer_transitions (
    PATIENT     VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    START_YEAR  SMALLINT       NOT NULL  COMMENT 'Calendar year this payer coverage began',
    END_YEAR    SMALLINT       NOT NULL  COMMENT 'Calendar year this payer coverage ended',
    PAYER       VARCHAR(36)    NOT NULL  COMMENT 'FK → payers.Id — the insurance payer during this period',
    OWNERSHIP   VARCHAR(50)              COMMENT 'Policy ownership: Self | Spouse | Guardian | Medicare | Medicaid | etc.',

    INDEX idx_patient             (PATIENT),
    INDEX idx_payer               (PAYER),
    INDEX idx_patient_years       (PATIENT, START_YEAR, END_YEAR),
    INDEX idx_payer_year          (PAYER, START_YEAR),
    INDEX idx_ownership           (OWNERSHIP)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Insurance coverage history per patient. Tracks which payer covered each patient and for what years.';
