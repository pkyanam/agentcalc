package script

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestRunExpressions(t *testing.T) {
	for _, tc := range []struct {
		language, code string
	}{
		{"python", "{'answer': data['n'] + 2}"},
		{"javascript", "({answer: data.n + 2})"},
	} {
		t.Run(tc.language, func(t *testing.T) {
			if _, err := exec.LookPath(map[string]string{"python": "python3", "javascript": "node"}[tc.language]); err != nil {
				t.Skip("runtime is not installed")
			}
			got, err := Run(context.Background(), tc.language, tc.code, map[string]any{"n": 5})
			if err != nil {
				t.Fatal(err)
			}
			if got == nil {
				t.Fatal("nil result")
			}
		})
	}
}

func TestRunFileAndErrors(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is not installed")
	}
	path := t.TempDir() + "/main.py"
	if err := writeTestFile(path, "def main(data):\n return data['value'] * 2\n"); err != nil {
		t.Fatal(err)
	}
	got, err := RunFile(context.Background(), "python", path, map[string]any{"value": 4})
	if err != nil || got == nil {
		t.Fatalf("RunFile: %#v, %v", got, err)
	}
	if _, err := Run(context.Background(), "python", "float('nan')", nil); err == nil {
		t.Fatal("expected strict JSON error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := Run(ctx, "python", "(__import__('time').sleep(2), 1)[1]", nil); err == nil {
		t.Fatal("expected timeout error")
	}
}

func writeTestFile(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0600)
}
