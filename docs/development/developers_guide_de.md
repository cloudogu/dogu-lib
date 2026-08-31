Entwicklerhandbuch — DCC-Client (Deutsch)

Übersicht

Dieses Handbuch beschreibt die Nutzung und Erweiterung des DCC-HTTP-Clients (doguv3/dcc). Es behandelt das DccClient-Interface, Konfigurationsoptionen, Fehlerbehandlung über clienterrors sowie Empfehlungen für Integration und Tests.

Komponenten

- DccClient (Interface)
  - GetLatest(ctx, namespace, name) -> *doguv3.Dogu
  - Get(ctx, identifier) -> *doguv3.Dogu
  - GetVersions(ctx, namespace, name) -> []string
  - GetAll(ctx) -> []doguv3.Identifier

- Implementierung: httpDccClient
  - Erzeugung über NewHttpDccClient(config *config.DccClientConfiguration, creds *config.Credentials)
  - Unterstützt Context-Abbruch und Timeouts
  - Optionales Caching (otter)

Konfiguration

DccClientConfiguration (config.DccClientConfiguration):
- DccApiBaseURL (string) – Basis-URL der entfernten DCC-API. Erforderlich.
- ProxySettings (struct) – Enabled, Server, Port, Username, Password. Wenn aktiviert, wird ein Proxy in http.Transport konfiguriert.
- Timeout (int64) – Anforderungs-Timeout in Sekunden (Standard 10s bei 0)
- UseCache (bool) – Aktiviert In-Memory-Cache
- CacheExpirySeconds (int64) – TTL für Cache-Einträge
- CacheMaximumDogus (int) – Maximale Einträge im Cache

Credentials

- Credentials{Username, Password} werden als HTTP Basic Auth gesetzt.
- Keine Credentials im Quellcode ablegen; lieber Environment-Variablen oder Secret-Manager nutzen.

Fehlerbehandlung (clienterrors)

Der Client verwendet typisierte Fehler, damit der Aufrufer gezielt reagieren kann:
- clienterrors.IsNotFoundError(err) – 404
- clienterrors.IsUnauthorizedError(err) – 401
- clienterrors.IsForbiddenError(err) – 403
- clienterrors.IsConnectionError(err) – Verbindungsfehler oder 500
- clienterrors.IsGenericError(err) – andere Fehler

Verwende diese Helferfunktionen anstelle von String-Matches. Beispiel:

    dogu, err := client.Get(ctx, ident)
    if err != nil {
        switch {
        case clienterrors.IsNotFoundError(err):
            // behandeln
        case clienterrors.IsUnauthorizedError(err):
            // neu authentifizieren
        default:
            // generischer Fehler
        }
    }

HTTP & Sicherheit

- Context: alle öffentlichen Methoden akzeptieren context.Context; HTTP-Anfragen verwenden NewRequestWithContext.
- Timeouts: Timeout konfigurierbar; Standard 10s.
- TLS: TLS-Verhalten wird über die Transport-Einstellungen konfiguriert. Zertifikatsprüfung in Produktion nicht deaktivieren; stattdessen Root-CAs oder Zertifikat-Pinning nutzen.
- Logging: LoggingRoundTripper entfernt User-Info und Query-Parameter aus der URL; Authorization-Header dürfen nicht geloggt werden.

Performance & Robustheit

- Antwortkörper-Leselimits: Antworten werden auf einen sinnvollen Maximalwert (8MB) begrenzt, um OOM zu vermeiden. Für größere Nutzlasten sollten Streaming/Decoder mit LimitReader verwendet werden.
- Cache: otter wird mit einem Zugriffsbasierten Ablauf verwendet. Tests sollten keine fragilen Zeit- oder Hit-Zähl-Annahmen treffen; lieber mocken oder Fake-Clock verwenden.

Datentypen — DoguIdentifier und DoguSpec

- DoguIdentifier (doguv3.Identifier)
  - Felder: DoguNamespace (string), Name (string), Version (string).
  - Verwendung: Identifiziert eine bestimmte Dogu-Version im Format <namespace>/<name>:<version>.
  - Validierung: Namespace und Name müssen dem regulären Ausdruck ^[a-z0-9_\-]+$ entsprechen (kleingeschriebene Buchstaben, Ziffern, Unterstrich, Bindestrich). Version muss SemVer sein (validiert mit der semver-Bibliothek).
  - Beispiel: official/redmine:0.0.1

- DoguSpec / Dogu (doguv3.Dogu)
  - Die Dogu-Struktur repräsentiert den im Registry veröffentlichten Dogu-Descriptor.
  - Wichtige Felder:
    - DoguNamespace: logischer Namespace (z. B. "official").
    - Name: Dogu-Name (DNS-kompatibel, meist klein).
    - Version: Dogu-/Helm-Chart-Version (SemVer).
    - AppVersion: Anwendungs-spezifische Version (freies Format).
    - PublishedAt: Veröffentlichungszeitpunkt (Timestamp).
    - DisplayName, Description, Categories, Tags: UI-Metadaten.
    - Logo, URL, Chart: Externe Referenzen (Icon-URL, Homepage, Helm-Chart-Referenz).
    - Applications: gebündelte Anwendungen mit Versionen (ApplicationVersion{Name, Version}).
    - Images: Liste vollständiger Container-Image-Referenzen.
    - DoguApis, ServiceAccounts, ExposedPorts, ConfigurableKeys, Upgrades: Integrations- und Fähigkeitsmetadaten.
  - JSON-Mapping: Felder verwenden PascalCase json-Tags (z. B. "DoguNamespace", "Name", "Version").
  - Beispiel (Auszug):

```json
{
  "Name": "redmine",
  "DoguNamespace": "official",
  "Version": "0.0.1",
  "AppVersion": "6.1.2",
  "PublishedAt": "2026-05-06T09:57:04.927Z",
  "DisplayName": "Redmine",
  "Description": "Project management and issue tracking",
  "Categories": ["Development Apps"],
  "Tags": [
    "pm",
    "projectmanagement",
    "issue",
    "task"
  ],
  "Logo": "https://cloudogu.com/images/dogus/redmine.png",
  "URL": "https://cloudogu.com/ecosystem",
  "Chart": "oci://registry.cloudogu.com/official/dogu/v3/charts/redmine",
  "Applications": [
    {
      "Name": "redmine",
      "Version": "6.1.2"
    },
    {
      "Name": "postgresql",
      "Version": "16.8"
    }
  ],
  "Images": [
    "registry.cloudogu.com/official/dogu/v3/images/redmine:6.1.2-45.7.0",
    "docker.io/postgres:16.8"
  ],
  "DoguApis": [
    "ServiceAccountRequest.k8s.cloudogu.com/v1",
    "ServiceAccountProducer.k8s.cloudogu.com/v1",
    "Exposition.k8s.cloudogu.com/v1",
    "ConfigValidation.dogu-validation.cloudogu.com/v1",
    "UpgradePath.dogu-migration.cloudogu.com/v1"
  ],
  "ServiceAccounts": {
    "Requests": [
      { "Type": "nexus", "Optional": true }
    ],
    "Producers": [
      { "Type": "redmine" }
    ]
  },
  "ExposedPorts": [
    { "Protocol": "tcp", "Port": 3000 }
  ],
  "ConfigurableKeys": [
    "logging/root"
  ],
  "Upgrades": [
    { "From": ">=44.0.0 <45.0.0", "To": "45.7.0", "IsMigration": true }
  ]
}
```

Testing

- Unit-Tests mit httptest.Server; prüfe Mapping von Statuscodes zu clienterrors.
- Vermeide time.Sleep; setze kurze TTLs oder nutze Fake-Clock.

Beispiel

Client erzeugen und Dogu abrufen:

    cfg := &config.DccClientConfiguration{DccApiBaseURL: "https://dcc.example/api/v3/dogus", Timeout: 15}
    creds := &config.Credentials{Username: "admin", Password: "secret"}
    client, err := NewHttpDccClient(cfg, creds)

    dogu, err := client.GetLatest(context.Background(), "official", "redmine")

Bei Fehlern nutze clienterrors-Helfer zum Verzweigen.

Konventionen

- Keine Credentials ins VCS einchecken.
- Go vet und staticcheck laufen lassen; neue Features mit Unit-Tests versehen.
- Sicherheitsrelevante Änderungen (TLS, Auth, Logging) mit kurzer Risikoabschätzung im PR dokumentieren.

Ende des Handbuchs.
