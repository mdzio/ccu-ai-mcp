# CCU-AI-MCP

_CCU-AI-MCP_ ist eine Implementierung des [Model Context Protocols](https://de.wikipedia.org/wiki/Model_Context_Protocol) (MCP) für die Smart-Home Zentrale [OpenCCU](https://openccu.de). Das Model Context Protocol dient zum Datenaustausch zwischen einer künstlichen Intelligenz (KI), insbesondere großen Sprachmodellen (LLM), und externen Systemen. OpenCCU ist ein freies und Open-Source-basiertes Betriebssystem für eine homematic IP© CCU-Zentrale. 

## Funktionsweise

_CCU-AI-MCP_ ist ein MCP-Server, der von KI-Agenten oder Dialogschnittstellen (Conversational-UI) angesteuert wird. Der KI-Agent ist Vermittler zwischen dem LLM und dem _CCU-AI-MCP_. Der Ablauf der Kommunikation zwischen den Komponenten ist wie folgt:
1. Der Benutzer gibt eine Anfrage (Prompt) in die Dialogschnittstelle ein. Diese wird an das LLM inklusive einer Auflisten von möglichen Werkzeugen an das LLM weitergeleitet.
2. Durch das LLM wird eine Antwort generiert. In der Antwort kann das LLM mitteilen, dass es ein Werkzeug (Tool) benutzen möchte.
3. Der KI-Agent ruft daraufhin das entsprechende Tool im _CCU-AI-MCP_ auf und leitet die Tool-Ausgabe zurück an das LLM.
4. Das LLM erstellt die finale Antwort aus der ursprünglichen Anfrage und der Tool-Ausgabe.

Die Schritte 2 und 3 sind in der Regel für den Benutzer nicht sichtbar und können bei Bedarf auch mehrmals erfolgen.

Gegenüber anderen MCP-Servern, die einen festen Satz an Tools bereitstellen, können diese beim _CCU-AI-MCP_ einfach über die Werkzeugkonfigurationsdatei `tools.toml` angepasst und erweitert werden. Der Aufbau der Konfigurationsdatei ist relativ einfach und die Tools bestehen aus HM-Skripten, die **auf** der CCU ausgeführt werden. Besitzer einer CCU sind meistens schon in Kontakt mit HM-Skripten gekommen, sodass keine neue Programmiersprache erlernt werden muss. Zudem ermöglichen HM-Skripte den Zugriff auf die gesamte Projektierung und Konfiguration der CCU. Alle Skriptausgaben mit der Funktion `WriteLine` werden automatisch an das LLM weitergeleitet.

## Sicherheit

LLMs können sich fehlerhaft verhalten oder sehr kreativ werden, um eine gestellte Aufgabe doch noch zu lösen. Beim _CCU-AI-MCP_ kann das LLM nur die HM-Skripte ausführen, die in der Werkzeugkonfiguration hinterlegt sind. In der Standardkonfiguration sind zudem alle HM-Skripte, die Datenpunkte, die Konfiguration oder Projektierung der CCU ändern können, deaktiviert.

## Voraussetzungen

Der _CCU-AI-MCP_ ist ein MCP-Server. Er enthält weder ein LLM (Large Language Model), eine Convertional-UI oder einen KI-Agenten. Diese müssen zusätzlich zum _CCU-AI-MCP_ bereit gestellt werden. Im Folgenden ist eine unvollständige Liste zu finden.

Generische KI-Agenten:
* OpenClaw
* Hermes Agent

Coding-Agenten:
* [Mistral Vibe for Code](https://mistral.ai/products/vibe/code/) (Anbindung siehe weiter unten)
* Claude Code

Conversational-UIs:
* [Open WebUI](https://github.com/open-webui/open-webui)

In der Regel können KI-Agenten oder Conversational-UIs lokale LLMs oder auch LLMs in der Cloud verwendet werden. Eine Möglichkeit, LLMs lokal auszuführen, ist [Ollama](https://ollama.com/), vorausgesetzt, die entsprechende Hardware ist vorhanden.

## Installation

Zur Installation vom CCU-AI-MCP muss die zur Rechnerplattform passende Datei in einem Verzeichnis entpackt werden. Pakete zur automatischen Installation existieren bisher nicht.

## Konfiguration

Die Hauptkonfigurationsdatei kann über das Befehlszeilenargument `-config` angegeben werden. Standardmäßig wird im Arbeitsverzeichnis nach der Datei `config.toml` gesucht. In der Hauptkonfigurationsdatei wird mit der Option `toolFile` eine zweite Konfigurationsdatei referenziert, in der nur die Werkzeuge spezifiziert werden.

Aufbau der Hauptkonfigurationsdatei mit Kommentaren:
```toml
# This file contains the general configuration for the CCU-AI-MCP server.

# For more information about the TOML format, see the official specification:
# https://toml.io/en/v1.1.0

[general]
# Log level for the application. Valid values: DEBUG, INFO, WARN, ERROR
logLevel = 'DEBUG'

# Path to the tools configuration file
toolFile = 'tools.toml'

[ccu]
# Specifies the IP address or hostname of the CCU.
address = 'homematic-raspi'

# If authentication for the CCU network API is active, specify the user and
# password. Otherwise, leave them empty.
user = ''
password = ''

[mcp]
# Transport type for MCP communication. Valid values: stdio, http, https
transport = 'stdio'

# Port for HTTP(S) transport
port = 2080

# Path to TLS certificate file for HTTPS transport (PEM format)
certFile = ''

# Path to TLS key file for HTTPS transport (PEM format)
keyFile = ''

# API key for MCP authentication using Bearer token. Leave empty to disable API
# key checking.
apiKey = ''

# CORS allowed origins for browser-based MCP clients
corsAllowedOrigins = ['*']

# Instructions for the MCP server
instructions = '''
This server is used for communication with a Homematic IP CCU or an openCCU. 
This is the central unit of a smart home. It allows querying sensors, controlling 
actuators, and starting automations (programs). Additionally, the CCU configuration
can be read. A commonly used parameter in tool calls is the ISEID. All objects in
the CCU have this unique ID.
'''
```

Beispielauszug aus der Konfigurationsdatei für die Werkzeuge:
```toml
[[tool]]
name = 'list_programs'
description = 'Lists all programs of the CCU. Active and visible flags are included.'
kind = 'hm-script'
enabled = true
script = '''
object eobj = dom.GetObject(ID_PROGRAMS);
if (eobj) {
	WriteLine("Result is a tab separated table with headers in the first line:");
	WriteLine("ISEID\tName\tDescription\tActive\tVisible");
	string id;
	integer count = 0;
	foreach (id, eobj.EnumIDs()) {
		object obj = dom.GetObject(id);
		WriteLine(obj.ID() # "\t" # obj.Name() # "\t" # obj.PrgInfo() # "\t" # obj.Active() # "\t" # obj.Visible());
		count = count + 1;
	}
	WriteLine("\nFound " # count # " programs.");
} else {
	WriteLine("ERROR: Object with ISEID ID_PROGRAMS not found.");
}
```

Relative Dateipfade für das Befehlszeilenargument `-config` oder die Konfigurationsoption `toolsFile` werden auf das Arbeitsverzeichnis bezogen.

## Start

Gestartet wird _CCU-AI-MCP_ auf der Konsole mit `./ccu-ai-mcp` (Linux) bzw. `ccu-ai-map.exe` (Windows). Optional kann über die Option `-config` der Pfad zur Hauptkonfigurationsdatei angegeben werden. In der Regel funktioniert auch ein Doppelklick in einem Dateimanager. Eine automatische Einrichtung als systemd- oder Windows-Dienst existiert derzeit nicht.

## Erstellung von Werkzeugen

Die Definition eines Tools erfolgt durch eine neue `[[tool]]` Sektion in der Werkzeugkonfigurationsdatei `tools.toml`.

Folgende Optionen zur Konfiguration eines Tools existieren in der `[[tool]]` Sektion:

Name        | Datentyp | Bedeutung
------------|---|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
name        | string | Der Name des Werkzeug sollte von der Funktionalität abgeleitet sein: z.B. `list_programs`. Die Schreibweise sollte [snake_case](https://en.wikipedia.org/wiki/Snake_case) sein.
description | string | Die Beschreibung sollte kurz die Funktionalität erläutern und die für das LLM in der Rückgabe erwartbaren Informationen auflisten.
kind        | string | Als Art des Tools wird bisher nur `hm-script` unterstützt.
enabled     | boolean | Hiermit können einzelne Tools aktiviert (`true`) oder deaktiviert (`false`) werden.
script      | string | Falls die Art des Tools `hm-script` ist, muss hier eine Vorlage für das auszuführende HM-Skript angegeben werden. Das HM-Skript kann Platzhalter für Parameter enthalten, die vor der Ausführung von dem LLM gesetzt werden müssen. Am besten wird ein mehrzeiliges Skript mit drei Hochkomma eingeleitet und abgeschlossen.

In der Skriptvorlage können Platzhalter für Parameter verwendet werden. Diese müssen durch doppelte geschweifte Klammern eingeschlossen werden, z.B. `{{ .iseid }}`. Jeder verwendete Parameter muss durch eine `[[tool.parameter]]` Sektion definiert werden. 

Folgende Optionen zur Konfiguration eines Parameters existieren in der `[[tool.parameter]]` Sektion:

Name | Datentyp | Bedeutung
---|---|---
name | string | Ein kurzer Bezeichner (ohne Leerzeichen) für den Parameter.
description | string | Aus der Beschreibung des Parameters muss genau hervorgehen, wie das LLM diesen zu füllen hat.
type | string | Der Datentyp des Parameters. Folgende Datentypen werden derzeit unterstützt: `string`, `integer`, `number`, `boolean`, `any`, `string[]`, `integer[]`, `number[]` und `boolean[]`.

### Tipps

* Alle Skriptpfade sollten eine (Fehler-)Meldung an das LLM zurückgeben.
* Eine leere Skriptrückgabe (z. B. durch ungültige HM-Skripte) oder eine Skriptrückgabe, die mit `ERROR:` beginnt, wird als Fehler an das LLM gemeldet.
* In der Werkzeugbeschreibung `description` sollte erwähnt werden, welche Informationen das LLM als Rückgabe erwarten kann.
* Tabellen sollten tabsepariert (TSV) sein und in der ersten Zeile Spaltenüberschriften besitzen. Falls die Tabelle keine Zeilen besitzt, sollte eine Meldung ausgegeben werden, dass keine Einträge vorhanden sind. Als Vorlage kann das Werkzeug `list_programs` genommen werden.

## Einbindung in KI-Agenten

Wie die verschiedenen KI-Agenten für MCP konfiguriert werden, ist der Dokumentation des jeweiligen KI-Agenten zu entnehmen. Im Folgenden wird als Beispiel die Konfiguration von _Mistral Vibe_ gezeigt.

### Mistral Vibe

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

#### Tipps

* Eine eventuell bereits vorhandene Konfigurationsoption `mcp_servers = []` muss entfernt werden.
* Beim STDIO-Transport werden Log-Meldungen von _CCU-AI-MCP_ in die Log-Datei von _vibe_ übernommen, wenn die Umgebungsvariable `LOG_LEVEL=DEBUG` gesetzt wird. Beispielaufruf: `LOG_LEVEL=DEBUG vibe` (Linux). Der Pfad zur _vibe_ Log-Datei ist `~/.vibe/logs/vibe.log` (Linux).
* Für einen besseren Datenschutz sind folgende Konfigurationsoptionen zu setzen:
  * `enable_telemetry = false`
  * `include_commit_signature = false` (bei Cloud-Modellen)
  * `enable_auto_update = false` (optional)
* Damit _vibe_ im Betriebssystem (selbst) installierte CA-Zertifikate berücksichtigt, ist die Option`enable_system_trust_store = true` zu setzen.
## Entwicklerdokumentation

* [Mistral Vibe Anwenderdokumentation](https://docs.mistral.ai/mistral-vibe/overview)
* [Mistral Vibe MCP-Integration bei DeepWiki](https://deepwiki.com/mistralai/mistral-vibe/3.5-mcp-integration)
* [Vibe Config Guide von chris-hatton](https://gist.github.com/chris-hatton/6e1a62be8412473633f7ef02d067547d)
* [MCP-GO](https://mcp-go.dev/getting-started)

## Lizenz und Haftungsausschluss

Dieses Projekt steht unter der [GNU General Public License V3](LICENSE.txt).

## Autoren

* [MDZIO](https://github.com/mdzio)
