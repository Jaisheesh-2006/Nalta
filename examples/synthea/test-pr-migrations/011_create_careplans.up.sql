-- Migration: 011_create_careplans.up.sql
-- Care plans assigned to patients with clinical goals and recommended activities.
-- Uses SNOMED CT codes. Multiple plans can be active simultaneously per patient.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS careplans (
    Id                VARCHAR(36)    NOT NULL  COMMENT 'UUID uniquely identifying this care plan',
    START             DATE           NOT NULL  COMMENT 'Date the care plan began',
    STOP              DATE                     COMMENT 'Date the care plan ended; NULL = still active',
    PATIENT           VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER         VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id during which plan was created',
    CODE              VARCHAR(50)    NOT NULL  COMMENT 'SNOMED CT care plan type code',
    DESCRIPTION       VARCHAR(500)             COMMENT 'Sensitive: care plan name (e.g., Diabetes self-management plan)',
    REASONCODE        VARCHAR(50)              COMMENT 'SNOMED CT code for the condition driving this plan',
    REASONDESCRIPTION VARCHAR(500)             COMMENT 'Sensitive: human-readable reason; may reveal underlying diagnosis',

    PRIMARY KEY (Id),
    INDEX idx_patient              (PATIENT),
    INDEX idx_encounter            (ENCOUNTER),
    INDEX idx_code                 (CODE),
    INDEX idx_reasoncode           (REASONCODE),
    INDEX idx_patient_active       (PATIENT, STOP),  -- STOP IS NULL = active plan
    INDEX idx_patient_code         (PATIENT, CODE),
    INDEX idx_start                (START)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Care plans assigned to patients outlining goals and treatment activities. Uses SNOMED CT codes.';
