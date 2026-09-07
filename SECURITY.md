VERDICT: APPROVED

## Security-Review

### Scanner-Einordnung
Für den Projekttyp `go-backend` wurden laut Scanner-Ausgabe keine anwendbaren Scanner ausgeführt. Es liegen daher keine Scanner-Befunde vor. Die Abwesenheit von Scanner-Output ist keine Evidenz für Abwesenheit von Schwachstellen; eine zukünftige Dependency-Prüfung mit `govulncheck` wäre sinnvoll, ist aber kein Verstoß gegen ein AC-Kriterium.

### Geprüfte Bereiche

- **Secrets:** Keine hartkodierten Schlüssel, Passwörter oder Tokens im Code. Keine Secrets in Logs.
- **Injection/Eingaben:** Keine SQL-, Command- oder Pfad-Injection. JSON-Decoding mit Größenlimit, Eingabevalidierung für `key`, `enabled` und `rollout_percent`. Keine unsichere Deserialisierung.
- **AuthN/AuthZ:** Keine Authentifizierung/Autorisierung implementiert. Dies ist durch kein AC-Kriterium gefordert und daher kein blockierender Befund; siehe Notes.
- **Dependencies:** Nur Standardbibliothek `net/http` etc.; keine externen Abhängigkeiten sichtbar. Kein Scanner-Ergebnis vorhanden.
- **Configuration/Transport:** Explizite Server-Timeouts gesetzt, keine permissiven CORS-Header, HTTPS/TLS nicht konfiguriert (kein AC-Kriterium; siehe Notes).

### Erfüllte Sicherheitskriterien

- **AC-12:** `decodeJSON` nutzt `http.MaxBytesReader` mit einem Limit von `1 MiB` und wird in `POST /flags` sowie `PUT /flags/{key}` verwendet. Bei Überschreitung antworten die Handler mit einer JSON-Fehlerantwort.
- **AC-13:** `sanitize` entfernt Steuerzeichen aus Methode und Pfad. Der Status wird als Integer formatiert. Jeder Request erzeugt genau einen Log-Eintrag.
- **AC-14:** `main.go` setzt `ReadHeaderTimeout: 5s`, `ReadTimeout: 10s`, `WriteTimeout: 10s`, `IdleTimeout: 60s`.
- **AC-15:** Es werden keine `Access-Control-Allow-*`-Header gesetzt; auch der Test `TestNoCORSHeaders` bestätigt dies.
- **AC-16:** Die Logging-Middleware protokolliert ausschließlich Methode, `r.URL.Path` (ohne Query-String), Status und Dauer. `user` und Request-Bodies werden nicht gelesen oder geloggt.
- **AC-17:** `evaluateFlagHandler` verwendet `user` nur für die FNV-1a-Hash-Berechnung; der Parameter wird nicht im Store abgelegt.
- **AC-18:** Die Request-Strukturen enthalten nur `key`, `enabled`, `description`, `rollout_percent`; unbekannte JSON-Felder werden durch `encoding/json` verworfen. Beim Update wird `key` aus dem Pfad übernommen und nicht aus dem Body.

### Findings
Keine blockierenden oder mittleren Sicherheitsbefunde.

### Notes (non-blocking)

- **Fehlende Authentifizierung/Autorisierung:** Die API ist ohne Zugriffsschutz nutzbar. Dies ist durch kein AC-Kriterium abgedeckt, sollte aber eingeplant werden, falls der Dienst außerhalb eines vertrauenswürdigen Netzes betrieben wird.
- **Kein TLS:** `ListenAndServe` verwendet HTTP ohne TLS. Für Produktivbetrieb sollte TLS z.B. über einen Reverse-Proxy oder `ListenAndServeTLS` ergänzt werden.
- **Unbegrenztes Store-Wachstum:** Der In-Memory-Store kann durch viele `POST /flags`-Anfragen mit jeweils bis zu 1 MiB großen Bodies viel Speicher belegen. Ein Mengen- oder Größenschutz pro Flag ist nicht gefordert, wäre aber eine sinnvolle Härtung.
- **Startup-Log:** Der Wert von `PORT` aus der Umgebung wird ungefiltert geloggt. Dies ist nicht über einen Request steuerbar und fällt nicht unter AC-13, könnte aber bei ungewöhnlichen Umgebungsvariablen unschöne Log-Zeilen erzeugen.