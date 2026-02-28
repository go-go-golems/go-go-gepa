# Tasks

## Deferred For Later (Do Not Execute Today)

- [ ] 0. Confirm kickoff decision and scope freeze
- [ ] 0.1 Reconfirm with owner that GEPA-21 is still deferred and no production changes should ship yet
- [ ] 0.2 Document freeze note in ticket changelog before implementation starts
- [ ] 0.3 Link final GEPA-20 stabilization status as dependency

- [ ] 1. Baseline and architecture audit
- [ ] 1.1 Re-read GEPA-14 primary design doc and summarize queue-related assumptions
- [ ] 1.2 Re-read GEPA-14 intern Q&A doc section on pending queues and session intents
- [ ] 1.3 Re-read GEPA-17 viewer plan and extract lifecycle telemetry requirements
- [ ] 1.4 Validate current runtime files still match assumptions (`pluginIntentRouting`, `pluginCardRuntimeSlice`, `PluginCardSessionHost`)
- [ ] 1.5 Produce updated architecture map (event -> intent -> ingest -> route)

- [ ] 2. Contract preparation
- [ ] 2.1 Define canonical lifecycle states for runtime intents
- [ ] 2.2 Define correlation IDs (`interactionId`, `intentId`) and propagation rules
- [ ] 2.3 Specify route outcome payload schema for success/failure/denied
- [ ] 2.4 Define queue dequeue outcome events for viewer consumption
- [ ] 2.5 Add/update type definitions for lifecycle events

- [ ] 3. RuntimeIntentEffectHost implementation
- [ ] 3.1 Choose host form (component, middleware, or hybrid) and document rationale
- [ ] 3.2 Implement queue polling/select loop with deterministic ordering
- [ ] 3.3 Implement domain effect execution handler
- [ ] 3.4 Implement system effect execution handler
- [ ] 3.5 Implement nav-specific handling policy (or formalize nav as derived only)
- [ ] 3.6 Implement dequeue behavior after successful execution
- [ ] 3.7 Implement failure policy (retry/no-retry/dead-letter strategy)
- [ ] 3.8 Add idempotency guard to prevent duplicate processing

- [ ] 4. Routing cutover
- [ ] 4.1 Update `dispatchRuntimeIntent` to ingest-only behavior
- [ ] 4.2 Remove immediate domain dispatch path from routing layer
- [ ] 4.3 Remove immediate system dispatch path from routing layer
- [ ] 4.4 Ensure capability policy remains enforced in one canonical place
- [ ] 4.5 Add feature flag for queue-first rollout toggle

- [ ] 5. Runtime slice updates
- [ ] 5.1 Add explicit fields for in-flight and completed outcomes (if needed)
- [ ] 5.2 Ensure dequeue actions support exact-ID removal and fallback behavior
- [ ] 5.3 Add queue-length safety limits and overflow handling
- [ ] 5.4 Add stale-intent cleanup strategy for removed sessions

- [ ] 6. GEPA-17 viewer alignment
- [ ] 6.1 Extend event bus payload model for lifecycle transition events
- [ ] 6.2 Add viewer grouping by `interactionId` and `intentId`
- [ ] 6.3 Add queue status chips/states (queued, executing, success, failed, denied)
- [ ] 6.4 Add filters for queue lifecycle event families
- [ ] 6.5 Add copy/export support for correlated lifecycle traces

- [ ] 7. Testing and validation
- [ ] 7.1 Add unit tests for ingest-only routing semantics
- [ ] 7.2 Add unit tests for effect host dequeue ordering
- [ ] 7.3 Add tests for capability-denied queue behavior
- [ ] 7.4 Add tests for nav queue policy
- [ ] 7.5 Add integration test: UI event -> intent -> queue -> effect -> dequeue
- [ ] 7.6 Add viewer integration test for lifecycle correlation rendering
- [ ] 7.7 Run typecheck/test for `go-go-os`, `go-go-app-inventory`, and `wesen-os`

- [ ] 8. Observability and operations
- [ ] 8.1 Add diagnostics counters for queue depth and processing lag
- [ ] 8.2 Add logging guidelines for failed intent execution
- [ ] 8.3 Add operator troubleshooting runbook for stuck queues
- [ ] 8.4 Define SLOs for runtime intent processing latency

- [ ] 9. Rollout
- [ ] 9.1 Enable queue-first flag in development profile only
- [ ] 9.2 Run manual exploratory sessions with runtime debug + event viewer open
- [ ] 9.3 Compare queue-first traces against baseline mixed-mode behavior
- [ ] 9.4 Enable for broader internal usage if stable
- [ ] 9.5 Remove feature flag after stabilization (optional)

- [ ] 10. Documentation closure
- [ ] 10.1 Update GEPA-14 with final decision note and outcome
- [ ] 10.2 Update GEPA-17 with final aligned event semantics
- [ ] 10.3 Add final migration postmortem and lessons learned
- [ ] 10.4 Close GEPA-21 with links to merged PRs and tests
