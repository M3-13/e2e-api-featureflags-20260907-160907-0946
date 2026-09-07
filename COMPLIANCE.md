VERDICT: APPROVED

## Zusammenfassung

Der Feature-Flag-Service entspricht im geprüften Stand den geplanten Datenschutz- und Sicherheitskriterien. Die relevanten Kriterien `AC-12` bis `AC-18` sind im Code umgesetzt. Es wurden keine verletzten Kriterien festgestellt, die einen Hotfix oder eine Änderungswelle erfordern. Es verbleiben nicht blockierende Hinweise für die nächste Planungsrunde.

## 1. DSGVO / GDPR

### Erfüllte Kriterien

- **AC-16** — Die Logging-Middleware protokolliert ausschließlich Methode, Pfad ohne Query-String, Statuscode und Dauer. Der `user`-Query-Parameter und Request-Bodies werden nicht geloggt.  
  Beleg: `middleware.go`, `withLogging`; `sanitize(r.Method)`, `sanitize(r.URL.Path)`, `status`, `duration`. Der Query-String ist nicht Bestandteil von `r.URL.Path`.  
  Bewertung: erfüllt.

- **AC-17** — Der `user`-Parameter aus `GET /flags/{key}/evaluate` wird nicht im Store gespeichert, sondern ausschließlich für die Hash-Berechnung verwendet.  
  Beleg: `handlers_evaluate.go`, `evaluateFlagHandler`; `evaluate` nutzt `user` nur im FNV-Hash. `store.go` speichert ausschließlich `Feature{Key, Enabled, Description, RolloutPercent}`.  
  Bewertung: erfüllt.

- **AC-18** — Beim Anlegen und Ändern werden nur `key`, `enabled`, `description` und `rollout_percent` übernommen; unbekannte JSON-Felder werden verworfen.  
  Beleg: `handlers_create.go` (`createFlagRequest` mit JSON-Tags), `handlers_mutate.go` (`updateFlagRequest` ohne `key`), `response.go` / `decodeJSON` (Standard-Decoder ignoriert unbekannte Felder).  
  Bewertung: erfüllt.

### Notes (nicht blockierend, keine Kriterienverletzung)

- **Freitext `description`**  
  Das Feld `description` kann potenziell personenbezogene Daten aufnehmen. Die Spec erlaubt es; dennoch sollte in einer Betriebsrichtlinie oder einer künftigen Validierung klargestellt werden, dass `description` keine personenbezogenen Daten enthalten darf.  
  Konkrete Empfehlung: zusätzliche Längen-/Zeichenvalidierung in `handlers_create.go` und `handlers_mutate.go` oder dokumentierte Konvention in `README.md`.

- **Hash-Verfahren des `user`-Parameters**  
  `handlers_evaluate.go` verwendet FNV-1a (`hash/fnv`). FNV ist kein kryptografischer Hash und kann bei geringer Entropie der User-IDs Rückschlüsse auf das personenbezogene Datum erlauben.  
  Konkrete Empfehlung: für künftige Iterationen HMAC (z. B. `crypto/hmac` mit SHA-256 und geheimem Schlüssel) verwenden, um eine echte Pseudonymisierung zu erreichen. Das aktuelle Verhalten verletzt `AC-17` nicht.

- **Dokumentation der Rechtsgrundlage**  
  Die API verarbeitet den `user`-Parameter transient. Im Repo ist keine Rechtsgrundlage nach Art. 6 DSGVO dokumentiert. Da der Code selbst keine Speicherung vornimmt, ist dies kein Blocker, sollte aber vom Verantwortlichen im Betrieb festgelegt und in `README.md` oder einer Datenschutz-Dokumentation nachgetragen werden.

## 2. EU Cyber Resilience Act (CRA)

### Erfüllte Kriterien

- **AC-12** — Request-Body-Limit für `POST /flags` und `PUT /flags/{key}`.  
  Beleg: `response.go`, `decodeJSON` mit `http.MaxBytesReader(nil, r.Body, maxBodyBytes)`; `maxBodyBytes = 1 << 20` (1 MiB).  
  Bewertung: erfüllt.

- **AC-13** — Entfernung von Steuerzeichen in Logs.  
  Beleg: `middleware.go`, `sanitize` entfernt alle `unicode.IsControl`-Zeichen aus Methode und Pfad.  
  Bewertung: erfüllt.

- **AC-14** — Explizite Server-Timeouts.  
  Beleg: `main.go`, `http.Server` mit `ReadHeaderTimeout: 5s`, `ReadTimeout: 10s`, `WriteTimeout: 10s`, `IdleTimeout: 60s`.  
  Bewertung: erfüllt.

- **AC-15** — Keine permissiven CORS-Header.  
  Beleg: Code setzt keine `Access-Control-Allow-*`-Header; Test `TestNoCORSHeaders` verifiziert dies.  
  Bewertung: erfüllt.

### Notes (nicht blockierend, außerhalb des aktuellen Spec)

- **TLS-Terminierung / sichere Kommunikation**  
  Der Server lauscht in `main.go` auf `":" + port` ohne TLS. Das war nicht Bestandteil der AC, ist aber für einen produktiven Netzbetrieb und eine CRA-konforme Sicherheitsbetrachtung relevant.  
  Konkrete Empfehlung: TLS-Terminierung vorsehen oder dokumentiert durch ein vorgeschaltetes Gateway sicherstellen.

- **Authentifizierung / Rate-Limiting**  
  Die API ist ohne Authentifizierung und ohne Rate-Limiting umgesetzt. Dies ist kein Verstoß gegen die vorhandenen AC, aber ein sicherheitsrelevantes Thema für den späteren Betrieb.  
  Konkrete Empfehlung: in der nächsten Planungsrunde Kriterien für AuthN/AuthZ und Rate-Limiting ergänzen.

- **SBOM / Update- und Patch-Fähigkeit**  
  Das Repo hat keine externen Abhängigkeiten (`go.mod` mit drei Zeilen), was die Lieferkette vereinfacht. Eine SBOM und ein dokumentierter Update-/Patch-Prozess fehlen jedoch.  
  Konkrete Empfehlung: SBOM erzeugen und in `README.md` oder einem separaten Sicherheitsdokument dokumentieren; Deployment-/Update-Weg beschreiben.

## 3. EU AI Act

Nicht anwendbar: Der Service enthält kein KI-Feature im Sinne des AI Act.

## 4. Pflichttexte & UI-Pflichten

Nicht anwendbar: Reines Go-Backend ohne End-User-UI, ohne Cookies und ohne öffentlichen Webauftritt. Impressums-, Cookie- und Consent-Pflichten bestehen für dieses Artefakt nicht. Eine Datenschutzerklärung ist Sache des Betreibers und würde ggf. auf einem zugehörigen Webauftritt bereitgestellt.

## 5. Barrierefreiheit (WCAG/BITV/EAA)

Nicht anwendbar: Keine öffentliche Weboberfläche; die API hat keine HTML-Nutzeroberfläche.

## Findings

Keine blockierenden oder änderungspflichtigen Befunde.