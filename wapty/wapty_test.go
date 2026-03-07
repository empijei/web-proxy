package wapty_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/empijei/tst"
	"github.com/empijei/web-proxy/wapty"
)

func TestWaptyRun(t *testing.T) {
	t.Skip("This is to run, not to test.")
	home := tst.Do(os.UserHomeDir())(t)
	dir := filepath.Join(home, ".wapty")
	tst.No(os.MkdirAll(dir, 0o700), t)
	tst.No(wapty.Run(t.Context(), dir), t)
}
