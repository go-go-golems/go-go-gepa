---
Title: Implementation Plan for Option A GEPA Extraction
Ticket: GP-03-GEPA-REPO-MIGRATION
Status: active
Topics:
    - architecture
    - migration
    - geppetto
    - glazed
    - tools
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: gepa/go-gepa-runner/cmd/gepa-runner/main.go
      Note: |-
        New standalone CLI entrypoint for GEPA optimize/eval workflow
        Standalone GEPA CLI command surface and optimize/eval workflow
    - Path: gepa/go-gepa-runner/cmd/gepa-runner/plugin_loader.go
      Note: JS plugin contract loading and evaluator dispatch
    - Path: gepa/go-gepa-runner/pkg/optimizer/gepa/optimizer.go
      Note: |-
        Extracted GEPA loop logic now owned by gepa repository
        Extracted GEPA optimization loop and selection behavior
    - Path: geppetto/pkg/doc/topics/14-js-api-user-guide.md
      Note: |-
        Updated documentation to reference external GEPA runner example script
        Geppetto documentation link migration to external GEPA script
    - Path: geppetto/pkg/js/modules/geppetto/plugins_module.go
      Note: |-
        Generic optimizer plugin helper kept in geppetto
        Generic optimizer plugin helper retained in geppetto
ExternalSources: []
Summary: Detailed migration plan to move GEPA-specific implementation out of geppetto while preserving reusable JS plugin primitives in geppetto.
LastUpdated: 2026-02-23T22:05:00-05:00
WhatFor: Provide an actionable roadmap and guardrails for extracting GEPA into a dedicated repository.
WhenToUse: Use when implementing or reviewing the GEPA extraction and ownership boundary changes.
---


# Implementation Plan for Option A GEPA Extraction

## Goal

Move GEPA-specific code and operational docs from `geppetto/` into `gepa/`, leaving only generic JS plugin infrastructure in `geppetto`, with clear ownership boundaries, validated build/test paths, and a traceable migration diary.

## Current State Snapshot

The codebase currently has two active worktrees:

1. `geppetto/` (Go repository; branch `task/add-gepa-optimizer`)
2. `gepa/` (Python-first repository with a new Go submodule at `go-gepa-runner/`)

Migration work already started:

1. `go-gepa-runner/` has extracted code from `geppetto/pkg/optimizer/gepa` and `geppetto/cmd/gepa-runner`.
2. Existing GEPA docmgr tickets were physically moved to `gepa/ttmp/`.
3. `.ttmp.yaml` root now points to `gepa/ttmp`.
4. `geppetto` has pending deletions for GEPA runner/optimizer plus moved ticket directories.

## Architecture Boundary (Target)

### Geppetto should own

1. Generic JS module capability:
   - `geppetto/plugins` helper exports
   - optimizer descriptor validation API (`defineOptimizerPlugin` / version constant)
2. Core inference and tooling primitives consumed by downstream apps.
3. Documentation for the generic plugin contract.

### GEPA repository should own

1. Optimizer implementation and experimentation loop:
   - `pkg/optimizer/gepa`
2. Runnable optimizer CLI:
   - `cmd/gepa-runner` (`optimize`, `eval`, recorder/reporting)
3. GEPA-specific examples and benchmark scripts.
4. GEPA roadmap, migration tickets, and implementation diaries.

## Coupling Analysis (How Intertwined)

Coupling between systems is moderate and cleanly separable:

1. Compile-time coupling:
   - `go-gepa-runner` imports `github.com/go-go-golems/geppetto` for inference/session APIs.
   - No geppetto import should point back to GEPA package after migration.
2. Runtime coupling:
   - GEPA evaluator plugins run in JS runtime with `require("geppetto")` and `require("geppetto/plugins")`.
   - This is expected and desirable: GEPA is a client of geppetto runtime APIs.
3. Contract coupling:
   - Shared plugin API version string (`gepa.optimizer/v1`) lives in geppetto helper module.
   - GEPA runner validates descriptors against that contract.

Net: GEPA is dependent on geppetto, but geppetto does not need to depend on GEPA. That is the intended one-way dependency direction.

## Migration Phases

### Phase 1: Extraction and ownership split

1. Keep only generic plugin helper support in geppetto.
2. Remove geppetto-owned GEPA runner/optimizer code paths.
3. Ensure `go-gepa-runner` builds/tests independently.
4. Move active GEPA docs to `gepa/ttmp`.

### Phase 2: Hardening standalone GEPA runner

1. Improve `go-gepa-runner/README.md` from scaffold placeholder to real usage docs.
2. Validate optimize/eval smoke flows inside new repo.
3. Ensure CI/lint/test commands run from `go-gepa-runner` root.
4. Add migration verification artifacts to ticket `sources/`.

### Phase 3: Cleanup and developer ergonomics

1. Update geppetto docs to point to external GEPA repo scripts/docs.
2. Verify no stale geppetto references to removed GEPA packages.
3. Document ongoing maintenance boundary in migration ticket.

## Validation Gates

A migration step is complete only if all gates pass:

1. `cd geppetto && go test ./... -count=1`
2. `cd gepa/go-gepa-runner && go test ./... -count=1`
3. `cd gepa/go-gepa-runner && go build ./cmd/gepa-runner`
4. `rg` scan from `geppetto` shows no `pkg/optimizer/gepa` or `cmd/gepa-runner` references in code (docs may reference external URL/path).
5. Ticket tasks/changelog/diary are updated with commit hashes.

## Execution Plan (Detailed)

### Step A: Finalize geppetto cleanup commit

Pseudo-flow:

```text
stage deleted GEPA code + moved ticket deletions + doc link update
run go test ./... in geppetto
commit with message describing extraction cleanup
```

Expected result:

1. Geppetto repository no longer carries GEPA optimizer/runner implementation.
2. Generic optimizer plugin contract support remains available.

### Step B: Finalize gepa extraction commit

Pseudo-flow:

```text
stage go-gepa-runner module files
stage moved ttmp tickets under gepa/ttmp
run go test/build under go-gepa-runner
commit as first ownership commit in gepa repo
```

Expected result:

1. GEPA code and docs now live where they are maintained.
2. Migration is reviewable with one coherent commit in `gepa`.

### Step C: Add migration governance docs

Pseudo-flow:

```text
create GP-03 migration analysis plan + tasks + diary
record exact commands and outcomes
commit ticket docs
upload plan doc to reMarkable
```

Expected result:

1. Migration has reproducible checklist and rationale.
2. Human-readable handoff exists outside terminal logs.

### Step D: Work tasks sequentially with diary and checkpoint commits

Pseudo-flow:

```text
for each task batch:
  implement
  validate
  mark task done
  append diary step with prompt context + outcomes
  commit
```

Expected result:

1. Traceable implementation history.
2. Small reviewable increments.

## Risks and Mitigations

1. Risk: accidental breakage in geppetto docs/tests due removed code.
   - Mitigation: run full geppetto tests after deletion commit.
2. Risk: scaffolded `go-gepa-runner` placeholders confuse maintainers.
   - Mitigation: replace README TODO with concrete optimize/eval examples.
3. Risk: drift between old and new ticket locations.
   - Mitigation: keep only `gepa/ttmp` authoritative and commit deletions from geppetto.
4. Risk: untracked local binaries accidentally committed.
   - Mitigation: stage explicit paths only; verify `git diff --cached --name-only` before commit.

## What "Done" Looks Like

1. `geppetto` contains only generic plugin helper support and no GEPA runner/optimizer package.
2. `gepa/go-gepa-runner` owns runnable GEPA CLI + optimizer package.
3. Existing GEPA tickets and new migration ticket are stored under `gepa/ttmp`.
4. Migration plan is uploaded to reMarkable.
5. Ticket tasks are checked off with diary evidence and commit hashes.
