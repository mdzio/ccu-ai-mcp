package config

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/BurntSushi/toml"
)

// Transport represents the transport for MCP.
type Transport int

const (
	STDIO Transport = iota
	HTTP
	HTTPS
)

var (
	transportStr = []string{
		STDIO:  "stdio",
		HTTP:   "http",
		HTTPS:  "https",
	}
	errTransport = errors.New("invalid transport identifier")
)

// String implements interface Stringer.
func (t Transport) String() string {
	if int(t) >= 0 && int(t) < len(transportStr) {
		return transportStr[t]
	}
	return fmt.Sprintf("Transport(%d)", t)
}

// MarshalText implements TextUnmarshaler (for e.g. JSON encoding). For the
// method to be found by the JSON encoder, use a value receiver.
func (t Transport) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements TextMarshaler (for e.g. JSON decoding).
func (t *Transport) UnmarshalText(text []byte) error {
	if idx := slices.Index(transportStr, string(text)); idx != -1 {
		*t = Transport(idx)
		return nil
	}
	return errTransport
}

// Main represents the full configuration from config.toml.
type Main struct {
	General General `toml:"general"`
	CCU     CCU     `toml:"ccu"`
	MCP     MCP     `toml:"mcp"`
}

// General represents the [general] section.
type General struct {
	LogLevel      slog.Level `toml:"logLevel"`
	ToolFile string `toml:"toolFile"`
}

// CCU represents the [ccu] section.
type CCU struct {
	Address  string `toml:"address"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

// MCP represents the [mcp] section.
type MCP struct {
	Transport          Transport `toml:"transport"`
	Port               int       `toml:"port"`
	CertFile           string    `toml:"certFile"`
	KeyFile            string    `toml:"keyFile"`
	CORSAllowedOrigins []string  `toml:"corsAllowedOrigins"`
	Instructions       string    `toml:"instructions"`
	APIKey             string    `toml:"apiKey"`
}

// Load reads and parses config.toml into a Main struct.
func Load(filePath string) (*Main, error) {
	var config Main
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
