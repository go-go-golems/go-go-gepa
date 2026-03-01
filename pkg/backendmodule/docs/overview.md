---
Title: GEPA Module Overview
DocType: guide
Topics:
  - backend
  - scripts
  - onboarding
Summary: "Overview of the GEPA script-runner backend module responsibilities."
Order: 1
---

# GEPA Module Overview

The GEPA backend module exposes script discovery and run lifecycle endpoints under:

- `/api/apps/gepa/...`

Primary responsibilities:

- Discover scripts from configured roots
- Start/cancel runs with concurrency and timeout controls
- Emit run events and timeline projections
- Expose reflection, schemas, and module docs endpoints

