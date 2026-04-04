package runit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestResolveServiceDir_SVDIR(t *testing.T) {
	dir := t.TempDir()
	getenv := func(k string) string {
		if k == "SVDIR" {
			return dir
		}
		return os.Getenv(k)
	}
	got, err := ResolveServiceDir(getenv)
	if err != nil || got != filepath.Clean(dir) {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestListServiceNames(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_ = os.Mkdir(filepath.Join(dir, "nginx"), 0o755)
	_ = os.Mkdir(filepath.Join(dir, "cron"), 0o755)
	_ = os.Mkdir(filepath.Join(dir, ".hidden"), 0o755)
	names, err := ListServiceNames(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "cron" || names[1] != "nginx" {
		t.Fatalf("%#v", names)
	}
}

func TestListServiceNames_notADir(t *testing.T) {
	t.Parallel()
	f := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ListServiceNames(f)
	if err == nil || !errors.Is(err, svcerr.ErrBackendUnavailable) {
		t.Fatalf("want %v, got %v", svcerr.ErrBackendUnavailable, err)
	}
}
