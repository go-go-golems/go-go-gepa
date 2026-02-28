# Apps Browser — Final Design

Three interconnected surfaces: **Apps Folder** (icon workspace), **Module Browser** (Smalltalk inspector), and **Health Dashboard** (operational overview). All coexist as windows in the OS desktop.

---

## 1. Apps Folder — Icon View (default entry point)

```
┌─ Mounted Apps ─────────────────────────────────────────── ○ □ ┐
│  ◀ ▶   2 apps  ·  2 healthy  ·  1 required              ⟳    │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│                                                                │
│       ┌──────┐                  ┌──────┐                       │
│       │ ◈ ●  │                  │   ●  │                       │
│       │ ┌──┐ │                  │ ┌──┐ │                       │
│       │ │▦▦│ │                  │ │▦▦│ │                       │
│       │ └──┘ │                  │ └──┘ │                       │
│       │      │                  │  ★   │                       │
│       └──────┘                  └──────┘                       │
│       Inventory                  GEPA                          │
│                                                                │
│                                                                │
│                                                                │
│                                                                │
└────────────────────────────────────────────────────────────────┘

  ● healthy   ○ unhealthy   ◈ required   ★ reflective
```

---

## 2. Apps Folder — with unhealthy module

```
┌─ Mounted Apps ─────────────────────────────────────────── ○ □ ┐
│  ◀ ▶   2 apps  ·  ⚠ 1 unhealthy  ·  1 required         ⟳    │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│                                                                │
│       ┌──────┐                  ┌──────┐                       │
│       │ ◈ ○  │                  │   ●  │                       │
│       │ ┌──┐ │                  │ ┌──┐ │                       │
│       │ │░░│ │                  │ │▦▦│ │                       │
│       │ └──┘ │                  │ └──┘ │                       │
│       │      │                  │  ★   │                       │
│       └──────┘                  └──────┘                       │
│       ⚠ Inventory                GEPA                          │
│                                                                │
│                                                                │
│                                                                │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## 3. Right-click context menu — reflective module

```
┌─ Mounted Apps ─────────────────────────────────────────── ○ □ ┐
│  ◀ ▶   2 apps  ·  2 healthy  ·  1 required              ⟳    │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│                                                                │
│       ┌──────┐                  ┌──────┐                       │
│       │ ◈ ●  │                  │   ●  │                       │
│       │ ┌──┐ │                  │ ┌──┐ │                       │
│       │ │▦▦│ │                  │ │▦▦│ │                       │
│       │ └──┘ │                  │ └──┘ │                       │
│       │      │                  │  ★   │                       │
│       └──────┘                  └──────┘                       │
│       Inventory                  GEPA                          │
│                                  ┌─────────────────────────┐   │
│                                  │  Get Info           ⌘I  │   │
│                                  │  Open in Browser    ⌘B  │   │
│                                  ├─────────────────────────┤   │
│                                  │  Inspect Reflection ⌘R  │   │
│                                  ├─────────────────────────┤   │
│                                  │  Copy Base URL          │   │
│                                  │  Copy Reflection URL    │   │
│                                  ├─────────────────────────┤   │
│                                  │  Health Check       ⌘H  │   │
│                                  └─────────────────────────┘   │
└────────────────────────────────────────────────────────────────┘
```

---

## 4. Right-click context menu — non-reflective module

```
       ┌──────┐
       │ ◈ ●  │
       │ ┌──┐ │
       │ │▦▦│ │
       │ └──┘ │
       │      │
       └──────┘
       Inventory
       ┌─────────────────────────┐
       │  Get Info           ⌘I  │
       │  Open in Browser    ⌘B  │
       ├─────────────────────────┤
       │  Inspect Reflection     │
       │  (not available)   ░░░  │
       ├─────────────────────────┤
       │  Copy Base URL          │
       ├─────────────────────────┤
       │  Health Check       ⌘H  │
       └─────────────────────────┘
```

---

## 5. Get Info window (from context menu → Get Info)

```
┌─ GEPA — Get Info ──────────────────────────── ○ □ ┐
│                                                     │
│         ┌──────┐                                    │
│         │   ●  │     GEPA                           │
│         │ ┌──┐ │     gepa                           │
│         │ │▦▦│ │                                    │
│         │ └──┘ │                                    │
│         │  ★   │                                    │
│         └──────┘                                    │
│                                                     │
│  ─── General ────────────────────────────────────   │
│                                                     │
│  Description:  GEPA script runner backend module    │
│  Required:     No                                   │
│  Base URL:     /api/apps/gepa/                      │
│                                                     │
│  ─── Health ─────────────────────────────────────   │
│                                                     │
│  Status:       ● Healthy                            │
│  Last check:   14:32:07                             │
│                                                     │
│  ─── Reflection ─────────────────────────────────   │
│                                                     │
│  Available:    ★ Yes (v1)                           │
│  URL:          /api/os/apps/gepa/reflection         │
│                                                     │
│  ─── APIs (5) ───────────────────────────────────   │
│                                                     │
│  GET  /scripts       list local scripts             │
│  POST /run           execute a script               │
│  GET  /events        event stream                   │
│  GET  /schemas/{id}  schema doc                     │
│  GET  /timeline      timeline entries               │
│                                                     │
│  ─── Schemas (4) ────────────────────────────────   │
│                                                     │
│  run-request  ·  run-response  ·  script-def  ·     │
│  event-entry                                        │
│                                                     │
│              [ Open in Browser ]                    │
└─────────────────────────────────────────────────────┘
```

---

## 6. Get Info — non-reflective module

```
┌─ Inventory — Get Info ─────────────────────── ○ □ ┐
│                                                     │
│         ┌──────┐                                    │
│         │ ◈ ●  │     Inventory                      │
│         │ ┌──┐ │     inventory                      │
│         │ │▦▦│ │                                    │
│         │ └──┘ │                                    │
│         │      │                                    │
│         └──────┘                                    │
│                                                     │
│  ─── General ────────────────────────────────────   │
│                                                     │
│  Description:  Inventory chat runtime, profiles,    │
│                timeline, and confirm APIs            │
│  Required:     ◈ Yes (required at startup)          │
│  Base URL:     /api/apps/inventory/                 │
│                                                     │
│  ─── Health ─────────────────────────────────────   │
│                                                     │
│  Status:       ● Healthy                            │
│  Last check:   14:32:07                             │
│                                                     │
│  ─── Reflection ─────────────────────────────────   │
│                                                     │
│  Available:    No                                   │
│                                                     │
│  This module does not publish reflective             │
│  API metadata yet.                                   │
│                                                     │
└─────────────────────────────────────────────────────┘
```

---

## 7. Module Browser — nothing selected (opened from "Open in Browser" or menu View → as Browser)

```
┌─ Module Browser ────────────────────────────────────────────── ○ □ ┐
│                                                            ⟳       │
├──────────────────┬──────────────────┬──────────────────────────────┤
│ Modules          │ APIs             │ Schemas                      │
│──────────────────│──────────────────│──────────────────────────────│
│                  │                  │                              │
│  ● Inventory  ◈  │                  │                              │
│  ● GEPA       ★  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
├──────────────────┴──────────────────┴──────────────────────────────┤
│                                                                    │
│  Select a module to inspect.                                       │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## 8. Module Browser — GEPA selected, APIs populate

```
┌─ Module Browser ────────────────────────────────────────────── ○ □ ┐
│                                                            ⟳       │
├──────────────────┬──────────────────┬──────────────────────────────┤
│ Modules          │ APIs             │ Schemas                      │
│──────────────────│──────────────────│──────────────────────────────│
│                  │                  │                              │
│  ● Inventory  ◈  │  GET  /scripts   │  run-request                 │
│ «GEPA»        ★  │  POST /run       │  run-response                │
│                  │  GET  /events    │  script-def                  │
│                  │  GET /schemas/…  │  event-entry                 │
│                  │  GET  /timeline  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
├──────────────────┴──────────────────┴──────────────────────────────┤
│ GEPA                                               ● healthy  ★   │
│────────────────────────────────────────────────────────────────────│
│ id:          gepa                                                  │
│ description: GEPA script runner backend module                     │
│ required:    no                                                    │
│ base:        /api/apps/gepa/                                       │
│ reflection:  available (v1)                                        │
└────────────────────────────────────────────────────────────────────┘
```

---

## 9. Module Browser — GEPA → API selected, schema cross-refs highlighted

```
┌─ Module Browser ────────────────────────────────────────────── ○ □ ┐
│                                                            ⟳       │
├──────────────────┬──────────────────┬──────────────────────────────┤
│ Modules          │ APIs             │ Schemas                      │
│──────────────────│──────────────────│──────────────────────────────│
│                  │                  │                              │
│  ● Inventory  ◈  │  GET  /scripts   │  ▸ run-request               │
│ «GEPA»        ★  │ «POST /run»      │  ▸ run-response              │
│                  │  GET  /events    │    script-def                │
│                  │  GET /schemas/…  │    event-entry               │
│                  │  GET  /timeline  │                              │
│                  │                  │  ▸ = referenced by           │
│                  │                  │    selected API              │
│                  │                  │                              │
│                  │                  │                              │
├──────────────────┴──────────────────┴──────────────────────────────┤
│ POST /api/apps/gepa/run                                            │
│────────────────────────────────────────────────────────────────────│
│ summary:         Execute a GEPA script by name or inline           │
│ tags:            script-runner                                     │
│                                                                    │
│ request_schema:  run-request                                       │
│ response_schema: run-response                                      │
│ error_schema:    —                                                 │
└────────────────────────────────────────────────────────────────────┘
```

---

## 10. Module Browser — GEPA → Schema selected

```
┌─ Module Browser ────────────────────────────────────────────── ○ □ ┐
│                                                            ⟳       │
├──────────────────┬──────────────────┬──────────────────────────────┤
│ Modules          │ APIs             │ Schemas                      │
│──────────────────│──────────────────│──────────────────────────────│
│                  │                  │                              │
│  ● Inventory  ◈  │  GET  /scripts   │ «run-request»                │
│ «GEPA»        ★  │  POST /run       │  run-response                │
│                  │  GET  /events    │  script-def                  │
│                  │  GET /schemas/…  │  event-entry                 │
│                  │  GET  /timeline  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
├──────────────────┴──────────────────┴──────────────────────────────┤
│ Schema: run-request                            format: json-schema │
│────────────────────────────────────────────────────────────────────│
│ uri: /api/apps/gepa/schemas/run-request                            │
│                                                                    │
│  {                                                                 │
│    "type": "object",                                               │
│    "properties": {                                                 │
│      "script_name": { "type": "string" },                          │
│      "args": { "type": "object" },                                 │
│      "timeout_ms": { "type": "integer" }                           │
│    },                                                              │
│    "required": ["script_name"]                                     │
│  }                                                                 │
└────────────────────────────────────────────────────────────────────┘
```

---

## 11. Module Browser — Inventory selected, reflection unavailable

```
┌─ Module Browser ────────────────────────────────────────────── ○ □ ┐
│                                                            ⟳       │
├──────────────────┬──────────────────┬──────────────────────────────┤
│ Modules          │ APIs             │ Schemas                      │
│──────────────────│──────────────────│──────────────────────────────│
│                  │                  │                              │
│ «Inventory»   ◈  │                  │                              │
│  ● GEPA       ★  │  Reflection is   │  Reflection is               │
│                  │  not available    │  not available               │
│                  │  for this module  │  for this module             │
│                  │  yet.             │  yet.                        │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
├──────────────────┴──────────────────┴──────────────────────────────┤
│ Inventory                                          ● healthy  ◈   │
│────────────────────────────────────────────────────────────────────│
│ id:          inventory                                             │
│ description: Inventory chat runtime, profiles, timeline,           │
│              and confirm APIs                                      │
│ required:    yes                                                   │
│ base:        /api/apps/inventory/                                  │
│ reflection:  not implemented (501)                                 │
└────────────────────────────────────────────────────────────────────┘
```

---

## 12. Health Dashboard (opened from menu Window → Health Dashboard)

```
┌─ Health Dashboard ─────────────────────────────────────────── ○ □ ┐
│                                         Last check: 14:32:07  ⟳   │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  Summary                                                           │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐            │
│  │       2       │ │       2       │ │       1       │            │
│  │    mounted    │ │    healthy    │ │   required    │            │
│  └───────────────┘ └───────────────┘ └───────────────┘            │
│                                                                    │
│  Modules                                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  ● Inventory     healthy    ◈ required     /api/apps/inve…  │   │
│  │  ● GEPA          healthy    ★ reflective   /api/apps/gepa/  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                    │
│                                                                    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## 13. Health Dashboard — degraded state

```
┌─ Health Dashboard ─────────────────────────────────────────── ○ □ ┐
│                                         Last check: 14:32:07  ⟳   │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ⚠ System degraded — 1 required module unhealthy                   │
│                                                                    │
│  Summary                                                           │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐            │
│  │       2       │ │    ⚠  1      │ │    ⚠  1      │            │
│  │    mounted    │ │    healthy    │ │   required    │            │
│  │               │ │   of 2       │ │   unhealthy   │            │
│  └───────────────┘ └───────────────┘ └───────────────┘            │
│                                                                    │
│  Modules                                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  ○ Inventory     UNHEALTHY  ◈ required     /api/apps/inve…  │   │
│  │    ┌────────────────────────────────────────────────────┐   │   │
│  │    │ database connection pool exhausted: dial tcp        │   │   │
│  │    │ 127.0.0.1:5432: connect: connection refused        │   │   │
│  │    └────────────────────────────────────────────────────┘   │   │
│  │                                                             │   │
│  │  ● GEPA          healthy    ★ reflective   /api/apps/gepa/  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## 14. Health Dashboard — many modules, future state

```
┌─ Health Dashboard ─────────────────────────────────────────── ○ □ ┐
│                                         Last check: 14:32:07  ⟳   │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  Summary                                                           │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐            │
│  │       6       │ │    ⚠  5      │ │       3       │            │
│  │    mounted    │ │    healthy    │ │   required    │            │
│  │               │ │   of 6       │ │               │            │
│  └───────────────┘ └───────────────┘ └───────────────┘            │
│                                                                    │
│  Modules                                          sort: health ▼   │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  ○ Auth          UNHEALTHY  ◈ required                      │   │
│  │    dial tcp 127.0.0.1:6379: connection refused              │   │
│  ├─────────────────────────────────────────────────────────────┤   │
│  │  ● Inventory     healthy    ◈ required                      │   │
│  │  ● Billing       healthy    ◈ required                      │   │
│  │  ● GEPA          healthy    ★ reflective                    │   │
│  │  ● Scheduler     healthy    ★ reflective                    │   │
│  │  ● Telemetry     healthy                                    │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## 15. Loading states

### Apps Folder loading
```
┌─ Mounted Apps ─────────────────────────────────────────── ○ □ ┐
│  ◀ ▶                                                ⟳    │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│                                                                │
│                                                                │
│                   Discovering mounted apps…                     │
│                                                                │
│                                                                │
│                                                                │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### Module Browser — reflection loading

```
┌─ Module Browser ────────────────────────────────────────────── ○ □ ┐
│                                                            ⟳       │
├──────────────────┬──────────────────┬──────────────────────────────┤
│ Modules          │ APIs             │ Schemas                      │
│──────────────────│──────────────────│──────────────────────────────│
│                  │                  │                              │
│  ● Inventory  ◈  │  ░░░░░░░░░░░░░  │  ░░░░░░░░░░░░░░░░░          │
│ «GEPA»        ★  │  ░░░░░░░░░░░░░  │  ░░░░░░░░░░░░░░░░░          │
│                  │  ░░░░░░░░░░░░░  │  ░░░░░░░░░░░░░░░░░          │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
│                  │                  │                              │
├──────────────────┴──────────────────┴──────────────────────────────┤
│ GEPA                                                 ● healthy ★  │
│────────────────────────────────────────────────────────────────────│
│ Loading reflection metadata…                                       │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## 16. Menu bar

```
┌────────────────────────────────────────────────────────────────────────────┐
│  ● Apps Browser    File    View    Module    Window    Help                 │
└────────────────────────────────────────────────────────────────────────────┘

  File ▼                  View ▼                   Module ▼
  ┌───────────────────┐   ┌─────────────────────┐  ┌─────────────────────────┐
  │ Refresh All   ⌘R  │   │ as Icons        ⌘1  │  │ Get Info            ⌘I  │
  ├───────────────────┤   │ as Browser      ⌘2  │  │ Inspect Reflection  ⌘R  │
  │ Close Window  ⌘W  │   ├─────────────────────┤  ├─────────────────────────┤
  └───────────────────┘   │ Show Unhealthy Only │  │ Copy Base URL           │
                          │ Show Required Only  │  │ Copy Reflection URL     │
  Window ▼                │ Show Reflective Only│  ├─────────────────────────┤
  ┌───────────────────┐   └─────────────────────┘  │ Health Check        ⌘H  │
  │ Mounted Apps      │                            ├─────────────────────────┤
  │ Module Browser    │                            │ Open Base URL in        │
  │ Health Dashboard  │                            │   System Browser        │
  └───────────────────┘                            └─────────────────────────┘
```

---

## 17. Desktop — all three windows coexisting

```
┌─ Health Dashboard ──────────────────────── ○ □ ┐
│                              14:32:07  ⟳       │
├────────────────────────────────────────────────┤
│  ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │    2     │ │    2     │ │    1     │       │
│  │ mounted  │ │ healthy  │ │ required │       │
│  └──────────┘ └──────────┘ └──────────┘       │
│ ┌──────────────────────────────────────────┐   │
│ │ ● Inventory   healthy  ◈   /api/apps/in… │   │
│ │ ● GEPA        healthy  ★   /api/apps/ge… │   │
│ └──────────────────────────────────────────┘   │
└────────────────────────────────────────────────┘
                    ┌─ Module Browser ─────────────────────── ○ □ ┐
                    │                                    ⟳       │
                    ├──────────────┬──────────┬──────────────────┤
                    │ Modules      │ APIs     │ Schemas          │
                    │──────────────│──────────│──────────────────│
                    │ ● Inventory◈ │GET /scr… │ run-request      │
                    │«GEPA»     ★ │POST /run │ run-response     │
                    │              │GET /eve… │ script-def       │
                    │              │GET /sch… │ event-entry      │
                    │              │GET /tim… │                  │
   ┌─ Mounted Apps ─────── ○ □ ┐  ├──────────┴──────────────────┤
   │  ◀ ▶              ⟳    │  │ GEPA  ● healthy  ★            │
   ├───────────────────────────┤  │ GEPA script runner module    │
   │                           │  │ /api/apps/gepa/              │
   │  ┌──────┐   ┌──────┐     │  └──────────────────────────────┘
   │  │ ◈ ●  │   │   ●  │     │
   │  │ ┌──┐ │   │ ┌──┐ │     │
   │  │ │▦▦│ │   │ │▦▦│ │     │
   │  │ └──┘ │   │ └──┘ │     │
   │  │      │   │  ★   │     │
   │  └──────┘   └──────┘     │
   │  Inventory    GEPA        │
   └───────────────────────────┘
```

---

# YAML Widget DSL + React Sketch

```yaml
# ═══════════════════════════════════════════════════════════════
# Apps Browser — Final Design
# Three window types sharing one data layer
# ═══════════════════════════════════════════════════════════════

app: apps-browser
version: 1

# ─── Shared Data Layer ────────────────────────────────────────

data:
  sources:
    apps_manifest:
      endpoint: GET /api/os/apps
      model: AppsManifestResponse
      polling: manual         # phase 1: manual refresh only
      cache_key: apps_list

    reflection:
      endpoint: GET /api/os/apps/{app_id}/reflection
      model: ModuleReflectionDocument
      fetch_strategy: lazy    # load on first inspect/select
      cache_key: reflection_{app_id}
      error_states:
        501: unsupported      # normal — not an error
        404: stale_selection  # suggest refresh
        500: internal_error   # show retry

  state:
    # — react sketch: useAppsBrowserStore() or context provider
    # — holds apps[], selectedAppId, reflection cache, load states
    # — single store shared across all three window types
    apps_list:
      type: AppManifestDocument[]
      load_state: idle | loading | loaded | error
      fetch_error: string | null
      last_refreshed_at: string | null

    selection:
      selected_app_id: string | null
      selected_api_id: string | null
      selected_schema_id: string | null

    reflection_cache:
      # map of app_id → ReflectionState
      type: Record<string, ReflectionState>
      # ReflectionState = idle | loading | loaded | unsupported | error

  actions:
    refresh_apps:
      # — fetch GET /api/os/apps, replace apps_list
      # — clear selection if selected app disappeared
      # — update last_refreshed_at
      trigger: toolbar refresh button, ⌘R, menu File → Refresh All

    select_module:
      # — set selected_app_id
      # — clear selected_api_id and selected_schema_id
      # — if module has reflection.available && not cached: fetch reflection
      trigger: click module in any list/icon

    select_api:
      # — set selected_api_id
      # — clear selected_schema_id
      # — compute cross-referenced schemas from api.request_schema/response_schema
      trigger: click API row in browser pane

    select_schema:
      # — set selected_schema_id
      # — if schema has uri and not yet fetched: optionally fetch live
      trigger: click schema row in browser pane

    health_check:
      # — alias for refresh_apps, could flash the health dot on the target
      trigger: context menu → Health Check, ⌘H

    copy_base_url:
      # — write /api/apps/{app_id}/ to clipboard
      trigger: context menu → Copy Base URL

    copy_reflection_url:
      # — write reflection.url to clipboard, noop if unavailable
      trigger: context menu → Copy Reflection URL

    open_get_info:
      # — open GetInfoWindow for selected_app_id
      trigger: context menu → Get Info, ⌘I

    open_browser_view:
      # — open ModuleBrowserWindow, pre-select current app
      trigger: context menu → Open in Browser, ⌘B, menu View → as Browser

    open_health_dashboard:
      # — open HealthDashboardWindow
      trigger: menu Window → Health Dashboard


# ─── Window: Mounted Apps (Icon Folder) ──────────────────────

windows:
  - id: mounted-apps-folder
    title: Mounted Apps
    kind: folder
    default: true            # opens on launch

    # — react sketch: <AppsFolderWindow />
    # — renders icon grid from apps_list
    # — toolbar: back/forward (noop phase 1), item count + health summary, refresh
    # — double-click icon → open_browser_view pre-selected to that module
    # — right-click icon → context_menu
    # — drag icon → copy base URL as text

    toolbar:
      left:
        - back_forward_buttons  # disabled in phase 1
      center:
        - status_summary
          # "{n} apps · {n} healthy · {n} required"
          # shows "⚠ {n} unhealthy" variant when any unhealthy
      right:
        - refresh_button

    content:
      type: icon_grid
      source: data.apps_list
      sort:
        - field: healthy
          order: asc          # unhealthy first
        - field: required
          order: desc         # required before optional
        - field: name
          order: asc

      icon:
        # — react sketch: <AppIcon app={app} />
        # — renders badge composition on a base module icon shape
        body: module_glyph    # ▦▦ normal, ░░ when unhealthy
        top_left:
          badge: required     # ◈ when app.required
          visible_when: app.required
        top_right:
          badge: health_dot   # ● green when healthy, ○ red when not
          color_when_healthy: green
          color_when_unhealthy: red
        bottom_center:
          badge: reflection   # ★ when reflection.available
          visible_when: app.reflection?.available
        label: app.name
        label_prefix_when_unhealthy: "⚠ "

    context_menu:
      # — react sketch: <AppContextMenu app={app} />
      # — items conditionally enabled based on app state
      items:
        - label: Get Info
          shortcut: ⌘I
          action: open_get_info
          enabled: always

        - label: Open in Browser
          shortcut: ⌘B
          action: open_browser_view
          enabled: always

        - separator

        - label: Inspect Reflection
          shortcut: ⌘R
          action: open_browser_view  # pre-selects module, loads reflection
          enabled: app.reflection?.available
          disabled_label: "(not available)"

        - separator

        - label: Copy Base URL
          action: copy_base_url
          enabled: always

        - label: Copy Reflection URL
          action: copy_reflection_url
          enabled: app.reflection?.available

        - separator

        - label: Health Check
          shortcut: ⌘H
          action: health_check
          enabled: always


  # ─── Window: Module Browser (Smalltalk-style) ──────────────

  - id: module-browser
    title: Module Browser
    kind: browser

    # — react sketch: <ModuleBrowserWindow />
    # — vertical split: top = three column panes, bottom = detail panel
    # — selection cascades left → right
    # — selecting module populates APIs + Schemas (or placeholder if no reflection)
    # — selecting API highlights cross-referenced schemas with ▸ marker
    # — selecting schema shows schema source in detail panel
    # — detail panel always shows the deepest selected item

    toolbar:
      right:
        - refresh_button

    layout:
      type: vertical_split
      ratio: [0.5, 0.5]

      top:
        # — react sketch: <BrowserColumns />
        type: horizontal_split
        ratio: [0.3, 0.35, 0.35]

        panes:
          - id: pane_modules
            # — react sketch: <ModuleListPane apps={apps} selected={selectedAppId} onSelect={selectModule} />
            header: Modules
            source: data.apps_list
            item:
              primary: app.name
              trailing_badges:
                - type: required    # ◈
                  visible_when: app.required
                - type: reflection  # ★
                  visible_when: app.reflection?.available
              leading:
                type: health_dot    # ● or ○
            selection: single
            on_select: select_module
            empty_state: null       # always has at least loaded apps

          - id: pane_apis
            # — react sketch: <APIListPane apis={reflection?.apis} selected={selectedApiId} onSelect={selectApi} crossRefs={...} />
            header: APIs
            source: reflection_cache[selected_app_id].doc.apis
            item:
              primary: "{method} {path}"
              monospace: true
            selection: single
            on_select: select_api
            empty_states:
              no_module_selected:
                text: ""
              reflection_unavailable:
                text: "Reflection is not available for this module yet."
                style: informational
              reflection_loading:
                skeleton: true

          - id: pane_schemas
            # — react sketch: <SchemaListPane schemas={reflection?.schemas} selected={selectedSchemaId} onSelect={selectSchema} highlighted={crossRefSchemaIds} />
            header: Schemas
            source: reflection_cache[selected_app_id].doc.schemas
            item:
              primary: schema.id
              leading:
                type: cross_ref_marker  # ▸ when schema is referenced by selected API
                visible_when: schema.id in cross_ref_set
            selection: single
            on_select: select_schema
            empty_states:
              # same structure as pane_apis
              reflection_unavailable:
                text: "Reflection is not available for this module yet."
                style: informational

      bottom:
        # — react sketch: <BrowserDetailPanel selection={...} />
        # — renders one of several detail views based on deepest selection
        id: detail_panel
        states:

          nothing_selected:
            text: "Select a module to inspect."

          module_selected:
            # — react sketch: <ModuleDetail app={app} reflection={reflectionState} />
            fields:
              - { label: id,          value: app.app_id }
              - { label: description, value: app.description, full_width: true }
              - { label: required,    value: app.required }
              - { label: base,        value: "/api/apps/{app_id}/", monospace: true }
              - { label: reflection,  value: computed_reflection_label }
            header:
              title: app.name
              right: health_badge + reflection_badge

          api_selected:
            # — react sketch: <APIDetail api={api} appId={appId} />
            header:
              title: "{method} /api/apps/{app_id}{path}"
            fields:
              - { label: summary,         value: api.summary }
              - { label: tags,            value: "api.tags | join ', '" }
              - { label: request_schema,  value: api.request_schema, link: true }
              - { label: response_schema, value: api.response_schema, link: true }
              - { label: error_schema,    value: api.error_schema, fallback: "—" }

          schema_selected:
            # — react sketch: <SchemaDetail schema={schema} />
            header:
              title: "Schema: {schema.id}"
              right: "format: {schema.format}"
            fields:
              - { label: uri, value: schema.uri, monospace: true }
            body:
              type: code_block
              language: json
              source: schema.embedded
              fallback: "Fetch from {uri} to view full schema."
              actions:
                - label: Fetch Live
                  # — fetch schema.uri, populate embedded into cache
                - label: Copy JSON
                  # — copy stringified embedded to clipboard


  # ─── Window: Get Info ───────────────────────────────────────

  - id: get-info
    title: "{app.name} — Get Info"
    kind: inspector
    spawned_by: open_get_info
    singleton_per: app_id    # one Get Info window per module

    # — react sketch: <GetInfoWindow app={app} reflection={reflectionState} />
    # — static info panel, no interactive selection
    # — sections rendered conditionally based on data availability
    # — icon rendered large at top with all badges

    layout:
      type: scroll
      sections:
        - id: header
          # — react sketch: <GetInfoHeader app={app} />
          # — large icon + app name + app id
          icon: full_app_icon   # same composition as folder icon, larger
          title: app.name
          subtitle: app.app_id

        - id: general
          header: General
          fields:
            - { label: Description, value: app.description }
            - { label: Required,    value: app.required, badge_when_true: "◈ Yes (required at startup)" }
            - { label: Base URL,    value: "/api/apps/{app_id}/", monospace: true }

        - id: health
          header: Health
          fields:
            - { label: Status,     value: health_label, icon: health_dot }
            - { label: Last check, value: data.last_refreshed_at }
          alert_when_unhealthy:
            body: app.health_error
            footer_when_required: "This is a required module. System may be degraded."

        - id: reflection
          header: Reflection
          when_available:
            fields:
              - { label: Available, value: "★ Yes ({reflection.version})" }
              - { label: URL,       value: reflection.url, monospace: true }
          when_unavailable:
            text: "This module does not publish reflective API metadata yet."
            style: informational

        - id: apis
          header: "APIs ({count})"
          visible_when: reflection loaded and apis present
          # — react sketch: <GetInfoAPIList apis={apis} />
          # — compact list: "METHOD /path    summary"
          type: compact_list
          source: reflection.apis
          item: "{method}  {path}    {summary}"

        - id: schemas
          header: "Schemas ({count})"
          visible_when: reflection loaded and schemas present
          # — react sketch: <GetInfoSchemaChips schemas={schemas} />
          type: chip_row
          source: reflection.schemas
          item: schema.id

      footer:
        actions:
          - label: Open in Browser
            action: open_browser_view


  # ─── Window: Health Dashboard ───────────────────────────────

  - id: health-dashboard
    title: Health Dashboard
    kind: dashboard
    spawned_by: open_health_dashboard
    singleton: true

    # — react sketch: <HealthDashboardWindow apps={apps} lastRefreshed={...} />
    # — no selection state of its own
    # — clicking a module row opens Get Info or selects in folder
    # — always sorted: unhealthy first, then required, then name

    toolbar:
      right:
        - last_check_label    # "Last check: 14:32:07"
        - refresh_button

    layout:
      type: scroll
      sections:
        - id: alert_banner
          # — react sketch: <DegradedBanner apps={apps} />
          # — only visible when any required module is unhealthy
          visible_when: any app where app.required && !app.healthy
          style: warning
          text: "⚠ System degraded — {n} required module(s) unhealthy"

        - id: summary_cards
          # — react sketch: <SummaryCards apps={apps} />
          # — three stat boxes side by side
          type: stat_row
          items:
            - label: mounted
              value: apps.length
              style: neutral

            - label: healthy
              value: apps.filter(healthy).length
              subtitle_when_degraded: "of {total}"
              style_when_degraded: warning

            - label: required
              value: apps.filter(required).length
              style: neutral
              # switches to warning style when any required is unhealthy
              style_when_any_required_unhealthy: warning
              label_when_any_required_unhealthy: "required unhealthy"

        - id: module_list
          # — react sketch: <HealthModuleList apps={apps} onClickModule={openGetInfo} />
          header: Modules
          type: list
          source: data.apps_list
          sort:
            - field: healthy
              order: asc
            - field: required
              order: desc
            - field: name
              order: asc

          item:
            # — react sketch: <HealthModuleRow app={app} />
            leading: health_dot
            primary: app.name
            secondary: health_label   # "healthy" or "UNHEALTHY"
            badges:
              - { type: required,   visible_when: app.required }
              - { type: reflection, visible_when: app.reflection?.available }
            trailing: truncated base_url

          item_expanded_when_unhealthy:
            # — when app.healthy == false, row auto-expands to show error
            # — react sketch: <HealthErrorBlock error={app.health_error} required={app.required} />
            body: app.health_error
            style: monospace_block
            footer_when_required: "This is a required module. System may be degraded."

          on_click_item: open_get_info
            # double-click or single-click opens Get Info window for that module


# ─── Menu Bar ─────────────────────────────────────────────────

menu_bar:
  # — react sketch: <AppsBrowserMenuBar />
  # — active when any Apps Browser window is frontmost
  # — Module menu items operate on the currently selected module (if any)

  menus:
    - label: File
      items:
        - { label: Refresh All,   shortcut: ⌘R,  action: refresh_apps }
        - separator
        - { label: Close Window,   shortcut: ⌘W,  action: close_frontmost_window }

    - label: View
      items:
        - { label: as Icons,      shortcut: ⌘1,  action: focus_or_open mounted-apps-folder }
        - { label: as Browser,    shortcut: ⌘2,  action: focus_or_open module-browser }
        - separator
        - { label: Show Unhealthy Only,  toggle: true, action: filter_unhealthy }
        - { label: Show Required Only,   toggle: true, action: filter_required }
        - { label: Show Reflective Only, toggle: true, action: filter_reflective }

    - label: Module
      # — all items require a selected module, grayed out otherwise
      items:
        - { label: Get Info,            shortcut: ⌘I, action: open_get_info }
        - { label: Inspect Reflection,  shortcut: ⌘R, action: open_browser_view, enabled: app.reflection?.available }
        - separator
        - { label: Copy Base URL,                      action: copy_base_url }
        - { label: Copy Reflection URL,                action: copy_reflection_url, enabled: app.reflection?.available }
        - separator
        - { label: Health Check,        shortcut: ⌘H, action: health_check }
        - separator
        - { label: Open Base URL in System Browser,    action: open_external_base_url }

    - label: Window
      items:
        - { label: Mounted Apps,     action: focus_or_open mounted-apps-folder }
        - { label: Module Browser,   action: focus_or_open module-browser }
        - { label: Health Dashboard, action: focus_or_open health-dashboard }


# ─── Navigation Map ──────────────────────────────────────────

navigation:
  # how windows connect to each other
  #
  #   ┌──────────────────┐
  #   │  Mounted Apps     │──── double-click ────▸ Module Browser (pre-selected)
  #   │  (icon folder)    │──── right-click ─────▸ Context Menu
  #   └──────────────────┘                             │
  #                                          Get Info ─┤── opens ──▸ Get Info Window
  #                                    Open in Browser ┤── opens ──▸ Module Browser
  #                                       Health Check ┘── triggers refresh
  #
  #   ┌──────────────────┐
  #   │  Module Browser   │──── schema link in detail ──▸ selects in schema pane
  #   │  (columns)        │
  #   └──────────────────┘
  #
  #   ┌──────────────────┐
  #   │  Health Dashboard │──── click module row ──▸ Get Info Window
  #   └──────────────────┘
  #
  #   ┌──────────────────┐
  #   │  Get Info Window  │──── Open in Browser ──▸ Module Browser (pre-selected)
  #   └──────────────────┘
```