// Package script evaluates small Python and JavaScript programs using the
// locally installed runtimes. Execution is full-trust local code; this
// package does not provide a sandbox.
package script

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxStdout = 8 << 20
	maxStderr = 64 << 10
)

// Run evaluates code as an expression with data available as data. Python
// expressions also have math and json available; JavaScript expressions have
// Math and JSON available. The result must be strict JSON.
func Run(ctx context.Context, language, code string, data any) (any, error) {
	if ctx == nil {
		return nil, errors.New("script: nil context")
	}
	lang, err := normalize(language)
	if err != nil {
		return nil, err
	}
	input, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("script: encode data: %w", err)
	}

	var argv []string
	var wrapper string
	switch lang {
	case "python":
		wrapper = pythonExpression(code)
		argv = []string{"-c", wrapper}
	case "javascript":
		wrapper = javascriptExpression(code)
		argv = []string{"-e", wrapper}
	}
	return execute(ctx, lang, argv, input)
}

// RunFile executes a script file. Python and JavaScript files must define a
// main(data) function whose return value is JSON serializable.
func RunFile(ctx context.Context, language, path string, data any) (any, error) {
	if ctx == nil {
		return nil, errors.New("script: nil context")
	}
	lang, err := normalize(language)
	if err != nil {
		return nil, err
	}
	code, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("script: read %q: %w", path, err)
	}
	input, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("script: encode data: %w", err)
	}
	var argv []string
	switch lang {
	case "python":
		argv = []string{"-c", pythonFile(string(code))}
	case "javascript":
		argv = []string{"-e", javascriptFile(string(code))}
	}
	return execute(ctx, lang, argv, input)
}

func normalize(language string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "python", "python3", "py":
		return "python", nil
	case "javascript", "js", "node", "nodejs":
		return "javascript", nil
	default:
		return "", fmt.Errorf("script: unsupported language %q (want python or javascript)", language)
	}
}

func pythonExpression(code string) string {
	return "import json, math, sys\n" +
		"data=json.load(sys.stdin)\n" +
		"try:\n" +
		" result=eval(" + repr(code) + ", {'__builtins__': __builtins__, 'data': data, 'math': math, 'json': json})\n" +
		" print(json.dumps(result, allow_nan=False, separators=(',', ':')))\n" +
		"except Exception as e:\n" +
		" raise SystemExit('script expression: '+str(e))\n"
}

func pythonFile(code string) string {
	return "import json, sys\n" + "data=json.load(sys.stdin)\n" + code + "\n" +
		"try:\n result=main(data)\n print(json.dumps(result, allow_nan=False, separators=(',', ':')))\n" +
		"except Exception as e:\n raise SystemExit('script main: '+str(e))\n"
}

func javascriptExpression(code string) string {
	return "const fs=require('fs'); const data=JSON.parse(fs.readFileSync(0,'utf8'));\n" +
		"let result; try { result=(" + code + "); } catch (e) { throw new Error('script expression: '+e.message); }\n" +
		"const out=JSON.stringify(result, (k,v) => { if (typeof v === 'number' && !Number.isFinite(v)) throw new Error('non-finite number'); return v; }); if (out === undefined) throw new Error('result is not JSON serializable'); process.stdout.write(out+'\\n');\n"
}

func javascriptFile(code string) string {
	return "const fs=require('fs'); const data=JSON.parse(fs.readFileSync(0,'utf8'));\n" + code + "\n" +
		"let result; try { result=main(data); } catch (e) { throw new Error('script main: '+e.message); }\n" +
		"const out=JSON.stringify(result, (k,v) => { if (typeof v === 'number' && !Number.isFinite(v)) throw new Error('non-finite number'); return v; }); if (out === undefined) throw new Error('result is not JSON serializable'); process.stdout.write(out+'\\n');\n"
}

func repr(s string) string { b, _ := json.Marshal(s); return string(b) }

func execute(ctx context.Context, lang string, argv []string, input []byte) (any, error) {
	name := "python3"
	if lang == "javascript" {
		name = "node"
	}
	runtime, err := exec.LookPath(name)
	if err != nil {
		return nil, fmt.Errorf("script: %s runtime %q not found: %w", lang, name, err)
	}
	cmd := exec.CommandContext(ctx, runtime, argv...)
	cmd.WaitDelay = 200 * time.Millisecond
	cmd.Stdin = bytes.NewReader(input)
	stdout, stderr := &limitedBuffer{limit: maxStdout}, &limitedBuffer{limit: maxStderr}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err = cmd.Run()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("script: %s: %w", lang, ctx.Err())
	}
	if stdout.exceeded {
		return nil, fmt.Errorf("script: %s output exceeds %d bytes", lang, maxStdout)
	}
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = err.Error()
		}
		return nil, fmt.Errorf("script: %s failed: %s", lang, detail)
	}
	dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	dec.UseNumber()
	var result any
	if err := dec.Decode(&result); err != nil {
		return nil, fmt.Errorf("script: %s returned invalid JSON: %w", lang, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("script: %s returned more than one JSON value", lang)
	}
	return result, nil
}

type limitedBuffer struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
}

func (b *limitedBuffer) Len() int       { return b.buffer.Len() }
func (b *limitedBuffer) Bytes() []byte  { return b.buffer.Bytes() }
func (b *limitedBuffer) String() string { return b.buffer.String() }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > b.limit {
		b.exceeded = true
		n := b.limit - b.Len()
		if n > 0 {
			_, _ = b.buffer.Write(p[:n])
		}
		return n, io.ErrShortWrite
	}
	return b.buffer.Write(p)
}
