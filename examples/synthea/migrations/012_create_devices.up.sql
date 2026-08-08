-- Migration: 012_create_devices.up.sql
-- Medical devices implanted in or used by patients (e.g., pacemakers, insulin pumps, stents).
-- The UDI (Unique Device Identifier) is a globally unique regulated identifier — treat as sensitive.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS devices (
    START       DATETIME       NOT NULL  COMMENT 'Datetime device was implanted or first used',
    STOP        DATETIME                 COMMENT 'Datetime device was removed or deactivated; NULL = still in use',
    PATIENT     VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER   VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id during which device was placed',
    CODE        VARCHAR(50)    NOT NULL  COMMENT 'SNOMED CT device type code',
    DESCRIPTION VARCHAR(500)             COMMENT 'Sensitive: device name (e.g., Implantable cardiac pacemaker); reveals clinical history',
    UDI         VARCHAR(255)             COMMENT 'Sensitive: FDA Unique Device Identifier — globally unique physical device ID',

    INDEX idx_patient          (PATIENT),
    INDEX idx_encounter        (ENCOUNTER),
    INDEX idx_code             (CODE),
    INDEX idx_patient_code     (PATIENT, CODE),
    INDEX idx_patient_active   (PATIENT, STOP),  -- STOP IS NULL = device still active
    INDEX idx_udi              (UDI),
    INDEX idx_start            (START)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Medical devices implanted or assigned to patients. UDI is a regulated globally-unique device identifier.';
