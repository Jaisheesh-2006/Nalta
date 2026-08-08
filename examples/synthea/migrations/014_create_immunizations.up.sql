-- Migration: 014_create_immunizations.up.sql
-- Vaccinations and immunizations administered to patients. Uses CVX (CDC vaccine) codes.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS immunizations (
    DATE        DATETIME       NOT NULL  COMMENT 'Datetime the immunization was administered',
    PATIENT     VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER   VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id',
    CODE        VARCHAR(50)    NOT NULL  COMMENT 'CVX vaccine code (CDC standard vaccine identifier)',
    DESCRIPTION VARCHAR(500)             COMMENT 'Human-readable vaccine name (e.g., Influenza seasonal injectable)',
    BASE_COST   DECIMAL(12,4)            COMMENT 'Sensitive: base administration cost in USD',

    INDEX idx_patient          (PATIENT),
    INDEX idx_encounter        (ENCOUNTER),
    INDEX idx_code             (CODE),
    INDEX idx_date             (DATE),
    INDEX idx_patient_code     (PATIENT, CODE),
    INDEX idx_patient_date     (PATIENT, DATE)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Vaccinations and immunizations administered to patients. Uses CDC CVX vaccine codes.';
