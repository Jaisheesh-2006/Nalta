-- Migration: 010_create_allergies.up.sql
-- Patient allergies and intolerances recorded during encounters. Uses SNOMED CT codes.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS allergies (
    START       DATE           NOT NULL  COMMENT 'Date the allergy was first recorded',
    STOP        DATE                     COMMENT 'Date the allergy was resolved; NULL = still active',
    PATIENT     VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER   VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id during which allergy was recorded',
    CODE        VARCHAR(50)    NOT NULL  COMMENT 'SNOMED CT allergen/allergy type code',
    DESCRIPTION VARCHAR(500)             COMMENT 'Sensitive: allergy name (e.g., Allergy to peanuts); reveals clinical history',

    INDEX idx_patient          (PATIENT),
    INDEX idx_encounter        (ENCOUNTER),
    INDEX idx_code             (CODE),
    INDEX idx_patient_code     (PATIENT, CODE),
    INDEX idx_patient_active   (PATIENT, STOP),  -- STOP IS NULL = active allergy
    INDEX idx_start            (START)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Patient allergies and intolerances. Uses SNOMED CT codes. Active allergies have a NULL STOP date.';
