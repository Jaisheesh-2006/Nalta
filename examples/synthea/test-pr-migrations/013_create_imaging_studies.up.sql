-- Migration: 013_create_imaging_studies.up.sql
-- Radiological and imaging studies (X-ray, MRI, CT scan, Ultrasound, etc.).
-- Uses DICOM standard codes for modality and SOP class. Body site uses SNOMED CT.
-- Depends on: patients (003), encounters (005)

CREATE TABLE IF NOT EXISTS imaging_studies (
    Id                   VARCHAR(36)    NOT NULL  COMMENT 'UUID uniquely identifying this imaging study',
    DATE                 DATETIME       NOT NULL  COMMENT 'Datetime the study was performed',
    PATIENT              VARCHAR(36)    NOT NULL  COMMENT 'FK → patients.Id',
    ENCOUNTER            VARCHAR(36)    NOT NULL  COMMENT 'FK → encounters.Id',
    BODYSITE_CODE        VARCHAR(50)              COMMENT 'SNOMED CT code for the body site imaged',
    BODYSITE_DESCRIPTION VARCHAR(500)             COMMENT 'Sensitive: human-readable body site (e.g., Thoracic structure); may reveal clinical concern',
    MODALITY_CODE        VARCHAR(50)              COMMENT 'DICOM modality code (e.g., DX = Digital Radiography, MR = MRI, CT = CT scan)',
    MODALITY_DESCRIPTION VARCHAR(255)             COMMENT 'Human-readable imaging modality name',
    SOP_CODE             VARCHAR(100)             COMMENT 'DICOM SOP class code identifying the image type produced',
    SOP_DESCRIPTION      VARCHAR(500)             COMMENT 'Human-readable description of the DICOM SOP image type',

    PRIMARY KEY (Id),
    INDEX idx_patient              (PATIENT),
    INDEX idx_encounter            (ENCOUNTER),
    INDEX idx_date                 (DATE),
    INDEX idx_modality             (MODALITY_CODE),
    INDEX idx_bodysite             (BODYSITE_CODE),
    INDEX idx_patient_date         (PATIENT, DATE),
    INDEX idx_patient_modality     (PATIENT, MODALITY_CODE)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Radiological and imaging studies (X-ray, MRI, CT, Ultrasound). Uses DICOM modality codes and SNOMED CT body site codes.';
