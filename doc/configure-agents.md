# Einbindung in KI-Agenten

Im folgenden sind spezielle Instruktionen und Einbindungsbeispiele für Agenten zu finden.

## Instruktionen für Agenten (AGENTS.md)

Viele Agenten können für ihre wahrzunehmende Aufgabe speziell instruiert werden. Dies erfolgt durch die Datei `AGENTS.md`. Ein zugeschnittenes Beispiel ist im Ordner `agents` zu finden.

## Mistral Vibe

MCP-Server werden bei [Mistral Vibe](https://mistral.ai/products/vibe) in der Konfigurationsdatei `~/.vibe/config.toml` (Linux) bzw. `%USERPROFILE%\.vibe\config.toml` (Windows) angegeben.

Vorlage für eine lokale Anbindung über STDIO-Transport:
```toml
[[mcp_servers]]
name = "ccu_smart_home"
transport = "stdio"
command = "<INSTALLATIONSVERZ.>/ccu-ai-mcp"
cwd = "<INSTALLATIONSVERZ.>"
```

Vorlage für eine Netzwerkanbindung über HTTP-Transport:
```toml
[[mcp_servers]]
name = "ccu_smart_home"
transport = "streamable-http"
url = "http://<CCU-AI-MCP RECHNER>:2080/mcp"
```
`<CCU-AI-MCP RECHNER>` ist der Name oder die IP-Adresse des Rechners, auf dem der CCU-AI-MCP gestartet wurde.

Vorlage mit HTTPS-Transport und API-Schlüssel:
```toml
[[mcp_servers]]
name = "ccu_smart_home"
transport = "streamable-http"
url = "https://<CCU-AI-MCP RECHNER>:2080/mcp"
headers = { "Authorization" = "Bearer <API SCHLÜSSEL>" }
```
`<API SCHLÜSSEL>` muss identisch sein mit dem Wert der Konfigurationsoption `apiKey` in der `config.toml`.

### Tipps

* Eine eventuell bereits vorhandene Konfigurationsoption `mcp_servers = []` muss entfernt werden.
* Beim STDIO-Transport werden Log-Meldungen von _CCU-AI-MCP_ in die Log-Datei von _vibe_ übernommen, wenn die Umgebungsvariable `LOG_LEVEL=DEBUG` gesetzt wird. Beispielaufruf: `LOG_LEVEL=DEBUG vibe` (Linux). Der Pfad zur _vibe_ Log-Datei ist `~/.vibe/logs/vibe.log` (Linux).
* Für einen besseren Datenschutz sind folgende Konfigurationsoptionen zu setzen:
  * `enable_telemetry = false`
  * `include_commit_signature = false` (bei Cloud-Modellen)
  * `enable_auto_update = false` (optional)
* Damit _vibe_ im Betriebssystem (selbst) installierte CA-Zertifikate berücksichtigt, ist die Option`enable_system_trust_store = true` zu setzen.
