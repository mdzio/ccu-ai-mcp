package config

import (
	"errors"
	"fmt"
	"slices"

	"github.com/BurntSushi/toml"
)

// ToolKind represents the type of a tool (e.g. HM-SCRIPT, XML-RPC).
type ToolKind int

const (
	HMScript ToolKind = iota
	XMLRPC
)

var (
	toolKindStr = []string{
		HMScript: "hm-script",
		XMLRPC:   "xml-rpc",
	}
	errToolKind = errors.New("invalid tool kind identifier")
)

// String implements interface Stringer.
func (k ToolKind) String() string {
	if int(k) >= 0 && int(k) < len(toolKindStr) {
		return toolKindStr[k]
	}
	return fmt.Sprintf("ToolKind(%d)", k)
}

// MarshalText implements TextUnmarshaler (for e.g. JSON encoding). For the
// method to be found by the JSON encoder, use a value receiver.
func (k ToolKind) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

// UnmarshalText implements TextMarshaler (for e.g. JSON decoding).
func (k *ToolKind) UnmarshalText(text []byte) error {
	if idx := slices.Index(toolKindStr, string(text)); idx != -1 {
		*k = ToolKind(idx)
		return nil
	}
	return errToolKind
}

// Tools represents the full configuration from tools.toml.
type Tools struct {
	Tools []Tool `toml:"tool"`
}

// Tool represents a [[tool]] section.
type Tool struct {
	Name        string      `toml:"name"`
	Description string      `toml:"description"`
	Kind        ToolKind    `toml:"kind"`
	Enabled     bool        `toml:"enabled"`
	Script      string      `toml:"script"`
	Parameters  []Parameter `toml:"parameter"`
}

// ParamType represents the type of a parameter (e.g. string, integer, boolean).
type ParamType int

const (
	String ParamType = iota
	Integer
	Number
	Boolean
)

var (
	paramTypeStr = []string{
		String:  "string",
		Integer: "integer",
		Number:  "number",
		Boolean: "boolean",
	}
	errParamType = errors.New("invalid parameter type identifier")
)

// String implements interface Stringer.
func (t ParamType) String() string {
	if int(t) >= 0 && int(t) < len(paramTypeStr) {
		return paramTypeStr[t]
	}
	return fmt.Sprintf("ParamType(%d)", t)
}

// MarshalText implements TextUnmarshaler (for e.g. JSON encoding). For the
// method to be found by the JSON encoder, use a value receiver.
func (t ParamType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements TextMarshaler (for e.g. JSON decoding).
func (t *ParamType) UnmarshalText(text []byte) error {
	if idx := slices.Index(paramTypeStr, string(text)); idx != -1 {
		*t = ParamType(idx)
		return nil
	}
	return errParamType
}

// Parameter represents a [[tool.parameter]] section.
type Parameter struct {
	Name        string    `toml:"name"`
	Description string    `toml:"description"`
	Type        ParamType `toml:"type"`
}

// LoadTools reads and parses tools.toml into a Tools struct.
func LoadTools(filePath string) (*Tools, error) {
	var config Tools
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
