# Tasks

## Completed

- [x] Create ticket workspace and base docs for `GEPA-01-EXTRACT-GEPPETTO-PLUGINS`.
- [x] Produce evidence-backed architecture analysis for plugin ownership boundaries.
- [x] Identify extractor + optimizer consumer blast radius (`geppetto`, `go-go-gepa`, legacy runners, extraction runner).
- [x] Define migration architecture with phased compatibility strategy.
- [x] Define how to carry a `registryIdentifier` in plugin metadata/reporting.
- [x] Maintain chronological investigation diary with command logs and findings.

## Next Implementation Tasks

- [ ] Add `go-go-gepa` native plugin contract module (`gepa/plugins`) for extractor + optimizer helpers.
- [ ] Add temporary compatibility alias `geppetto/plugins` in runtimes that migrate first.
- [ ] Add `registryIdentifier` field to optimizer + extractor descriptor decode metadata structs.
- [ ] Propagate `registryIdentifier` into host context, hook option tags, JSON reports, and CLI output metadata.
- [ ] Extend recorder schema with `plugin_registry_identifier` and update inserts/queries.
- [ ] Update docs/scripts to prefer `require("gepa/plugins")`.
- [ ] Add tests for metadata propagation and backward compatibility.
- [ ] Remove legacy alias after migration window.
