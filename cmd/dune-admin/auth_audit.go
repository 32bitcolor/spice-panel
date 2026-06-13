package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

// openAuditLog opens (or creates) an append-only JSON-lines audit log.
// O_APPEND keeps single-line writes atomic on POSIX, so concurrent handlers
// cannot interleave entries.
func openAuditLog(path string) (*slog.Logger, func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, nil, fmt.Errorf("create audit dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304,G702 -- path is operator-configured config dir
	if err != nil {
		return nil, nil, fmt.Errorf("open audit log: %w", err)
	}
	logger := slog.New(slog.NewJSONHandler(f, nil))
	return logger, func() { _ = f.Close() }, nil
}

// installAuditSink opens the audit log and installs the middleware sink.
// Returns a restore func that closes the file and removes the sink.
func installAuditSink(path string) (func(), error) {
	logger, closeFn, err := openAuditLog(path)
	if err != nil {
		return nil, err
	}
	setAuditSink(func(claims *sessionClaims, r *http.Request, status int) {
		logger.Info("mutation",
			"user", claims.Sub,
			"name", claims.Name,
			"auth", claims.Method,
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"ip", clientIP(r),
		)
	})
	return func() {
		setAuditSink(nil)
		closeFn()
	}, nil
}

// initAuditLog wires the audit sink when auth is enabled. Idempotent across
// live config re-applies.
var auditRestore func()

func initAuditLog() {
	if auditRestore != nil {
		return
	}
	restore, err := installAuditSink(filepath.Join(configDir(), "audit.log"))
	if err != nil {
		logAuthError("audit log unavailable (mutations will not be recorded): " + err.Error())
		return
	}
	auditRestore = restore
}
