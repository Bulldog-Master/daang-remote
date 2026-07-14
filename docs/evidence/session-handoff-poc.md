# Session Handoff Contract PoC — Evidence Report

**Status:** experimental evidence, not production readiness.

This document accompanies `poc/session-handoff/`. Nothing in it constitutes a
security certification, an unlinkability claim, a complete compromise-
containment claim, or a post-quantum security claim.

---

## 1. Bounded question

Can Daang Remote implement the ADR-0003 and ADR-0004 Control Plane → Data
Plane handoff properties in a small, local, deterministic Go proof of
concept without:

- passing long-lived identity across the plane seam;
- creating transferable bearer authority;
- allowing one compromised session to gain authority over another session;
- silently weakening the accepted ADR contracts?

## 2. Repository state before implementation

- `Bulldog-Master/daang-remote` `main` at `70ce97a87a5dbf80e55742759bf6e7e280580cc7`.
- Contents at the repository root: `.github/`, `adrs/`, `README.md`.
- No implementation substrate. No `go.mod`, no `Cargo.toml`, no `package.json`
  for a runtime component, no source tree under `src/`, `internal/`, `pkg/`.
- ADR-0002, ADR-0003, ADR-0004 all merged and Accepted. Foundry 1
  authoritative and unmodified.

## 3. Why Go was selected for this experiment

Go was chosen only as an experimental substrate for a bounded, local,
deterministic PoC. The reasons that matter for this PoC:

- The standard library covers HMAC, SHA-256, `crypto/rand`, JSON encoding,
  the `testing` package (including subtests and fuzzing), and the race
  detector — no third-party dependency is required to model the contract.
- The build and test loop is fast enough to keep the red / green sequence
  visible in one session.
- Concurrency primitives (`sync.Mutex`, goroutines) plus `-race` provide
  cheap coverage for a Data Plane that must handle many concurrent
  independent sessions.

Go was **not** selected because Go's type system proves cross-session
containment. That property is cryptographic and state-isolation, and the
PoC tests it explicitly under adversarial exposure (§16).

## 4. Go is not the permanent product-language decision

This PoC does not select Go for Daang Remote. Rust, Go, and a split
Go/Rust architecture remain open. After this PoC, the substrate decision
must be re-evaluated in the presence of implementation evidence and
external substrate evidence (for example the xxDK integration surface).

## 5. Implementation structure

```
poc/session-handoff/
├── go.mod                    // module root, Go 1.25
├── README.md                 // experimental framing, run instructions
├── handoff.go                // mock Issuer + mock Validator + Artifact + Event
└── handoff_test.go           // 30 mandatory tests + concurrency + fuzz seed
```

No controllers, services, repositories, adapters, infrastructure, internal
platform layers, or production package hierarchies were introduced.

## 6. Experimental cryptographic construction

- **Issuer key.** 32 random bytes from `crypto/rand`. Represents *global
  issuer authority*.
- **Per-session subkey.** `HMAC-SHA256(issuerKey, "dhr-poc/session-key/v1|"
  || sessionID)`. Represents *session-local material* for a single session.
  Delivered to the Data Plane validator out-of-band (not modelled in this
  PoC).
- **Artifact signature.** `HMAC-SHA256(sessionKey, canonical(fields))` over
  a deterministic JSON encoding of the artifact excluding `Signature` and
  `RecipientBind`.
- **Recipient binding.** `HMAC-SHA256(sessionKey,
  "dhr-poc/recipient-bind/v1|" || recipient)`, verified separately from the
  main signature.
- **Freshness / replay.** Random 128-bit nonce per artifact. Validator
  records seen nonces per session and rejects duplicates. Single-use.
- **Revocation / invalidation authority.** Requires the caller to present
  the *session-local key* (or, by extension, the issuer key). Neither is
  the artifact signature.

Every choice above is **experimental**, **non-production**, **replaceable**,
and **selected only to make the architectural property testable**. It is
not evidence of production security, and it is not evidence of post-quantum
security. The PoC deliberately did not select the final token format,
final signature or MAC scheme, final key hierarchy, final key lifecycle,
final session-binding protocol, final identity system, final transport
protocol, or final PQ mechanism.

**Trust-level modelling requirement.** The experimental construction must
model at least two distinct trust levels: global issuer authority must
remain separate from session-local authorization or validation material.
The partial-compromise tests in §16 depend on that separation. This models
the property under test and does not select a production key hierarchy.

## 7. Assumptions

- Out-of-band delivery of the per-session key from Issuer to Validator is
  assumed. The PoC does not model transport, TLS, xxDK, or key wrapping.
- The clock is monotonic-enough for tests; a 5-second forward-skew tolerance
  is allowed for issued-at.
- The adversary in §16 has ONLY the per-session key for Session A. The
  global issuer key is uncompromised.
- The validator's `self` role name is the recipient token; a real Data
  Plane would need to bind to a stronger recipient identity.

## 8. Language-neutral test specifications

The Go tests implement the specifications below. Each specification is
stated in terms of properties, preconditions, adversarial action, expected
outcome, and evidence, so it can be reimplemented in Rust or another
language without changing its security intent.

Every spec below shares this format:

- **Property:** the invariant the PoC must uphold.
- **Preconditions:** what state must exist before the action.
- **Action:** what the caller / adversary attempts.
- **Expected outcome:** accept or reject.
- **Evidence:** the Go test that exercises the spec.

| #  | Property                                                         | Preconditions                                                   | Action                                                | Expected  | Evidence Go test |
|----|------------------------------------------------------------------|-----------------------------------------------------------------|-------------------------------------------------------|-----------|------------------|
| 1  | Valid artifact is accepted                                       | Session installed, cap granted                                  | Verify + Use view_screen                              | accept    | `TestValidArtifactAccepted` |
| 2  | Altered artifact is rejected                                     | Valid artifact                                                  | Mutate `Purpose`                                      | reject    | `TestAlteredArtifactRejected` |
| 3  | Expired artifact is rejected                                     | Short TTL                                                       | Sleep past expiry, Verify                             | reject    | `TestExpiredArtifactRejected` |
| 4  | Wrong recipient is rejected                                      | Artifact issued to other role                                   | Verify                                                | reject    | `TestWrongRecipientRejected` |
| 5  | Wrong session is rejected                                        | Two sessions                                                    | Overwrite `SessionID`                                 | reject    | `TestWrongSessionRejected` |
| 6  | Wrong purpose is rejected                                        | Valid artifact                                                  | Mutate `Purpose`                                      | reject    | `TestPurposeTamperRejected` |
| 7  | Replay is rejected                                               | Valid artifact used once                                        | Use twice                                             | reject    | `TestReplayRejected` |
| 8  | Revoked capability is rejected                                   | Session cap revoked with session key                            | Use revoked cap                                       | reject    | `TestRevokedRejected` |
| 9  | Invalidated session is rejected                                  | Session invalidated with session key                            | Verify                                                | reject    | `TestInvalidatedRejected` |
| 10-14 | View-only cannot escalate                                     | Only view_screen granted                                        | Use keyboard/pointer/clipboard/audio/file             | reject    | `TestViewOnlyCannotEscalate` |
| 15 | Clipboard does not imply file transfer                           | Only clipboard granted                                          | Use file transfer                                     | reject    | `TestClipboardNoFileTransfer` |
| 16 | File transfer does not imply input                               | Only file_transfer granted                                      | Use keyboard/pointer                                  | reject    | `TestFileTransferNoInput` |
| 17 | Revocation is per-capability                                     | Two caps granted; one revoked                                   | Use the other                                         | accept    | `TestRevokeIsolation` |
| 18 | Expiry does not cross sessions                                   | Session A expired; Session B fresh                              | Use Session B                                         | accept    | `TestExpirySessionIsolation` |
| 19 | Invalidation does not cross sessions                             | Session A invalidated                                           | Use Session B                                         | accept    | `TestInvalidateSessionIsolation` |
| 20 | One session's artifact cannot authorize another                  | Two sessions                                                    | Move A's artifact onto B                              | reject    | `TestArtifactCannotCrossSession` |
| 21 | Artifact carries no prohibited long-lived identity fields        | Valid artifact                                                  | Inspect JSON representation and struct                | pass      | `TestArtifactHasNoProhibitedFields` |
| 22 | Return-flow events are bounded lifecycle only                    | Session with revoke + invalidate                                | Inspect every recorded event                          | pass      | `TestEventsBounded` |
| 23 | Return-flow events have no prohibited identity fields            | Session ended                                                   | Inspect events                                        | pass      | `TestEventsHaveNoProhibitedFields` |
| 24 | Return-flow state does not create user/account records           | Session lifecycle                                               | Inspect Validator surface                             | pass      | `TestNoLongLivedUserRecord` |
| 25 | Exposure of A's key cannot create B authority                    | A key exposed; issuer key safe                                  | Forge B artifact with A's key                         | reject    | `TestPartialCompromise_CannotCreateB` |
| 26 | Exposure of A's key cannot validate forged B artifacts           | A key exposed                                                   | Multiple forged B (varying purpose/caps)              | reject    | `TestPartialCompromise_CannotValidateForgedB` |
| 27 | Exposure of A's key cannot authorize B capability                | A key exposed                                                   | Forge B artifact for every capability                 | reject    | `TestPartialCompromise_CannotAuthorizeB` |
| 28 | Exposure of A's key cannot revoke B                              | A key exposed                                                   | Revoke B with A's key                                 | reject    | `TestPartialCompromise_CannotRevokeB` |
| 29 | Exposure of A's key cannot invalidate B                          | A key exposed                                                   | Invalidate B with A's key                             | reject    | `TestPartialCompromise_CannotInvalidateB` |
| 30 | Session B remains independently valid after A material exposed   | A key exposed                                                   | Use B legitimately                                    | accept    | `TestPartialCompromise_BRemainsValid` |

The Go test file is the executable form; the table above is the source of
truth for the property being tested and is the version that would carry
across to a Rust reimplementation.

## 9. Commands run

```
go version
# → go version go1.25.7 linux/amd64

go test ./...
go test -race ./...
go test -run=^$ -fuzz=FuzzVerifyMalformed -fuzztime=2s ./...
```

## 10. Failing-test evidence (red phase)

The initial `handoff.go` shipped as a compile-only stub whose method bodies
returned `nil` or `errors.New("unimplemented")`. Running the full test file
against that stub produced **28 failing tests / 0 passing**, including the
concurrency and partial-compromise tests, plus a fuzz panic. Excerpt from
`/tmp/red.txt`:

```
--- FAIL: TestValidArtifactAccepted (0.00s)
    handoff_test.go:38: issue: unimplemented
--- FAIL: TestAlteredArtifactRejected (0.00s)
--- FAIL: TestReplayRejected (0.00s)
--- FAIL: TestRevokedRejected (0.00s)
--- FAIL: TestInvalidatedRejected (0.00s)
--- FAIL: TestPartialCompromise_CannotCreateB (0.00s)
--- FAIL: TestPartialCompromise_CannotRevokeB (0.00s)
--- FAIL: TestPartialCompromise_CannotInvalidateB (0.00s)
...
FAIL
```

Because the workflow used to record the red phase reused a single file
(not two separate branches), the red output was captured to a scratch file
during construction; the same output is reproducible by replacing every
method body in `handoff.go` with `nil` / `errors.New("unimplemented")` and
re-running `go test ./...`. The commit sequence on this branch preserves
the intent (`test: add failing handoff contract tests`, then `feat:
implement experimental handoff contract`) but the red run itself lives in
this document, not in a runnable commit — recorded honestly per the brief.

## 11. Passing-test evidence (green phase)

```
$ go test ./...
ok  github.com/Bulldog-Master/daang-remote/poc/session-handoff  0.035s
```

Summary of the verbose run:

- Package count: **1**
- Test count: **28 top-level tests** (30 mandatory numbered specifications
  covered; several specifications share a single table-driven Go test) +
  **1 fuzz function seeded with 3 corpus inputs** + **6 subtests** in
  `TestVerifyBadInputs`.
- Passed: **all**
- Failed: **0**
- Skipped: **0**
- Go version: **go1.25.7 linux/amd64**

## 12. Race-detector evidence

```
$ go test -race ./...
ok  github.com/Bulldog-Master/daang-remote/poc/session-handoff  1.101s
```

`TestConcurrentSessions` exercises 32 independent sessions with 20
operations each across goroutines; no race was reported.

## 13. Fuzzing evidence

```
$ go test -run=^$ -fuzz=FuzzVerifyMalformed -fuzztime=2s ./...
fuzz: elapsed: 0s, gathering baseline coverage: 3/3 completed, now fuzzing with 64 workers
fuzz: elapsed: 3s, execs: 314138 (104708/sec), new interesting: 73 (total: 76)
PASS
ok  github.com/Bulldog-Master/daang-remote/poc/session-handoff  3.048s
```

No malformed artifact verified. No crash. 314,138 executions in ~3 seconds.

## 14. What the PoC demonstrates

Under the experimental construction and the modelled adversary:

- The Control Plane can hand session-scoped authority to the Data Plane
  without embedding long-lived identity.
- The Data Plane can validate every property (session, recipient, purpose,
  freshness, expiry, replay, revocation, invalidation, session binding,
  recipient binding) *independently* — it does not re-run user
  authentication.
- Capabilities are strictly independent; no capability implies another;
  deny-by-default is enforced.
- Return-flow events remain bounded to lifecycle state and carry no user,
  account, device, or content-derived material.
- Exposure of one session's per-session key does not create, validate,
  authorize, revoke, or invalidate any other session, and does not
  compromise the global issuer authority.

## 15. What the PoC does not demonstrate

- Production security, unlinkability, complete compromise containment, or
  post-quantum resistance.
- Any property of a real transport (xxDK, cMixx, QUIC, WebRTC, TLS).
- Any property of real OS integration (screen capture, input injection,
  clipboard, audio, file transfer).
- Any solution to event timing, order, frequency, or cross-session
  metadata patterns.
- Any final choice of token format, MAC/signature scheme, key hierarchy,
  key lifecycle, session-binding protocol, identity system, or PQ
  mechanism.
- Anti-forensic properties of the exposed session's own data.

**Single-process, in-memory scope.** The PoC uses single-process, in-memory
replay, revocation and invalidation state. It does not demonstrate
distributed replay detection, revocation propagation, consistency across
validators, recovery after validator restart, or resistance to
partition-related races.

## 16. Partial-compromise scenario and results

**Setup.** Issuer creates two sessions A and B, each with an independent
session-local key derived from the same uncompromised issuer key. Session
A is granted every capability; Session B is granted every capability. The
adversary is given exactly `sessionKey_A`.

**Attempts and results.**

| Attempt                                              | Result |
|------------------------------------------------------|--------|
| Forge Session B artifact using A's key               | Rejected (`bad signature`) |
| Try every purpose / capability combination for B     | Rejected |
| Attempt to Use each capability on B                  | Rejected |
| Revoke a Session B capability with A's key           | Rejected (`bad proof`) |
| Invalidate Session B with A's key                    | Rejected (`bad proof`) |
| Continue using Session B legitimately                | Accepted |

Every cross-session attempt failed. Session B remained independently valid.

This demonstrates only the modelled property under the experimental
construction. It does not prove complete compromise containment for
Daang Remote overall.

## 17. Did Go help or obstruct the accepted ADRs?

- **ADR-0002 plane separation.** Helped. Splitting the module into
  `Issuer` and `Validator` types with no shared mutable state made the
  seam explicit. Nothing in the language forced a leak across it.
- **ADR-0003 handoff contract.** Helped. `crypto/hmac` + `encoding/json` +
  `crypto/rand` were sufficient to express the contract without pulling
  in a framework. Determinism of the canonical encoding was straightforward.
- **ADR-0004 Data Plane enforcement.** Helped for this scope. The
  independent per-session state made every required check trivially
  local. `sync.Mutex` plus the race detector kept concurrency honest.
  Go did **not** demonstrate anything about the media path or OS-level
  integration; those remain open.

Where Go did **not** help: it neither proved nor could prove the
cross-session containment property structurally. That property had to be
tested cryptographically under adversarial exposure (§16), which is
language-neutral.

## 18. ADR obligations not cleanly implementable

None encountered inside this PoC's scope. The obligations exercised
(session-scoped authority, independent Data Plane validation, capability
independence, revocation, invalidation, return-flow event bounds) all
implemented cleanly.

## 19. Architectural conflicts requiring another bounded PoC

- **Media path substrate.** This PoC did not implement screen capture,
  input injection, or transport. If those layers force a Rust or
  split-language decision, that finding will not surface here.
- **Real key delivery.** The PoC assumed out-of-band key delivery. The
  real handoff needs a bounded PoC that models an actual transport (for
  example xxDK) and its failure modes.

## 20. Evidence supporting an accepted ADR

- ADR-0003's separation of "issuance" from "validation" survived contact
  with an executable Data Plane that never consulted the Issuer at
  authorization time.
- ADR-0004's independent-enforcement requirement survived: every check
  ran locally against per-session state.

## 21. Evidence challenging an accepted ADR

None so far. This PoC did not exercise the transport-level or media-path
obligations of ADR-0002 / ADR-0004, so it cannot challenge those parts.

## 22. Reversal cost observed after implementation

Low. The PoC is one Go module, one implementation file, one test file, no
dependencies, ~600 lines. Reversing the substrate decision (for example
by porting to Rust) reduces to reimplementing the language-neutral test
specifications in §8 and matching the interface surface of `Issuer` and
`Validator`.

## 23. Recommendation for the post-PoC language/component-boundary decision

Do not select a permanent product language on the strength of this PoC.
The evidence supports these next steps, in order:

1. Reopen the substrate decision *after* the media-path and transport
   substrate analyses are in front of the founder.
2. Consider a bounded Rust reimplementation of the same test
   specifications in §8, purely to compare reversal cost, ergonomics, and
   any failure the Go construction masks.
3. Consider a bounded PoC that models a real transport (for example
   xxDK) with the handoff contract layered on top; treat that PoC as the
   substrate-decision trigger, not this one.
4. Only after (1)-(3) should the founder commit to Go, Rust, or a split
   architecture for the production Control Plane / Data Plane.

## 24. Not-claims

This PoC does not claim production readiness, security certification,
unlinkability, complete compromise containment, or post-quantum security.

## 25. Gate-review corrections

The four independent gate reports on PR #5 against frozen commit
`f3a58b0f` produced a consolidated set of corrections. Two required code
changes (behavior is the evidence at implementation-review stage); the
rest are documentary clarifications that add no new claims.

### 25.1 Code corrections

- **A1 — Issuer.SessionKey removed from production surface.** The
  exported accessor `Issuer.SessionKey(sessionID)` has been deleted from
  `handoff.go`. Documenting it as "test-only" while leaving it linkable
  from production code weakens the very boundary the PoC is designed to
  test. Partial-compromise tests (#25–#30) obtain Session A's local key
  from `StartSession`'s return value, exactly as legitimate installers
  would. A test-only helper `testSessionKey` lives in `export_test.go`
  in the same package; it is compiled into the test binary only and is
  unreachable from any production consumer of the package.
- **P2 — capability_denied event narrowed to true capability denials.**
  `Validator.Use` previously emitted `capability_denied` for every
  failure path, including invalid signature, wrong recipient, unknown or
  invalidated session, expired or future-dated artifact, replayed
  nonce, and nil artifact. Under ADR-0004's bounded return-flow
  contract, those are validation failures, not capability denials —
  emitting the same lifecycle event for malformed or hostile artifacts
  turned the event stream into a probing oracle. `Use` now emits
  `capability_denied` only after the artifact has passed every
  authenticity, integrity, binding, freshness, replay and session check
  but the requested capability is absent from the artifact, absent from
  the installed session grant set, or has been revoked. Validation
  failures fail closed with no return-flow event. Two new tests enforce
  this: `TestUseValidationFailuresEmitNoCapabilityDenied` (#31,
  negative) and `TestUseGenuineCapabilityDenialEmitsEvent` (#32,
  positive).

### 25.2 Evidence-only clarifications

- **S1.** This PoC treats possession of the raw session-local key as
  authorization proof for revoke, invalidate and end-session operations.
  There is no challenge-response, no session-side counter, and no
  freshness protocol for administrative operations. A production design
  would replace raw-key possession with a bounded proof-of-possession
  protocol.
- **S2.** The construction assumes that exposure of a session-derived
  HMAC subkey does not reveal the global issuer key. This is the
  standard PRF assumption on HMAC-SHA256 and is required for the
  partial-compromise results in §16; it is recorded as an assumption,
  not proved by the PoC.
- **S3.** Issued-at handling is intentionally asymmetric: only
  excessive future skew (>5s ahead of the validator clock) is rejected.
  No independent lower-bound age check exists beyond the signed
  `ExpiresAt`. A production design would add a bounded past-age cap.
- **A2.** Out-of-band delivery of the session-local key from Issuer to
  Validator is assumed and not modelled. A separate bounded PoC would
  model an actual transport (for example xxDK) and its failure modes.
- **P1.** Return-flow events (`Event.At`) use precise wall-clock
  timestamps from `time.Now`. This PoC does not model timing-correlation
  defenses; a production Data Plane would coarsen or bucket event
  timestamps before persistence or emission.
- **Q1.** The red-phase evidence in §10 is documentary: it records the
  failing outputs against the compile-only stub that existed before the
  implementation was committed. The green-phase reproduction in §11 and
  the fuzz/race evidence in §12–§13 are executable and reproducible
  from the current tree with the commands in §9.
- **Q2.** The fuzz target `FuzzVerifyMalformed` mutates exactly three
  fields — `Signature`, `Nonce`, and `RecipientBind`. Targeted tests,
  not fuzzing, cover mutation of `Recipient` (#4), `SessionID` (#5,
  #20), `Purpose` (#2, #6), `Capabilities` (#10–#16), and `Version`
  (rejected structurally in `verifyLocked`).
- **Q3.** Concurrent partial-compromise coverage is future scope. The
  race-detector run in §12 exercises concurrent legitimate use across
  many sessions; it does not drive an adversary concurrently with a
  legitimate holder of Session B. A follow-up PoC would add this.

### 25.3 Re-run after corrections

The corrections were verified locally against Go 1.25.7:

```
go test ./...        # PASS
go test -race ./...  # PASS (no data race reported)
go test -run='^$' -fuzz=FuzzVerifyMalformed -fuzztime=2s ./...  # PASS
```

No ADR-0002 / ADR-0003 / ADR-0004 changes were made. Foundry was not
modified. No production mechanism was selected. Founder approval on
PR #5 remains unchecked.

## 26. Go toolchain floor correction

This section is a post-merge accuracy correction to the module
directive originally shipped with the PoC. It changes evidence and
one build-time directive only. No PoC code, tests, ADRs
(ADR-0002 / ADR-0003 / ADR-0004), or Foundry material are modified,
and the historical record of merged PR #5 is preserved.

### 26.1 What is being corrected

- The PoC was originally generated with a Go 1.25 module directive
  (`go 1.25` in `poc/session-handoff/go.mod`).
- Review of the PoC source found no Go 1.25-specific language,
  standard-library, or toolchain feature in use.
- The module floor is therefore being lowered to `go 1.22`, which is
  the lowest Go version for which direct execution evidence currently
  exists (see 26.2). This does not select the permanent Daang Remote
  product toolchain and does not establish xxDK compatibility.

### 26.2 Evidence provenance (Go 1.22.2 execution)

- An independent external reviewer (not this Lovable session, and not
  the drafting agent) temporarily lowered the `go` directive to
  `1.22` and executed the otherwise unchanged PoC source with
  Go 1.22.2.
- That reviewer reported the standard test suite passing and the race
  detector reporting no data races under Go 1.22.2.
- The reviewer did NOT independently verify a meaningful active fuzz
  campaign under Go 1.22.2. The earlier §25.3 fuzz line
  (`-fuzztime=2s`) is a smoke check under Go 1.25.7 only and is not
  extended to Go 1.22.2 by this correction.
- This Lovable environment did not rerun `go test`, `go test -race`,
  or `go test -fuzz` against the repository. All Go 1.22.2 results in
  this section are attributed to the independent reviewer that ran
  them.

### 26.3 xxDK v4.8.4 compatibility (bounded)

- xxDK v4.8.4's published build instructions reference Go 1.17.X and
  GCC/cgo. Directly retrieved from
  `https://gitlab.com/elixxir/client/-/raw/v4.8.4/go.mod` (tag
  `v4.8.4` of `gitlab.com/elixxir/client/v4`), the exact directives
  present in that `go.mod` are:

  ```
  module gitlab.com/elixxir/client/v4

  go 1.21
  ```

  No `toolchain` directive is present in that file.
- xxDK is NOT added as a dependency of the PoC by this change. The
  module directive alignment here (`go 1.22` ≥ xxDK's `go 1.21`) is
  informational only and does not constitute an xxDK integration
  claim.

### 26.4 Out of scope for this correction

- Distributed replay detection across data-plane nodes.
- Revocation propagation across data-plane nodes.
- Selection of the permanent product toolchain floor.
- Any change to authentication, capability semantics, session-key
  handling, or the four-gate record of the merged PR #5.
