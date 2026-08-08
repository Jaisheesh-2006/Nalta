-- Migration: 002_create_payers.up.sql
-- Insurance payers (health plans) with financial performance metrics.
-- Created early — referenced by encounters, medications, payer_transitions.

CREATE TABLE IF NOT EXISTS payers (
    Id                      VARCHAR(36)     NOT NULL,
    NAME                    VARCHAR(255)    NOT NULL,
    ADDRESS                 VARCHAR(255),
    CITY                    VARCHAR(100),
    STATE_HEADQUARTERED      VARCHAR(50),
    ZIP                     VARCHAR(20),
    PHONE                   VARCHAR(30),
    AMOUNT_COVERED          DECIMAL(15,4)   NOT NULL DEFAULT 0,
    AMOUNT_UNCOVERED        DECIMAL(15,4)   NOT NULL DEFAULT 0,
    REVENUE                 DECIMAL(15,4)   NOT NULL DEFAULT 0,
    COVERED_ENCOUNTERS      INT             NOT NULL DEFAULT 0,
    UNCOVERED_ENCOUNTERS    INT             NOT NULL DEFAULT 0,
    COVERED_MEDICATIONS     INT             NOT NULL DEFAULT 0,
    UNCOVERED_MEDICATIONS   INT             NOT NULL DEFAULT 0,
    COVERED_PROCEDURES      INT             NOT NULL DEFAULT 0,
    UNCOVERED_PROCEDURES    INT             NOT NULL DEFAULT 0,
    COVERED_IMMUNIZATIONS   INT             NOT NULL DEFAULT 0,
    UNCOVERED_IMMUNIZATIONS INT             NOT NULL DEFAULT 0,
    UNIQUE_CUSTOMERS        INT             NOT NULL DEFAULT 0,
    QOLS_AVG                DECIMAL(10,6),
    MEMBER_MONTHS           INT             NOT NULL DEFAULT 0,

    PRIMARY KEY (Id),
    INDEX idx_name  (NAME),
    INDEX idx_state (STATE_HEADQUARTERED)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Insurance payers (health plans) with aggregate financial and coverage metrics.';
