//go:build integration

package pgtest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// defaultImage is the multi-arch index digest of postgres:18 (PostgreSQL 18.6)
// pulled on 2026-11-11. Override with PGTEST_IMAGE only for a deliberate upgrade.
const defaultImage = "postgres:18@sha256:e3d24f6906d0760bf4da0dae872415391ededefad23fd471c47dc53b0ddf3c42"

const containerLabel = "226os.pgtest=1"

// Why the docker CLI and not testcontainers-go: it avoids a new large
// dependency tree (docker client, ryuk reaper) that would need its own ADR; the
// cost is that a hard-killed test run can leave a container, which carries the
// label above so it is easy to find and remove.

type container struct {
	id       string
	hostPort string
	password string
}

func (c *container) dsn(database string) string {
	return fmt.Sprintf("postgres://postgres:%s@%s/%s?sslmode=disable", c.password, c.hostPort, database)
}

func docker(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func startPostgres(ctx context.Context) (*container, error) {
	image := os.Getenv("PGTEST_IMAGE")
	if image == "" {
		image = defaultImage
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	c := &container{password: hex.EncodeToString(raw)}

	out, err := docker(ctx, "run", "-d", "--rm", "--label", containerLabel,
		"-e", "POSTGRES_PASSWORD="+c.password,
		"-p", "127.0.0.1::5432",
		image,
		"-c", "fsync=off", "-c", "synchronous_commit=off", "-c", "full_page_writes=off")
	if err != nil {
		return nil, fmt.Errorf("docker run %s: %w: %s", image, err, out)
	}
	c.id = out

	out, err = docker(ctx, "port", c.id, "5432/tcp")
	if err != nil {
		c.remove()
		return nil, fmt.Errorf("docker port: %w: %s", err, out)
	}
	line, _, _ := strings.Cut(out, "\n")
	host, port, err := net.SplitHostPort(strings.TrimSpace(line))
	if err != nil {
		c.remove()
		return nil, fmt.Errorf("parse docker port %q: %w", out, err)
	}
	if host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	c.hostPort = net.JoinHostPort(host, port)

	if err := c.waitReady(ctx); err != nil {
		c.remove()
		return nil, err
	}
	return c, nil
}

// waitReady waits for the second "ready to accept connections" log line: the
// image entrypoint runs a temporary server during init, then restarts it.
func (c *container) waitReady(ctx context.Context) error {
	const ready = "database system is ready to accept connections"
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		logs, _ := docker(ctx, "logs", c.id)
		if strings.Count(logs, ready) >= 2 {
			conn, err := pgx.Connect(ctx, c.dsn("postgres"))
			if err == nil {
				var one int
				err = conn.QueryRow(ctx, "SELECT 1").Scan(&one)
				_ = conn.Close(ctx)
				if err == nil {
					return nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	return errors.New("postgres container not ready within 90s")
}

func (c *container) remove() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, _ = docker(ctx, "rm", "-f", c.id)
}

var (
	pg           *container
	templateOnce sync.Once
	templateErr  error
	dbCounter    atomic.Int64
)

const templateDB = "tmpl_migrated"

func TestMain(m *testing.M) { os.Exit(runTests(m)) }

// runTests keeps the container cleanup in a defer: os.Exit in TestMain would
// skip it.
func runTests(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	c, err := startPostgres(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pgtest: cannot start postgres:", err)
		return 1
	}
	defer c.remove()
	pg = c
	return m.Run()
}

func repoRoot(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test directory")
		}
		dir = parent
	}
}

func migrationsFS(t testing.TB) fs.FS {
	t.Helper()
	return os.DirFS(filepath.Join(repoRoot(t), "db", "migrations"))
}

func newProvider(db *sql.DB, fsys fs.FS) (*goose.Provider, error) {
	return goose.NewProvider(goose.DialectPostgres, db, fsys)
}

func openSQL(dsn string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return stdlib.OpenDB(*cfg), nil
}

func adminConn(ctx context.Context, t testing.TB) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(ctx, pg.dsn("postgres"))
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func createDB(ctx context.Context, t testing.TB, template string) string {
	t.Helper()
	name := fmt.Sprintf("t_%d", dbCounter.Add(1))
	q := fmt.Sprintf(`CREATE DATABASE %s`, pgx.Identifier{name}.Sanitize())
	if template != "" {
		q += " TEMPLATE " + pgx.Identifier{template}.Sanitize()
	}
	if _, err := adminConn(ctx, t).Exec(ctx, q); err != nil {
		t.Fatalf("create database: %v", err)
	}
	return name
}

// migratedDB returns a fresh database cloned from a template that has all
// migrations applied from zero (once per run).
func migratedDB(ctx context.Context, t testing.TB) *pgx.Conn {
	t.Helper()
	templateOnce.Do(func() {
		adm, err := pgx.Connect(ctx, pg.dsn("postgres"))
		if err != nil {
			templateErr = err
			return
		}
		defer func() { _ = adm.Close(ctx) }()
		if _, err := adm.Exec(ctx, `CREATE DATABASE `+templateDB); err != nil {
			templateErr = err
			return
		}
		db, err := openSQL(pg.dsn(templateDB))
		if err != nil {
			templateErr = err
			return
		}
		defer func() { _ = db.Close() }()
		p, err := newProvider(db, migrationsFS(t))
		if err != nil {
			templateErr = fmt.Errorf("goose provider: %w", err)
			return
		}
		if _, err := p.Up(ctx); err != nil {
			templateErr = fmt.Errorf("goose up: %w", err)
		}
	})
	if templateErr != nil {
		t.Fatalf("template database: %v", templateErr)
	}
	name := createDB(ctx, t, templateDB)
	conn, err := pgx.Connect(ctx, pg.dsn(name))
	if err != nil {
		t.Fatalf("connect %s: %v", name, err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

// pgErr extracts the PostgreSQL error (SQLSTATE and constraint/index name).
func pgErr(err error) *pgconn.PgError {
	var e *pgconn.PgError
	if errors.As(err, &e) {
		return e
	}
	return nil
}
