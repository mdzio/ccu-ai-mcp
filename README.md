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
* Analysiere alle CCU-Programme und mache einen Vorschlag je Programm für eine kurze Beschreibung.
* Übertrage diese Beschreibungen auf die CCU-Programme.

## Werkzeuge/Tools

Folgende Werkzeuge/Tools werden vom _CCU-AI-MCP_ für die KI bereit gestellt:
* `list_programs` – Listet alle CCU-Programme mit Eigenschaften und letzter Ausführung
* `read_program` – Liest alle Eigenschaften, Regeln, Bedingungen, Zeitgebener und Aktionen eines CCU-Programms
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
  * Erstellung von Wenn/Dann-Programmen
* Installation als Docker-Container
* Anbindung des CCU-Historians
* Für das LLM durchsuchbare Dokumentation aller HM-Geräte

## Handbuch

Ein ausführliches Handbuch zu Installation, Konfiguration, Start, Erstellung von Werkzeugen und Einbindung in KI-Agenten ist im [Wiki](https://github.com/mdzio/ccu-ai-mcp/wiki) zu finden.

## Lizenz und Haftungsausschluss

Dieses Projekt steht unter der [GNU General Public License V3](LICENSE.txt). Abgeleitete Werke dürfen nicht unter einer permissiveren Lizenz (z.B. MIT) veröffentlicht werden. Insbesondere darf der Quelltext nicht als Eingabe (z.B. Prompt, Trainingsdaten) für eine KI (z.B. LLM) verwendet werden, wenn die Ziellizenz nicht mit der GPLv3 kompatibel ist.

## Autoren

* [MDZIO](https://github.com/mdzio)
