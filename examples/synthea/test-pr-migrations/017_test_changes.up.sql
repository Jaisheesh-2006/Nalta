-- Migration: 017_test_changes.up.sql
-- Test migration for E2E validation of schema-guard.
-- Drops a PII column (PASSPORT) and adds a new undocumented table.

ALTER TABLE patients DROP COLUMN PASSPORT;

CREATE TABLE IF NOT EXISTS audit_log (
    id          INT             AUTO_INCREMENT PRIMARY KEY,
    actor       VARCHAR(255)    NOT NULL COMMENT 'User or system that performed the action',
    action      VARCHAR(255)    NOT NULL COMMENT 'Description of what changed',
    target_table VARCHAR(100)   COMMENT 'Table affected by the action',
    target_id   VARCHAR(36)     COMMENT 'Primary key of the affected row',
    created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_actor      (actor),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Audit trail for compliance and debugging. Not yet documented in context.yaml.';
