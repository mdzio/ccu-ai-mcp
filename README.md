# CCU-AI-MCP

_CCU-AI-MCP_ ist eine Implementierung des [Model Context Protocols](https://de.wikipedia.org/wiki/Model_Context_Protocol) (MCP) für die Smart-Home Zentrale [OpenCCU](https://openccu.de). Das Model Context Protocol dient zum Datenaustausch zwischen einer künstlichen Intelligenz (KI), insbesondere großen Sprachmodellen (LLM), und externen Systemen. OpenCCU ist ein freies und Open-Source-basiertes Betriebssystem für eine homematic IP© CCU-Zentrale. 

## Herausstellungsmerkmale

Folgende Funktionalitäten sind bei diesem MCP-Server für die CCU besonders hervorzuheben:
* Einfache Erweiterbarkeit durch Anwender
* Mehr Werkzeuge/Tools für das _Verständnis_ und das _Anpassen_ der Konfiguration und die Programmierung der CCU

Gegenüber anderen MCP-Servern, die einen festen Satz an Tools bereitstellen, können diese beim _CCU-AI-MCP_ einfach über die **Werkzeugkonfigurationsdatei `tools.toml` angepasst** und erweitert werden. An der _CCU_ selbst müssen keinerlei Einstellungen oder Anpassungen vorgenommen werden. Der Aufbau der Konfigurationsdatei ist relativ einfach und die Tools bestehen aus **HM-Skripten**, die zwar im _CCU-AI-MCP_ konfiguriert werden, aber **auf** der CCU ausgeführt werden. Besitzer einer CCU sind meistens schon in Kontakt mit HM-Skripten gekommen, sodass keine neue Programmiersprache erlernt werden muss. Zudem ermöglichen HM-Skripte den Zugriff auf die gesamte Projektierung und Konfiguration der CCU. Alle Skriptausgaben mit der Funktion `WriteLine()` werden automatisch an das LLM weitergeleitet.

Ein besonderer Schwerpunkt wird auch auf das **Auslesen** und die **Anpassung** der **CCU-Projektierung** gelegt. Dadurch kann die KI die aktuelle Projektierung zwecks Überprüfung noch einmal detailliert erklären, Inkonsistenzen suchen oder auch Bennungen verbessern. Schnell kann man sich ein Überblick über eine fremde CCU (z.B. von einem Freund) verschaffen und nach Projektierungsfehlern suchen.

Das ist eigentlich nicht erwähnenswert, aber natürlich wird auch das Lesen von Sensorwerten und das Ansteuern von Aktoren unterstützt.

## Bespiele für LLM-Prompts

* Wie ist die Ursache-Wirkungs-Kette vom Helligkeitssensor außen bis zum Rollladen im Wohnzimmer an der Terrasse?
* Untersuche alle konfigurierbaren Namen und Beschreibungen in der CCU hinsichtlich Verständlichkeit, Funktionsbezug und Rechtschreibung.
* Korrigiere diese!

## Werkzeuge/Tools

Folgende Werkzeuge/Tools werden vom _CCU-AI-MCP_ für die KI bereit gestellt:
* `list_programs` – Listet alle CCU-Programme mit Eigenschaften und letzter Ausführung
* `update_program` – Aktualisiert Eigenschaften eines CCU-Programms (z. B. Name, Beschreibung, aktiv, sichtbar, bedienbar)
* `execute_program` – Führt ein CCU-Programm aus
* `list_system_variables` – Listet alle Systemvariablen mit Eigenschaften inkl. Referenzen zu Programmen
* `update_system_variable` – Aktualisiert Eigenschaften einer CCU-Systemvariablen (z. B. Name, Beschreibung)
* `list_devices` – Listet alle Geräte mit Typ und Adresse
* `update_device` – Aktualisiert Eigenschaften eines Geräts (z. B. Name)
* `list_channels_of_device` – Listet alle Kanäle eines Geräts inkl. Referenzen zu Programmen
* `update_channel` – Aktualisiert Eigenschaften eines Gerätekanals (z. B. Name)
* `list_data_points_of_channel` – Listet alle Datenpunkte eines Kanals
* `read_data_points` – Liest Werte mehrerer Datenpunkte
* `write_data_point` – Schreibt einen Wert in einen Datenpunkt
* `list_rooms` – Listet alle konfigurierten Räume
* `update_room` – Aktualisiert Eigenschaften eines Raums (Name, Beschreibung, Kanäle hinzufügen/entfernen)
* `list_functions` – Listet alle Gewerke
* `update_function` – Aktualisiert Eigenschaften eines Gewerks (Name, Beschreibung, Kanäle hinzufügen/entfernen)
* `read_service_messages` – Liest aktive Servicemeldungen (z. B. Batterie leer)
* `read_alarm_messages` – Liest aktive Alarmsystemvariablen
* `read_system_info` – Gibt Systeminformationen der CCU zurück, u.a. Firmware-Version, Dateisystem, Uptime und RAM-Nutzung
* `execute_script` – Führt ein beliebiges HomeMatic-Skript auf der CCU aus

## Projektstatus

Der _CCU-AI-MCP_ besitzt bereits die vollständige Basisfunktionalität und Werkzeuge können auch, wie oben beschrieben, beliebig angepasst und erweitert werden. Regelmäßig wird der _CCU-AI-MCP_ zusammen mit produktiven CCU's eingesetzt. Für die Installation und Konfiguration sind allerdings manuelle Eingriffe nötig. Insbesondere dies soll in zukünftigen Versionen vereinfacht werden. 

Hier ist noch eine Liste von Ideen für zukünftige Erweiterungen:
* Ergänzung weiterer HM-Skripte (Hierbei ist Hilfe sehr willkommen.)
  * Erstellung von Systemvariablen, Räumen und Gewerken
  * Detaillierte Analyse inkl. Zeitgeber von Wenn/Dann-Programmen
  * Erstellung von Wenn/Dann-Programmen
* Installation als Add-On auf der CCU
* Installation als Docker-Container
* Anbindung des CCU-Historians
* Für das LLM durchsuchbare Dokumentation aller HM-Geräte

## Funktionsweise

_CCU-AI-MCP_ ist ein MCP-Server, der von KI-Agenten oder Dialogschnittstellen (Conversational-UI) angesteuert wird. Der KI-Agent ist Vermittler zwischen dem LLM und dem _CCU-AI-MCP_. Der Ablauf der Kommunikation zwischen den Komponenten ist wie folgt:
1. Der Benutzer gibt eine Anfrage (Prompt) in die Dialogschnittstelle ein. Diese wird an das LLM inklusive einer Auflisten von möglichen Werkzeugen an das LLM weitergeleitet.
2. Durch das LLM wird eine Antwort generiert. In der Antwort kann das LLM mitteilen, dass es ein Werkzeug (Tool) benutzen möchte.
3. Der KI-Agent ruft daraufhin das entsprechende Tool im _CCU-AI-MCP_ auf und leitet die Tool-Ausgabe zurück an das LLM.
4. Das LLM erstellt die finale Antwort aus der ursprünglichen Anfrage und der Tool-Ausgabe.

Die Schritte 2 und 3 sind in der Regel für den Benutzer nicht sichtbar und können bei Bedarf auch mehrmals erfolgen.

## Sicherheit

LLMs können sich fehlerhaft verhalten oder sehr kreativ werden, um eine gestellte Aufgabe doch noch zu lösen. Beim _CCU-AI-MCP_ kann das LLM nur die HM-Skripte ausführen, die in der Werkzeugkonfiguration hinterlegt sind. Zudem können einzelne Werkzeuge in der Werkzeugkonfiguration deaktiviert werden.

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
logLevel = 'INFO'

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
This MCP server is used for communication with a Homematic IP CCU or an openCCU. 
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
description = 'Lists all programs of the CCU. Active and visible flags and last execution times are included.'
kind = 'hm-script'
enabled = true
script = '''
! Enumerating programs
object eobj = dom.GetObject(ID_PROGRAMS);
if (eobj) {
	WriteLine("Result is a markdown table:");
	WriteLine("| ISEID | Name | Description | Active | Visible | Last execution time |");
	WriteLine("|-------|------|-------------|--------|---------|---------------------|");
	string id;
	integer count = 0;
	foreach (id, eobj.EnumUsedIDs()) {
		object obj = dom.GetObject(id);
		WriteLine("| " # obj.ID() # " | " # obj.Name() # " | " # obj.PrgInfo() # " | " # obj.Active() # " | " #
            obj.Visible() # " | " # obj.ProgramLastExecuteTime() # " |");
		count = count + 1;
	}
	WriteLine("\nFound " # count # " programs.");
} else {
	WriteLine("ERROR: Object with ISEID ID_PROGRAMS not found.");
}
'''
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
optional | boolean | Gibt an, ob der Parameter optional ist (`true`) oder Pflicht ist (`false`). Standardmäßig ist ein Parameter Pflicht (Default: `false`).

### Tipps

* Alle Skriptpfade sollten eine (Fehler-)Meldung an das LLM zurückgeben.
* Eine leere Skriptrückgabe (z. B. durch ungültige HM-Skripte) oder eine Skriptrückgabe, die mit `ERROR:` beginnt, wird als Fehler an das LLM gemeldet.
* In der Werkzeugbeschreibung `description` sollte erwähnt werden, welche Informationen das LLM als Rückgabe erwarten kann.
* Tabellen sollten mit Markdown formatiert sein und in der ersten Zeile Spaltenüberschriften besitzen. Falls die Tabelle keine Zeilen besitzt, sollte eine Meldung ausgegeben werden, dass keine Einträge vorhanden sind. Als Vorlage kann das Werkzeug `list_programs` genommen werden.

## Einbindung in KI-Agenten

Wie die verschiedenen KI-Agenten für MCP konfiguriert werden, ist der Dokumentation des jeweiligen KI-Agenten zu entnehmen. Beispiele (z.B. für _Mistral Vibe_) sind [in diesem Dokument](doc/configure-agents.md) zu finden.

## Lizenz und Haftungsausschluss

Dieses Projekt steht unter der [GNU General Public License V3](LICENSE.txt).

## Autoren

* [MDZIO](https://github.com/mdzio)
