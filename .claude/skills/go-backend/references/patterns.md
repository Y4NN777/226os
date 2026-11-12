# go-backend — reference patterns

Illustrative shapes, not approved code. Check the real packages under `src/api/internal/platform` before reusing a name; the domain modules follow the layout in `SKILL.md`.

## Domain errors and transport mapping
```go
// internal/solve/domain/errors.go
var (
    ErrNotFound   = errors.New("not found")
    ErrForbidden  = errors.New("forbidden")
    ErrConflict   = errors.New("conflict")
)
type ValidationError struct{ Violations []Violation }
func (e *ValidationError) Error() string { return "validation failed" }

// internal/solve/repository: translate driver errors once
if errors.Is(err, pgx.ErrNoRows) { return domain.Question{}, fmt.Errorf("get question: %w", domain.ErrNotFound) }
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "23505" { return fmt.Errorf("accept answer: %w", domain.ErrConflict) }

// internal/platform/httpx: single mapping point -> RFC 9457
switch {
case errors.As(err, &vErr):               writeProblem(w, r, 422, "validation-error", vErr.Violations)
case errors.Is(err, domain.ErrForbidden): writeProblem(w, r, 403, "forbidden", nil)
case errors.Is(err, domain.ErrNotFound):  writeProblem(w, r, 404, "not-found", nil)
case errors.Is(err, domain.ErrConflict):  writeProblem(w, r, 409, "conflict", nil)
default: slog.ErrorContext(r.Context(), "unhandled", "err", err); writeProblem(w, r, 500, "internal", nil)
}
```
Never put `err.Error()` into `detail` for 5xx.

## Handler shape (transport only)
```go
func (h *Handler) acceptAnswer(w http.ResponseWriter, r *http.Request) {
    subj, ok := identity.SubjectFrom(r.Context())   // set by session middleware
    if !ok { h.problem(w, r, identity.ErrUnauthenticated); return }
    qID, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil { h.problem(w, r, malformedID); return }
    var in acceptAnswerRequest
    if err := httpx.DecodeStrict(w, r, &in, maxBody); err != nil { h.problem(w, r, err); return }
    out, err := h.svc.AcceptAnswer(r.Context(), subj, qID, in.AnswerID)
    if err != nil { h.problem(w, r, err); return }
    httpx.JSON(w, http.StatusOK, toDTO(out))
}
```
`DecodeStrict` = `http.MaxBytesReader` + `json.Decoder.DisallowUnknownFields()` + reject trailing data.

## Service with transaction port
```go
type Store interface {
    LockQuestion(ctx context.Context, id uuid.UUID) (domain.Question, error) // SELECT ... FOR UPDATE
    SetAccepted(ctx context.Context, qID, aID uuid.UUID) error
}
type TxRunner interface { WithTx(ctx context.Context, fn func(Store) error) error }

func (s *Service) AcceptAnswer(ctx context.Context, sub identity.Subject, qID, aID uuid.UUID) (domain.Question, error) {
    ctx, cancel := context.WithTimeout(ctx, s.dbTimeout); defer cancel()
    var q domain.Question
    err := s.tx.WithTx(ctx, func(st Store) error {
        var err error
        if q, err = st.LockQuestion(ctx, qID); err != nil { return err }
        if !q.OwnedBy(sub.UserID) { return domain.ErrForbidden }   // authz inside tx boundary
        return st.SetAccepted(ctx, qID, aID)                         // DB partial unique index is the backstop
    })
    return q, err
}
```
Authorization and mutation in the same transaction: `docs/spec-kit/AUTHORIZATION_MATRIX.md:47`.

## Server and shutdown
```go
srv := &http.Server{Addr: cfg.Addr, Handler: router,
    ReadHeaderTimeout: 5*time.Second, ReadTimeout: 15*time.Second,
    WriteTimeout: 30*time.Second, IdleTimeout: 60*time.Second} // values pin per ADR / perf baseline
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
go func() { _ = srv.ListenAndServe() }() // log non-ErrServerClosed
<-ctx.Done()
shCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second); defer cancel()
_ = srv.Shutdown(shCtx); pool.Close()
```
Timeout numbers above are UNVERIFIED placeholders; set from the G6 perf baseline (`docs/RULES.md:30`).

## Table-driven test skeleton
```go
func TestAcceptAnswer(t *testing.T) {
    tests := []struct{ name string; sub identity.Subject; seed fakeStore; wantErr error }{
        {"owner accepts", alice, seeded, nil},
        {"non-owner forbidden", bob, seeded, domain.ErrForbidden},
        {"missing question", alice, empty, domain.ErrNotFound},
        {"store failure propagates", alice, failing, errBoom},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            _, err := newSvc(tc.seed).AcceptAnswer(context.Background(), tc.sub, qID, aID)
            if !errors.Is(err, tc.wantErr) { t.Fatalf("got %v want %v", err, tc.wantErr) }
        })
    }
}
```
Sources: https://go.dev/wiki/TableDrivenTests · https://pkg.go.dev/github.com/jackc/pgx/v5 · https://pkg.go.dev/net/http#Server.Shutdown
