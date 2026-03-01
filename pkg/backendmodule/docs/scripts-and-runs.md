---
Title: GEPA Scripts And Runs
DocType: guide
Topics:
  - scripts
  - runtime
  - timeline
Summary: "How GEPA scripts are listed and executed through run APIs."
Order: 2
---

# GEPA Scripts And Runs

Flow:

1. List scripts: `GET /scripts`
2. Start run: `POST /runs`
3. Poll run: `GET /runs/{run_id}`
4. Stream events: `GET /runs/{run_id}/events`
5. Read timeline: `GET /runs/{run_id}/timeline`
6. Cancel when needed: `POST /runs/{run_id}/cancel`

