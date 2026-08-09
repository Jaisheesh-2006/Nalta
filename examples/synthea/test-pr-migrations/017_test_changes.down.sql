-- Migration: 017_test_changes.down.sql
-- Reversal of the test migration: re-adds PASSPORT and drops audit_log.

DROP TABLE IF EXISTS audit_log;

ALTER TABLE patients ADD COLUMN PASSPORT VARCHAR(50) COMMENT 'PII: Passport number' AFTER DRIVERS;
