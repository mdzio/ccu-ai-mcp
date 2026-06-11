package ccu

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"text/template"

	"golang.org/x/text/encoding/charmap"
)

const (
	// max. size of a valid response, if not specified: 10 MB
	// (max. size of a single response line is always 64 KB)
	scriptRespLimit = 10 * 1024 * 1024
)

var scriptLog = slog.With("component", "script-client")

// ScriptClient executes HM scripts remotely on the CCU.
type ScriptClient struct {
	// IP address or network name of the CCU
	Addr string

	// If access is local, the proxy server can be bypassed.
	UseInternalPort bool

	// Limits the size of a valid response
	RespLimit int64

	// User name for HTTP basic authentication
	UserName string

	// Password for HTTP basic authentication
	Password string
}

// Execute remotely executes a HM script on the CCU.
func (sc *ScriptClient) Execute(script string) ([]string, error) {
	scriptLog.Debug("Executing HM script", "script", script)

	// encode request body with ISO8859-1
	var reqBuf bytes.Buffer
	reqWriter := charmap.ISO8859_1.NewEncoder().Writer(&reqBuf)
	reqWriter.Write([]byte(script))

	// build address
	var port string
	if sc.UseInternalPort {
		port = "8183"
	} else {
		port = "8181"
	}
	addr := "http://" + sc.Addr + ":" + port + "/tclrega.exe"

	// http post
	httpReq, err := http.NewRequest(http.MethodPost, addr, bytes.NewReader(reqBuf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("HTTP request creation failed on %s: %w", addr, err)
	}
	if sc.UserName != "" || sc.Password != "" {
		httpReq.SetBasicAuth(sc.UserName, sc.Password)
	}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed on %s: %w", addr, err)
	}
	defer httpResp.Body.Close()

	// check status
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 299 {
		return nil, fmt.Errorf("HTTP request failed on %s with code: %s", addr, httpResp.Status)
	}

	// limit response size
	limit := sc.RespLimit
	if limit == 0 {
		limit = scriptRespLimit
	}
	limitReader := io.LimitReader(httpResp.Body, limit)

	// decode response body with ISO8859-1
	decReader := charmap.ISO8859_1.NewDecoder().Reader(limitReader)

	// read response and split lines
	scn := bufio.NewScanner(decReader)
	var resp []string
	for scn.Scan() {
		l := scn.Text()
		if strings.HasPrefix(l, "<xml><exec>") {
			break
		}
		resp = append(resp, l)
	}
	if scn.Err() != nil {
		return nil, fmt.Errorf("Parsing of response failed from %s: %v", addr, scn.Err())
	}
	scriptLog.Debug("HM script response", "response", strings.Join(resp, "\\n"))
	return resp, nil
}

// ExecuteTmpl executes a HM script template with the specified data remotely on the CCU.
func (sc *ScriptClient) ExecuteTmpl(tmpl *template.Template, data any) ([]string, error) {
	// fill template
	var sb strings.Builder
	err := tmpl.Execute(&sb, data)
	if err != nil {
		return nil, fmt.Errorf("Rendering of HM script template with data %v failed: %v", data, err)
	}

	// execute script
	resp, err := sc.Execute(sb.String())
	if err != nil {
		return nil, err
	}
	return resp, nil
}
