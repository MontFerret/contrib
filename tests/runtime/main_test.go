package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/wire/client"
	"github.com/MontFerret/wire/pkg/failure"
)

func TestRuntimeProcess(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "runtime")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build runtime: %v\n%s", err, output)
	}

	version, err := exec.CommandContext(t.Context(), "go", "list", "-m", "-f", "{{if .Replace}}{{.Replace.Version}}{{else}}{{.Version}}{{end}}", "github.com/MontFerret/ferret/v2").Output()
	if err != nil {
		t.Fatal(err)
	}

	wantVersion := strings.TrimPrefix(strings.TrimSpace(string(version)), "v")
	for _, tc := range []struct {
		name   string
		signal os.Signal
		args   []string
	}{
		{"default listener and SIGTERM", syscall.SIGTERM, nil},
		{"explicit listener and SIGINT", os.Interrupt, []string{"--listen=127.0.0.1:0", "--log-level=debug"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary, tc.args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}

			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				if cmd.ProcessState == nil {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
				}
			})

			reader := bufio.NewReader(stdout)
			line, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("read endpoint: %v, exit: %v, stderr: %s", err, cmd.Wait(), stderr.String())
			}

			addr := endpointAddress(t, line)
			remote, err := client.New(ctx, addr, client.WithInsecure())
			if err != nil {
				t.Fatal(err)
			}

			defer remote.Close()

			version, err := remote.Version(ctx)
			if err != nil || version.String() != wantVersion || wantVersion == "" {
				t.Fatalf("handshake version = %q, want %q, error = %v", version, wantVersion, err)
			}

			testQueries(t, ctx, remote)

			if err := cmd.Process.Signal(tc.signal); err != nil {
				t.Fatal(err)
			}

			extra, err := io.ReadAll(reader)
			if err != nil || len(extra) != 0 {
				t.Fatalf("extra stdout = %q, read error = %v", extra, err)
			}

			if err := cmd.Wait(); err != nil {
				t.Fatalf("shutdown: %v\n%s", err, stderr.String())
			}
		})
	}

	occupied, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	defer occupied.Close()

	for _, tc := range []struct{ arg, diagnostic string }{
		{"--listen=" + occupied.Addr().String(), "bind Wire listener"},
		{"--listen=:0", "listen address"},
		{"--listen=0.0.0.0:0", "listen address"},
		{"--listen=localhost:0", "listen address"},
		{"--listen=127.0.0.1:65536", "listen address"},
		{"--log-level=invalid", "log level"},
		{"--param=x:1", "param"},
		{"query.fql", "positional"},
	} {
		t.Run(tc.arg, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary, tc.arg)
			stdout, err := cmd.Output()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || len(stdout) != 0 || !strings.Contains(string(exit.Stderr), tc.diagnostic) {
				t.Fatalf("stdout = %q, exit = %v, want diagnostic %q", stdout, err, tc.diagnostic)
			}
		})
	}
}

func endpointAddress(t *testing.T, line string) string {
	t.Helper()
	address, ok := strings.CutPrefix(line, "wire+tcp://")
	addr, err := netip.ParseAddrPort(strings.TrimSuffix(address, "\n"))
	if !ok || err != nil || addr.Addr().String() != "127.0.0.1" || addr.Port() == 0 {
		t.Fatalf("invalid Wire endpoint announcement %q", line)
	}

	return addr.String()
}

func testQueries(t *testing.T, ctx context.Context, remote api.Runtime) {
	t.Helper()
	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "<html><body><h1>Wire</h1></body></html>")
	}))
	defer page.Close()

	for _, tc := range []struct {
		query string
		want  string
		opts  []api.SessionOption
	}{
		{"RETURN {ok: true}", `{"ok":true}`, nil},
		{"RETURN @lab.fixture", `["one",2]`, []api.SessionOption{api.WithParam("lab", map[string]any{"fixture": []any{"one", 2}})}},
		{`RETURN TOML::DECODE("count = 7")`, `{"count":7}`, nil},
		{`RETURN INNER_TEXT(DOCUMENT(@url), "h1")`, `"Wire"`, []api.SessionOption{api.WithParam("url", page.URL)}},
	} {
		result, err := runAndCollect(ctx, remote, tc.query, tc.opts...)
		if err != nil || result == nil || result.Metadata.ContentType != "application/json" || string(result.Data) != tc.want {
			t.Fatalf("%s: result = %#v, error = %v", tc.query, result, err)
		}
	}

	_, err := remote.Compile(ctx, api.NewSource("bad.fql", "RETURN ("))
	var diagnostic *client.Error
	if !errors.As(err, &diagnostic) || diagnostic.Category != failure.CategoryCompilation || len(diagnostic.Diagnostics) == 0 || diagnostic.Diagnostics[0].Source.Name != "bad.fql" {
		t.Fatalf("expected compilation diagnostic for bad.fql, got %v", err)
	}

	_, err = runAndCollect(ctx, remote, `RETURN DB::SQLITE::CLOSE("invalid")`)
	var execution *failure.Failure
	if !errors.As(err, &execution) || execution.Category != failure.CategoryExecution || len(execution.Diagnostics) == 0 {
		t.Fatalf("expected execution diagnostic, got %T %v", err, err)
	}
}

func runAndCollect(ctx context.Context, remote api.Runtime, query string, opts ...api.SessionOption) (content *api.Content, err error) {
	output, err := remote.Run(ctx, api.NewSource("test.fql", query), opts...)
	if err != nil {
		return nil, err
	}

	defer func() { err = errors.Join(err, output.Close()) }()

	return output.Collect(ctx)
}
