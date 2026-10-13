# Review der Shell-Skript-Werkzeuge

Prüfdatum: 10.10.2026  
Branch: `vibe/issue-4-shell-script-965957`  
Geprüfter Commit: `8368042` – Add shell script tools for access to the CCU operating system

## 1. Hohe Priorität: Kindprozesse können den Timeout überleben

**Fundstelle:** `tools.go:193–197`, Funktion `handleShellScript`.

`exec.CommandContext` startet `sh -c` und beendet bei Kontextabbruch standardmäßig nur den gestarteten Shell-Prozess. Von der Shell gestartete Kindprozesse können weiterlaufen. `WaitDelay` begrenzt das Warten und schließt gegebenenfalls offene Ausgabepipes, beendet aber nicht zuverlässig die gesamte Prozesshierarchie.

**Auswirkung:** Trotz gemeldetem Timeout oder Abbruch können Programme weiterhin Ressourcen beanspruchen oder Änderungen am System vornehmen. Das konfigurierte Zeitlimit von 60 Sekunden begrenzt damit nicht zuverlässig die gesamte Skriptausführung.

**Verbesserung:** Unter Linux eine eigene Prozessgruppe starten und beim Timeout bzw. Abbruch die gesamte Gruppe beenden. Betriebssystemspezifische Implementierung so kapseln, dass die vorhandenen Windows- und macOS-Builds weiter funktionieren.

**Validierung:** Durch Prüfung des Projektcodes und der lokalen Go-Implementierung von `CommandContext` bestätigt. Ein Laufzeittest auf Linux/BusyBox steht aus.

**Empfohlener Regressionstest:** Ein Skript startet einen Kindprozess, der nach einer Verzögerung eine Testdatei schreiben würde. Nach Abbruch darf weder der Kindprozess weiterlaufen noch die Datei entstehen.

## 2. Mittlere Priorität: Fehler der Log-Suche werden als Erfolg gemeldet

**Fundstelle:** `tools.toml:1112`, Werkzeug `read_ccu_log`.

Die Pipeline `grep -i -- PATTERN "$FILE" | tail -n "$LINES"` liefert ohne zusätzliche Fehlerbehandlung den Exitcode von `tail`. Ein Fehler in `grep` geht damit für die Statusauswertung verloren.

**Reproduktion:** Mit dem ungültigen regulären Ausdruck `[` meldet `grep` einen Fehler, während die Pipeline Exitcode 0 liefert. Dies wurde lokal mit Git Bash bestätigt. Der Shell-Handler setzt bei diesem Exitcode `IsError=false`, obwohl die Suche fehlgeschlagen ist.

**Verbesserung:** Den Exitstatus von `grep` separat erfassen und unterscheiden:

- Status 0: Treffer vorhanden.
- Status 1: Keine Treffer; kein Ausführungsfehler.
- Status 2: Such- oder Lesefehler; als Fehler an den MCP-Client melden.

Die Umsetzung muss mit der BusyBox-Shell der CCU funktionieren. Ein pauschales Aktivieren von `pipefail` allein unterscheidet nicht zwischen fehlenden Treffern und einem echten Fehler.

**Empfohlene Regressionstests:** Gültiges Muster mit Treffern, gültiges Muster ohne Treffer, ungültiges Muster und Lesefehler. Dabei sowohl Ausgabetext als auch `IsError` prüfen.

## 3. Verbesserung: Zeilenanzahl eindeutig behandeln und begrenzen

**Fundstelle:** `tools.toml:1111`, Parameter `lines`.

Die Bedingung `{{ if .lines }}` unterscheidet nicht zwischen einem fehlenden Parameter und dem Wert 0. Ein Aufruf mit `lines=0` verwendet deshalb den Standardwert 100. Ein zulässiger Wertebereich wird bisher nicht definiert.

**Verbesserung:** Festlegen, ob 0 zulässig ist oder als ungültige Eingabe zurückgewiesen wird. Fehlende Werte ausdrücklich von angegebenen Werten unterscheiden und einen sinnvollen Wertebereich validieren. Die Parameterbeschreibung muss dieses Verhalten erläutern.

**Empfohlene Regressionstests:** Fehlender Wert, 0, negative Werte, regulärer positiver Wert und Überschreitung der festgelegten Obergrenze.

## 4. Verbesserung: Ausführungstests erweitern

**Fundstelle:** `tools_test.go`, insbesondere `TestHandleShellScript`.

Vorhanden sind Tests für Quoting, lokale Adressen, den begrenzten Puffer, Template-Parsing sowie erfolgreiche und fehlgeschlagene Shell-Ausführung. Es fehlen gezielte Tests für:

- Timeout und expliziten Kontextabbruch.
- Verbleibende Kindprozesse nach Abbruch.
- Ausgabekürzung einschließlich Hinweistext im vollständigen Handler-Ergebnis.
- Fehlerhafte Suchmuster und fehlende Treffer in `read_ccu_log`.
- Grenzfälle des Parameters `lines`.

**Verbesserung:** Die Ausführungstests unter Linux und möglichst mit der auf der CCU verwendeten BusyBox-Shell laufen lassen. Übersprungene Tests dürfen nicht als Nachweis für funktionierende Shell-Ausführung gelten.

## Positive Befunde

- `shq` implementiert korrektes POSIX-Quoting für die geprüften Werte, einschließlich Apostrophen und Shell-Sonderzeichen.
- Der Ausgabepuffer begrenzt den gespeicherten Inhalt und kennzeichnet abgeschnittene Ausgabe.
- stdout und stderr werden gemeinsam erfasst.
- Shell-Werkzeuge werden bei einer nicht lokalen konfigurierten CCU-Adresse übersprungen.

## Durchgeführte Prüfungen und Grenzen

- `go test -v ./...`: erfolgreich; `TestHandleShellScript` wurde unter Windows übersprungen, weil `sh` nicht im Suchpfad verfügbar war.
- `go vet ./...`: erfolgreich.
- Fehlerstatus der `grep`/`tail`-Pipeline: lokal mit Git Bash reproduziert.
- Verhalten von `CommandContext` und `WaitDelay`: anhand des lokalen Go-Quellcodes geprüft.
- Kein vollständiger Lauf unter Linux/BusyBox und kein Live-Test auf einer CCU.
- Im Rahmen des Reviews wurden keine Implementierungsänderungen vorgenommen.
