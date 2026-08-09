package main

import (
	"context"
	"database/sql"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestIntrospectSchema_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	colRows := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT"}).
		AddRow("ingredients", "id", "int", "NO", "").
		AddRow("ingredients", "toxicity_class", "text", "YES", "").
		AddRow("products", "id", "int", "NO", "")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COALESCE(COLUMN_DEFAULT, '')")).WillReturnRows(colRows)

	fkRows := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME"})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME")).WillReturnRows(fkRows)

	tblRows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE"}).
		AddRow("ingredients", "BASE TABLE").
		AddRow("products", "BASE TABLE")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, TABLE_TYPE")).WillReturnRows(tblRows)

	ctx := context.Background()
	out, err := IntrospectSchema(ctx, db)
	if err != nil {
		t.Fatalf("IntrospectSchema failed: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(out))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestIntrospectSchema_RetryThenSuccess(t *testing.T) {
	rand.Seed(1)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	// First attempt fails
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COALESCE(COLUMN_DEFAULT, '')")).WillReturnError(sql.ErrConnDone)
	// Second attempt succeeds
	colRows := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT"}).AddRow("t1", "c1", "int", "NO", "")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COALESCE(COLUMN_DEFAULT, '')")).WillReturnRows(colRows)

	fkRows := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME"})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME")).WillReturnRows(fkRows)

	tblRows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE"}).AddRow("t1", "BASE TABLE")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, TABLE_TYPE")).WillReturnRows(tblRows)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := IntrospectSchema(ctx, db)
	if err != nil {
		t.Fatalf("IntrospectSchema failed after retries: %v", err)
	}
	if len(out) != 1 || out[0].Name != "t1" {
		t.Fatalf("unexpected result: %#v", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestIntrospectSchema_SkipView(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	colRows := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT"}).
		AddRow("v_view", "vcol", "int", "NO", "").
		AddRow("t", "id", "int", "NO", "")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COALESCE(COLUMN_DEFAULT, '')")).WillReturnRows(colRows)

	fkRows := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME"})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME")).WillReturnRows(fkRows)

	tableRows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE"}).
		AddRow("v_view", "VIEW").
		AddRow("t", "BASE TABLE")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME, TABLE_TYPE")).WillReturnRows(tableRows)

	ctx := context.Background()
	out, err := IntrospectSchema(ctx, db)
	if err != nil {
		t.Fatalf("IntrospectSchema failed: %v", err)
	}
	if len(out) != 1 || out[0].Name != "t" {
		t.Fatalf("expected only base table 't', got: %#v", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
