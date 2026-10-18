package main

import (
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

func TestHandleShellScript(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	if fi, err := os.Stat(shellScriptDir); err != nil || !fi.IsDir() {
		t.Skipf("%s not available", shellScriptDir)
	}

	templ := template.Must(template.New("test").Funcs(shellFuncs).Parse(
		"pwd\necho {{ .msg | shq }}\necho to stderr >&2\nexit {{ .code | shq }}\n"))
	call := func(msg string, code float64) *mcp.CallToolResult {
		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]any{"msg": msg, "code": code}
		res, err := handleShellScript(templ, req)
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
