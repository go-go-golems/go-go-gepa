---
Title: Investigate JavaScript-Registered SEM Reducers and Event Handlers
Ticket: GEPA-06-JS-SEM-REDUCERS-HANDLERS
Status: active
Topics:
    - gepa
    - event-streaming
    - js-vm
    - sem
    - pinocchio
    - geppetto
    - go-go-os
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: design-doc/01-javascript-registered-sem-reducers-and-event-handler-architecture.md
      Note: Primary investigation and architecture recommendation
    - Path: reference/01-investigation-diary.md
      Note: Chronological command log and findings
    - Path: scripts/js-sem-reducer-handler-prototype.js
      Note: Prototype showing handler overwrite vs composable model
ExternalSources: []
Summary: 'Clarifies that geppetto already supports JS event handlers for geppetto events, pinocchio owns backend SEM projection in Go, and frontend SEM registration exists in go-go-os app runtime; proposes a staged plan for JS reducer/handler goals.'
LastUpdated: 2026-02-26T18:30:00-05:00
WhatFor: Define implementation path for JavaScript SEM reducer/handler extensibility
WhenToUse: Use when planning dynamic event reaction/projection features across geppetto/pinocchio/go-go-os
---

# Investigate JavaScript-Registered SEM Reducers and Event Handlers

## Overview

This ticket investigates where SEM projection and event reaction logic live today, and what is required to support JavaScript-registered reducers and handlers.

## High-Level Conclusions

1. `geppetto` already supports JavaScript event handlers for geppetto event stream (`start`/`partial`/`final`).
2. `pinocchio` currently performs backend SEM projection in Go; backend JS reducer registration is not present.
3. `go-go-os` app runtime already supports JS/TS SEM handler registration, but with single-handler overwrite semantics.
4. GEPA-04 streaming events are now part of the baseline and were incorporated into this analysis.

## Primary Artifacts

1. `design-doc/01-javascript-registered-sem-reducers-and-event-handler-architecture.md`
2. `reference/01-investigation-diary.md`
3. `scripts/js-sem-reducer-handler-prototype.js`

## Tracking

1. See [tasks.md](./tasks.md)
2. See [changelog.md](./changelog.md)
