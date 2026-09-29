package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"text/template"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mdzio/ccu-ai-mcp/config"
)

func TestShellQuote(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"abc", `'abc'`},
		{"", `''`},
		{nil, `''`},
		{"it's", `'it'\''s'`},
		{"a b; rm -rf /", `'a b; rm -rf /'`},
		{"$HOME `id`", "'$HOME `id`'"},
		{float64(100), `'100'`},
		{float64(1000000), `'1000000'`},
		{float64(1.5), `'1.5'`},
		{true, `'true'`},
		{[]any{"a", "b c", float64(3)}, `'a' 'b c' '3'`},
		{[]any{}, ``},
	}
	for _, c := range cases {
		if got := shellQuote(c.in); got != c.want {
			t.Errorf("shellQuote(%#v) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestIsLocalAddress(t *testing.T) {
	cases := []struct {
		addr string
		want bool
	}{
		{"localhost", true},
		{"LOCALHOST", true},
		{"localhost:2001", true},
		{"127.0.0.1", true},
		{"127.0.0.1:8181", true},
		{"127.1.2.3", true},
		{"::1", true},
		{"[::1]:8181", true},
		{"homematic-raspi", false},
		{"192.168.1.10", false},
		{"192.168.1.10:8181", false},
		{"ccu.local", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isLocalAddress(c.addr); got != c.want {
			t.Errorf("isLocalAddress(%q) = %v, want %v", c.addr, got, c.want)
		}
	}
}

func TestLimitedBuffer(t *testing.T) {
	b := &limitedBuffer{limit: 5}
	if n, err := b.Write([]byte("abc")); n != 3 || err != nil {
		t.Fatalf("Write returned %d, %v", n, err)
	}
	if b.truncated {
		t.Fatal("buffer must not be truncated yet")
	}
	if n, err := b.Write([]byte("defgh")); n != 5 || err != nil {
		t.Fatalf("Write returned %d, %v", n, err)
	}
	if !b.truncated || b.String() != "abcde" {
		t.Fatalf("got %q (truncated=%v), want \"abcde\" (truncated=true)", b.String(), b.truncated)
	}
	// further writes are discarded
	if n, err := b.Write([]byte("xyz")); n != 3 || err != nil || b.String() != "abcde" {
		t.Fatalf("got %q, %d, %v", b.String(), n, err)
	}
}

func TestHandleShellScript(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	if fi, err := os.Stat(shellScriptDir); err != nil || !fi.IsDir() {
		t.Skipf("%s not available", shellScriptDir)
	}

	toolCfg := config.Tool{Name: "test", Kind: config.ShellScript}
	templ := template.Must(template.New("test").Funcs(shellFuncs).Parse(
		"pwd\necho {{ .msg | shq }}\necho to stderr >&2\nexit {{ .code | shq }}\n"))
	call := func(msg string, code float64) *mcp.CallToolResult {
		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]any{"msg": msg, "code": code}
		res, err := handleShellScript(context.Background(), toolCfg, templ, req)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		return res
	}
	text := func(res *mcp.CallToolResult) string {
		if len(res.Content) != 1 {
			t.Fatalf("unexpected content: %#v", res.Content)
		}
		return res.Content[0].(mcp.TextContent).Text
	}

	// successful execution
	res := call("hello 'world'", 0)
	if res.IsError {
		t.Fatalf("unexpected error result: %s", text(res))
	}
	if got, want := text(res), shellScriptDir+"\nhello 'world'\nto stderr\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// failed execution
	res = call("x", 3)
	if !res.IsError {
		t.Fatalf("expected error result, got: %s", text(res))
	}
	if got := text(res); !strings.HasPrefix(got, "ERROR: Shell script failed with exit code 3\n") || !strings.Contains(got, "to stderr\n") {
		t.Errorf("unexpected error text: %q", got)
	}
}

// TestToolsConfigTemplates checks that all script templates of the shipped
// tools.toml are parseable.
func TestToolsConfigTemplates(t *testing.T) {
	tools, err := config.LoadTools("tools.toml")
	if err != nil {
		t.Fatalf("loading tools.toml failed: %v", err)
	}
	if len(tools.Tools) == 0 {
		t.Fatal("no tools found")
	}
	for _, toolCfg := range tools.Tools {
		templ := template.New(toolCfg.Name)
		if toolCfg.Kind == config.ShellScript {
			templ = templ.Funcs(shellFuncs)
		}
		if _, err := templ.Parse(toolCfg.Script); err != nil {
			t.Errorf("parsing template of tool %s failed: %v", toolCfg.Name, err)
		}
	}
}
