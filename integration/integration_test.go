//go:build integration
// +build integration

package integration

import (
    "database/sql"
    "encoding/json"
    "os/exec"
    "testing"
    "time"
    "io"
    "os"
    "bytes"
    "reflect"
    "path/filepath"

    _ "github.com/go-sql-driver/mysql"
)

func TestIntegration_IntrospectAndMergeAgainstExamples(t *testing.T) {
    // Determine project root (one level up from this integration package)
    wd, err := os.Getwd()
    if err != nil {
        t.Fatalf("getwd: %v", err)
    }
    projectRoot := filepath.Clean(filepath.Join(wd, ".."))

    // Bring up docker compose
    cmd := exec.Command("docker", "compose", "up", "-d")
    cmd.Dir = projectRoot
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("docker compose up failed: %v\n%s", err, string(out))
    }

    // Build server binary
    build := exec.Command("go", "build", "-o", "nalta", ".")
    build.Dir = projectRoot
    if out, err := build.CombinedOutput(); err != nil {
        t.Fatalf("go build failed: %v\n%s", err, string(out))
    }

    // Wait for MySQL to be ready by attempting connections
    dsn := "cosmo:cosmo@tcp(127.0.0.1:3306)/cosmo_db"
    deadline := time.Now().Add(180 * time.Second)
    var db *sql.DB
    var lastPingErr error
    for time.Now().Before(deadline) {
        db, err = sql.Open("mysql", dsn)
        if err == nil {
            if lastPingErr = db.Ping(); lastPingErr == nil {
                break
            }
            db.Close()
        }
        time.Sleep(1 * time.Second)
    }
    if lastPingErr != nil {
        t.Fatalf("failed to connect/ping db after 180s: %v", lastPingErr)
    }
    defer db.Close()

    // Run server to dump schema to stdout
    run := exec.Command("./nalta", "--dsn", dsn, "--context", filepath.Join("examples", "context.yaml"), "--dump-schema", "-")
    run.Dir = projectRoot
    stdout, err := run.StdoutPipe()
    if err != nil {
        t.Fatalf("stdout pipe: %v", err)
    }
    if err := run.Start(); err != nil {
        t.Fatalf("start server: %v", err)
    }

    // read stdout until process exits
    var buf bytes.Buffer
    if _, err := io.Copy(&buf, stdout); err != nil {
        t.Fatalf("reading stdout: %v", err)
    }
    run.Wait()

    // Compare JSON to example
    var got interface{}
    if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
        t.Fatalf("unmarshal server output: %v", err)
    }

    examplePath := filepath.Join(projectRoot, "examples", "schema_full_sample.json")
    exampleData, err := os.ReadFile(examplePath)
    if err != nil {
        t.Fatalf("read example: %v", err)
    }
    var want interface{}
    if err := json.Unmarshal(exampleData, &want); err != nil {
        t.Fatalf("unmarshal example: %v", err)
    }

    if !reflect.DeepEqual(got, want) {
        t.Fatalf("merged schema does not match example\nGot: %s\nWant: %s", string(buf.Bytes()), string(exampleData))
    }

    // Now test explain_column via dump-column
    srv2 := exec.Command("./nalta", "--dsn", dsn, "--context", filepath.Join("examples", "context.yaml"), "--dump-column", "ingredients:toxicity_class")
    srv2.Dir = projectRoot
    out2, err := srv2.Output()
    if err != nil {
        // If it fails, we might want to see the error from stderr, but err will typically capture the exit status.
        // Actually, to get stderr on failure, it's better to use an ExitError cast if needed, but for the test output this is sufficient.
        t.Fatalf("dump-column failed: %v\n%s", err, string(out2))
    }
    var gotCol interface{}
    if err := json.Unmarshal(out2, &gotCol); err != nil {
        t.Fatalf("unmarshal column output: %v", err)
    }

    expectedPath := filepath.Join(projectRoot, "examples", "explain_column_sample.json")
    expectedCol, err := os.ReadFile(expectedPath)
    if err != nil {
        t.Fatalf("read example column: %v", err)
    }
    var wantCol interface{}
    if err := json.Unmarshal(expectedCol, &wantCol); err != nil {
        t.Fatalf("unmarshal expected column: %v", err)
    }
    if !reflect.DeepEqual(gotCol, wantCol) {
        t.Fatalf("column output mismatch\nGot: %s\nWant: %s", string(out2), string(expectedCol))
    }
}
