package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestBatchValidatesRequiredConvertFields(t *testing.T) {
	var out, stderr bytes.Buffer
	code := Main([]string{"batch"}, strings.NewReader(`{"id":"missing","command":"convert","from":"m","to":"cm"}`+"\n"), &out, &stderr, "test")
	if code != 1 {
		t.Fatalf("batch exit code = %d, want 1", code)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["ok"] != false || !strings.Contains(got["error"].(string), `requires field "value"`) {
		t.Fatalf("unexpected response: %v", got)
	}
}

func TestBatchPreservesZeroID(t *testing.T) {
	var out, stderr bytes.Buffer
	code := Main([]string{"batch"}, strings.NewReader(`{"id":0,"command":"eval","expr":"6*7"}`+"\n"), &out, &stderr, "test")
	if code != 0 {
		t.Fatalf("batch exit code = %d", code)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if id, ok := got["id"].(float64); !ok || id != 0 {
		t.Fatalf("zero ID was not preserved: %v", got)
	}
}

func TestBatchExactResourceLimit(t *testing.T) {
	// The result would require millions of bits; the exact evaluator must
	// reject it before constructing that result.
	literal := strings.Repeat("9", 100)
	line := `{"id":"large","command":"exact","expr":"` + literal + ` ^ 10000"}` + "\n"
	var out, stderr bytes.Buffer
	code := Main([]string{"batch"}, strings.NewReader(line), &out, &stderr, "test")
	if code != 1 {
		t.Fatalf("batch exit code = %d, want 1", code)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["ok"] != false || !strings.Contains(got["error"].(string), "one million bits") {
		t.Fatalf("unexpected response: %v", got)
	}
}
