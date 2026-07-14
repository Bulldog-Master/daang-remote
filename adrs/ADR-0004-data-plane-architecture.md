# ADR-0004: Data Plane architecture

- **Status:** Proposed
- **Type:** Operational
- **Expires:** Does not expire
- **Date:** 2026-07-14
- **Foundry generation:** 1
- **Product:** Daang Remote
- **Supersedes:** none
- **Superseded by:** none
- **Related:** ADR-0002; ADR-0003; Daang Remote product charter

---

## Context

ADR-0002 established two logically separate planes: a Control Plane that
establishes authorized sessions and a Data Plane that carries the interactive
remote-access work of an established session. ADR-0003 defined the Control
Plane's bounded architecture and the plane-crossing handoff, including its
minimization, session and purpose scoping, lifetime bounds, recipient
binding, replay resistance, revocation and invalidation, prohibition of
direct or derivable stable identity across the seam, restricted return flow,
least-privilege capability model, separation of authentication from
authorization, and bidirectional trust assumptions.

ADR-0004 answers exactly one bounded question:

> How does the Daang Remote Data Plane operate an authorized remote-access
> session while conforming to ADR-0002's plane separation and ADR-0003's
> handoff, trust, privacy, security, and return-flow contracts, without
> selecting a transport, codec, cryptographic primitive, programming
> language, framework, or implementation?

This ADR does not redesign the Control Plane and does not renegotiate the
handoff contract because a different Data Plane shape might be simpler.
It also does not select any mechanism and does not begin implementation.

## Decision

Daang Remote introduces a bounded **Data Plane architecture** whose only
responsibility is to operate a single authorized remote-access session
between an already-authorized initiator role and an already-authorized
target role, using only the authority delivered by the Control Plane
handoff.

The Data Plane:

1. Consumes the handoff artifact defined by ADR-0003 and independently
   validates it before doing any session work.
2. Carries interactive input, output, and related media/data streams for the
   session's declared purpose.
3. Enforces the session's capability set, purpose scope, lifetime, and
   revocation state at every operation.
4. Emits only the minimum bounded return-flow signals that the Control Plane
   requires for lifecycle, capability, and health management.
5. Does not perform user authentication, does not issue capabilities, does
   not extend authority, does not persist session material beyond the
   session, and does not promote transient Data Plane metadata into
   long-lived identity or account state.
6. Selects no transport, codec, cryptographic primitive, token format,
   language, framework, deployment topology, or platform. All such choices
   are deferred to later mechanism-selection ADRs backed by product
   evidence.

The Data Plane is subordinate to ADR-0002 (plane separation) and to
ADR-0003 (Control Plane contract and handoff). Where architectural judgment
would weaken an ADR-0003 obligation, the obligation wins and the tension is
recorded as a conflict routed to a bounded proof-of-concept.

## ADR-0003 conformance obligations

The Data Plane treats the following ADR-0003 requirements as binding
architectural obligations, not as design suggestions:

- **Minimized handoff.** The Data Plane must operate on the minimum
  necessary information delivered through the handoff. It must not depend
  on out-of-band identity, durable account state, or historical linkage.
- **Session-scoped and purpose-scoped authority.** Every operation must be
  bounded to the specific session and the specific negotiated purpose
  carried in the handoff. Cross-session or cross-purpose reuse is
  prohibited.
- **Limited lifetime.** Authority is valid only within the negotiated
  lifetime and must be treated as invalid outside it, independent of
  whether an explicit revocation was received.
- **Recipient binding.** Authority is valid only for the Data Plane
  recipient/role intended by the Control Plane; artifacts intercepted or
  replayed to a different recipient must not be usable.
- **Replay resistance.** The Data Plane must reject reuse of a handoff
  artifact or capability outside its intended single-session context.
- **Revocation and invalidation.** The Data Plane must be able to observe
  and honor explicit revocation and separately honor session-state
  invalidation. Expiry, explicit revocation, and session invalidation are
  distinct events with distinct handling.
- **Non-transferability where required.** Where later threat analysis
  requires it, the Data Plane must treat authority as non-transferable and
  reject use by any party other than the bound recipient, including
  proof-of-possession semantics rather than mere bearer semantics.
- **No direct or derivable stable identity across the seam.** The Data
  Plane must not require, receive, derive, log, or emit stable identifiers
  linking a session to a durable user or device across the seam. Routing
  and correlation metadata carried through the handoff must remain
  ephemeral, session-scoped, and non-linkable across sessions by design.
- **Restricted return flow.** The Data Plane must return only the bounded
  set of signals the Control Plane actually needs (see Return-flow
  lifecycle and metadata). It must not open a general-purpose reporting
  channel.
- **Least-privilege capability enforcement.** The Data Plane must deny by
  default, permit only granted capabilities, and refuse any escalation
  attempt sourced from Data Plane content or metadata.
- **Separation of authentication from authorization.** The Data Plane does
  not authenticate users. It consumes already-granted capabilities and
  enforces them. Establishing claims about who a party is is the Control
  Plane's job; granting the Data Plane the right to do a specific thing
  is authorization delivered through the handoff.
- **Bidirectional trust assumptions.** The Data Plane trusts the Control
  Plane only for correctly-issued authority under ADR-0003; the Control
  Plane trusts the Data Plane only to enforce capabilities and return
  bounded signals honestly. Neither trusts network intermediaries,
  transport providers, relays, or endpoint operating systems beyond what
  is explicitly analyzed at each local trust boundary.
- **No promotion of transient state.** Transient Data Plane metadata (peer
  addresses, capability identifiers, ephemeral session identifiers, event
  timings, per-session counters) must not be promoted into long-lived
  identity or account state by any Data Plane component.

Silently weakening, reinterpreting, or replacing any of these obligations
is out of scope for ADR-0004.

## Responsibility model

Each responsibility below is described only at property level. Purpose,
inputs, outputs, latency sensitivity, bandwidth profile, ordering,
delivery, loss tolerance, confidentiality, integrity, authorization,
metadata exposed, correlation risk, local trust assumptions, failure
behavior, and deliberate deferrals are stated per responsibility. Where a
property is identical to one already stated for the plane as a whole (for
example: authority derives only from the handoff, denial by default), it
is not restated.

### Interactive screen updates

- **Purpose.** Deliver visual state of the remote endpoint to the initiator
  with interactive responsiveness.
- **Inputs.** Frame or region updates emitted by the remote endpoint's
  capture surface.
- **Outputs.** Displayable visual updates at the initiator.
- **Latency sensitivity.** High. Perceived interactivity depends on it.
- **Bandwidth profile.** Bursty, potentially sustained-high; content and
  activity dependent.
- **Ordering.** Later frames semantically supersede earlier ones for the
  same region; strict global ordering is not required.
- **Delivery.** Eventual visual convergence is required; individual frame
  delivery is not.
- **Loss tolerance.** Individual updates may be dropped or coalesced when
  superseded.
- **Confidentiality.** Required end-to-end across untrusted intermediaries.
- **Integrity.** Required; corrupted or forged updates must be rejected or
  cause safe failure.
- **Authorization.** Only permitted if the session capability set includes
  screen viewing for the negotiated purpose.
- **Metadata exposed.** Existence of a session, its rough duration, and
  aggregate traffic profile may be inferable to observers on the path
  (see Data Plane metadata and observer model).
- **Correlation risks.** Traffic-shape fingerprinting of application
  behavior; per-session activity patterns.
- **Local trust assumptions.** The remote OS capture surface is trusted
  only to produce frames of the intended surface at the intended scope.
- **Failure behavior.** Degrade to lower fidelity or pause; never expose
  unauthorized surfaces.
- **Deferred.** Codec, capture mechanism, region model, fidelity tiers,
  numeric thresholds.

### Keyboard input

- **Purpose.** Deliver initiator key events to the remote endpoint.
- **Inputs.** Local key events from the initiator OS input surface.
- **Outputs.** Injected key events at the remote endpoint.
- **Latency sensitivity.** Very high.
- **Bandwidth profile.** Low, bursty.
- **Ordering.** Strict per-session ordering required.
- **Delivery.** Reliable delivery required; lost events must not silently
  disappear.
- **Loss tolerance.** None for individual events within an active
  interaction.
- **Confidentiality.** Required; keystrokes may include secrets.
- **Integrity.** Required.
- **Authorization.** Permitted only if the capability set includes
  keyboard input for the negotiated purpose.
- **Metadata exposed.** Timing of keystrokes may be inferable as small
  low-latency packets to on-path observers.
- **Correlation risks.** Keystroke-timing fingerprints of typing behavior.
- **Local trust assumptions.** Remote OS input injection surface is trusted
  to deliver events only within the session's scope.
- **Failure behavior.** Fail closed; do not buffer and later replay events
  after a session boundary or capability change.
- **Deferred.** Input injection mechanism, per-platform quirks.

### Mouse and pointer input

- **Purpose.** Deliver pointer motion, buttons, wheel, and gesture events.
- **Latency sensitivity.** Very high.
- **Bandwidth profile.** Low to moderate; high during continuous motion.
- **Ordering.** Strict per-session ordering required.
- **Delivery.** Reliable; some motion samples may be coalesced when
  perceptibly equivalent.
- **Loss tolerance.** None for discrete events (button, wheel notch);
  motion samples may be down-sampled.
- **Confidentiality, integrity, authorization.** As for keyboard input,
  scoped to pointer capability.
- **Metadata exposed.** Continuous small packet cadence.
- **Correlation risks.** Motion cadence and dwell fingerprints.
- **Failure behavior.** Fail closed.
- **Deferred.** Pointer model, gesture handling, high-precision devices.

### Clipboard synchronization

- **Purpose.** Move clipboard content between initiator and remote endpoint
  when authorized.
- **Inputs.** Clipboard change events on either side.
- **Outputs.** Clipboard state updates on the other side.
- **Latency sensitivity.** Moderate.
- **Bandwidth profile.** Highly variable; individual transfers may be
  large.
- **Ordering.** Latest-wins per direction is acceptable; strict global
  ordering is not required.
- **Delivery.** Reliable; partial transfers must not be observable as
  clipboard state.
- **Loss tolerance.** Failed transfers must fail visibly, not silently
  corrupt clipboard state.
- **Confidentiality.** Required; clipboard often contains secrets.
- **Integrity.** Required.
- **Authorization.** Permitted only if the capability set includes
  clipboard sync, with direction (initiator→remote, remote→initiator, or
  both) bound to the granted capability.
- **Metadata exposed.** Transfer size and timing.
- **Correlation risks.** Distinctive per-user clipboard patterns; large
  transfers act as timing beacons.
- **Local trust assumptions.** OS clipboard surface at each end is trusted
  only for read/write of the currently-active clipboard state.
- **Failure behavior.** Fail closed; drop unauthorized directions
  silently at the enforcement point.
- **Deferred.** Clipboard formats, size limits, sanitization policy.

### Audio

- **Purpose.** Carry audio from the remote endpoint to the initiator, and
  optionally in the reverse direction when authorized.
- **Latency sensitivity.** High.
- **Bandwidth profile.** Sustained moderate.
- **Ordering.** Playback order required.
- **Delivery.** Best-effort with loss concealment; strict reliability not
  required.
- **Loss tolerance.** Small losses tolerable with concealment; large
  losses must degrade audibly rather than desynchronize.
- **Confidentiality, integrity.** Required.
- **Authorization.** Per direction, scoped by the granted capability.
- **Metadata exposed.** Continuous stream cadence.
- **Correlation risks.** Content-based cadence patterns; silence/activity
  timing.
- **Failure behavior.** Mute rather than expose unauthorized audio.
- **Deferred.** Codec, sample rates, echo/noise handling.

### Video (future)

- **Purpose.** Camera or auxiliary video streams when authorized.
- **Latency sensitivity.** High for interactive use, lower for playback.
- **Bandwidth profile.** Sustained high.
- **Other properties.** Follow the same shape as screen updates and audio,
  with per-capability directionality.
- **Deferred.** Codec, capture pipeline, resolution tiers, use cases.

### File transfer (future)

- **Purpose.** Move discrete files between the two endpoints when
  authorized.
- **Latency sensitivity.** Low.
- **Bandwidth profile.** Bulk.
- **Ordering.** Per-transfer ordering required; strict cross-transfer
  ordering not required.
- **Delivery.** Reliable; partial or corrupted transfers must be visibly
  failed.
- **Confidentiality, integrity.** Required.
- **Authorization.** Explicit capability including direction and, where
  meaningful, scope (path, quota).
- **Metadata exposed.** File sizes and transfer timing.
- **Correlation risks.** File-size fingerprints; per-user transfer
  patterns.
- **Failure behavior.** Fail visibly; never leave partial files as
  authoritative.
- **Deferred.** Transfer protocol, resumption, integrity mechanism,
  metadata policy.

### Session lifecycle enforcement

- **Purpose.** Start, sustain, and terminate the session in accordance with
  the handoff, capability set, expiry, revocation, and invalidation.
- **Inputs.** Handoff artifact; capability updates; revocation notices;
  invalidation notices; local health signals.
- **Outputs.** Session state transitions; bounded return-flow events (see
  below).
- **Authorization.** Lifecycle transitions authorized only by ADR-0003
  interfaces or by locally-observable safety conditions (for example loss
  of validation).
- **Failure behavior.** On uncertainty about authority, terminate the
  session and emit the corresponding bounded return-flow event.
- **Deferred.** Reconnect strategy, session identifier lifetimes, backoff.

### Capability enforcement

- **Purpose.** Ensure every operation is inside the granted capability set,
  purpose, and lifetime.
- **Properties.** See Capability enforcement section for required
  behavior. Enforcement is repeated at the interface contract, the local
  trust boundary between Data Plane and OS surfaces, and the verification
  criterion.

### Bounded return-flow reporting

- **Purpose.** Provide the Control Plane with the minimum lifecycle,
  capability, and health information required to manage sessions and
  revocation.
- **Properties.** See Return-flow lifecycle and metadata.

## Logical interfaces

Only property contracts are defined. No API, endpoint, schema, wire
format, language, framework, library, or deployment topology is defined.

### Control Plane → Data Plane handoff

- Carries only the ADR-0003 handoff artifact and its associated bounded
  parameters (session, purpose, capability set, lifetime, recipient
  binding, freshness, revocation reference, invalidation reference).
- Must be independently verifiable by the Data Plane (see Handoff
  validation).
- Must not carry stable identity across the seam, and must not carry
  material used for cross-session correlation.
- Must be single-session in intent; the interface must not allow a single
  artifact to authorize multiple sessions.
- Failure to validate must prevent any Data Plane work from starting.

### Data Plane → Control Plane return flow

- Carries only bounded lifecycle, capability, and health events (see
  Return-flow lifecycle and metadata).
- Must not carry session content, per-event fine-grained user activity,
  or metadata beyond what the Control Plane requires for its stated
  responsibilities.
- Must degrade safely: inability to deliver return-flow events must not
  cause the Data Plane to keep operating beyond its authority.

### Local client input → Data Plane

- Carries local OS input events destined for the session.
- Must be scoped to the active session's capability set; unauthorized
  event categories must be dropped at the boundary.
- Must not be persisted beyond the session.

### Data Plane → remote endpoint

- Carries authorized input and control events into the remote endpoint's
  input surfaces.
- Must be scoped by capability and purpose.
- Must be integrity-protected end-to-end across untrusted intermediaries.

### Remote endpoint → Data Plane

- Carries authorized capture output (screen, audio, future video, future
  file data).
- Must be scoped to the surfaces the session is authorized to observe.
- Must be integrity-protected end-to-end.

### Data Plane ↔ relay or intermediary infrastructure

- Treats all such infrastructure as untrusted for content confidentiality
  and integrity.
- May rely on infrastructure for reachability and delivery only.
- Must not depend on relay-side identity, relay-side authorization
  decisions, or relay-side long-term correlation.

### Data Plane ↔ future transport mechanisms

- Contract stated in terms of properties (confidentiality, integrity,
  authenticity, replay resistance, congestion behavior, reconnect
  semantics), not any specific protocol.
- Any future transport must be evaluatable against these properties.

### Data Plane ↔ future codec or media-processing components

- Contract stated in terms of properties (bounded metadata leakage,
  bounded side channels, integrity, failure containment). No codec is
  chosen.

### Data Plane ↔ operating-system capture and input surfaces

- Trusted only to expose exactly the surfaces the user's OS-level
  authorization has granted the running Data Plane component.
- Must not be assumed to enforce cross-application isolation beyond what
  the OS provides.
- Failures at this boundary must fail closed at the Data Plane level.

## Handoff validation

Before any session work begins, the Data Plane must independently
validate, as architectural requirements:

- Provenance of the handoff artifact (that it originates from the Control
  Plane authority responsible for this session).
- Integrity of the artifact and its bound parameters.
- Intended Data Plane recipient or endpoint role.
- Intended session.
- Negotiated purpose.
- Capability set.
- Freshness.
- Expiry.
- Explicit revocation status where applicable.
- Session-state invalidation.
- Recipient binding.
- Non-transferability or proof-of-possession where later threat analysis
  requires it.

The Data Plane does not repeat user authentication. It must not treat
possession of the artifact alone as sufficient transferable authority.

No validation mechanism is selected in this ADR. Later mechanism-selection
ADRs must justify the chosen mechanism against these properties.

## Capability enforcement

The Data Plane must, as required properties:

- Enforce only the capabilities granted for this session.
- Enforce only the negotiated purpose.
- Apply least privilege at every operation.
- Deny by default; unknown or unrecognized operation categories are
  denied.
- Honor mid-session revocation promptly.
- Honor expiry independently of explicit revocation.
- Honor session invalidation independently of expiry and revocation.
- Prohibit reuse of any handoff or capability artifact across sessions.
- Prohibit any capability escalation from Data Plane content or metadata,
  including from content received from the remote endpoint or from
  local input.
- Capabilities must be enforced independently by capability type.
- Granting one capability must not implicitly grant any other capability.
- At minimum, screen viewing, input injection, clipboard synchronization,
  audio, file transfer, and future video capabilities must be independently
  grantable, independently enforceable, and independently revocable during an
  active session.
- The Data Plane must fail closed when a requested capability has not been
  explicitly granted.

These are required behaviors, not implementation guidance. Mechanism
selection is deferred.

## Return-flow lifecycle and metadata

The return path receives full architectural treatment because it is where
the Data Plane's most sensitive residual metadata lives.

For each lifecycle event category below, the ADR states: why the Control
Plane needs it, minimum information required, who may observe it, timing
granularity, whether immediate reporting is actually required, whether
delay may be acceptable, whether batching may be acceptable, whether
normalization may be possible, whether omission may be possible, what
residual correlation remains, and what is deferred.

For brevity of prose (not of analysis) the properties are grouped where
they genuinely coincide.

### Session started

- **Why needed.** Control Plane must know an authorized session has begun
  to manage lifetime and revocation.
- **Minimum information.** Reference to the specific session; no user or
  device identity beyond what the Control Plane already holds.
- **Observers.** Control Plane; on-path observers may already infer
  session existence from traffic (see observer model).
- **Timing granularity.** Coarse acceptable.
- **Immediate required?** Not strictly; small delay acceptable.
- **Batching / normalization / omission.** Batching acceptable across
  bursts; normalization to coarse timing acceptable; omission not
  acceptable because revocation depends on knowing sessions exist.
- **Residual correlation.** Session-start timing may correlate with
  Control Plane authorization timing.
- **Deferred.** Timing granularity policy, transport for return flow.

### Session ended cleanly / session ended unexpectedly

- **Why needed.** Free resources; drive Control Plane state; support
  audit.
- **Minimum information.** Session reference and end classification
  (clean vs unexpected).
- **Immediate required?** Preferable but not strict; short delay
  acceptable.
- **Batching / normalization.** Normalization of exact end time to
  coarser granularity acceptable; batching across sessions acceptable if
  it does not delay revocation-relevant transitions.
- **Residual correlation.** Session duration is revealed; duration
  distribution can fingerprint use over time.
- **Deferred.** Duration granularity policy.

### Capability exhausted / capability denied

- **Why needed.** Control Plane must know when granted authority is used
  up or when denials indicate misconfiguration or attack.
- **Minimum information.** Session reference; capability class; event
  category. No per-operation user-content data.
- **Immediate required?** Denials of interest for abuse detection may be
  reported promptly; routine exhaustion may be delayed or aggregated.
- **Batching / normalization.** Aggregation across a session acceptable;
  cross-session aggregation must not enable per-user profiling.
- **Residual correlation.** Rare denial patterns may fingerprint
  behavior.

### Revocation received / revocation acknowledged / session invalidated

- **Why needed.** These are the primary safety events; the Control Plane
  must know that its revocation and invalidation reached effect.
- **Minimum information.** Session reference; category (revocation
  received, acknowledgement, invalidation).
- **Immediate required?** Yes for acknowledgement, to close the safety
  loop.
- **Batching / normalization / omission.** Not acceptable to omit;
  minimal normalization only.
- **Residual correlation.** Timing of these events may correlate with
  Control Plane actions by design; this is intentional.

### Transport degraded / relay changed / reconnect attempted / recovery succeeded or failed

- **Why needed.** Health signal; needed to distinguish user-visible
  outages from silent capability failure.
- **Minimum information.** Session reference; category; coarse outcome.
- **Immediate required?** Not strict; short delay acceptable.
- **Batching / normalization.** Batching and normalization of exact
  timing acceptable and encouraged.
- **Residual correlation.** Reconnect and relay-change patterns can
  reveal network conditions of the user; over many sessions this can
  fingerprint a user's network environment.
- **Deferred.** Batching windows, normalization granularity, whether
  some categories can be omitted entirely.

### Cross-cutting privacy analysis of return flow

The following are stated as required analyses, not solutions:

- Individual events can reveal session behavior; the Control Plane must
  receive only the minimum needed.
- Event timing can link Control Plane and Data Plane activity; timing
  granularity is a privacy parameter, not an implementation detail.
- Rare event combinations can fingerprint a session; the return-flow
  vocabulary should not proliferate.
- The distribution of events across many sessions can become a behavioral
  fingerprint associated with one user or device; long-term return-flow
  storage must be designed with this in mind.
- Return-flow storage can create cross-session linkage; storage lifetime
  and aggregation policy require dedicated later analysis.

No padding, batching, delay, or normalization mechanism is selected here.
Required privacy properties and deferred evidence needs are stated; the
mechanism-selection ADR that eventually chooses these must justify them
against measured observer models.

## Local trust-boundary analysis

Only boundaries directly touched by the Data Plane. This is not the
system-wide trust-boundary ADR.

For each boundary: what is trusted, what is not trusted, data crossing,
authority crossing, secrets or session artifacts involved, metadata
visible, failure consequences, required protections, unknowns, deferred
decisions.

### Control Plane ↔ Data Plane

- **Trusted.** Control Plane to issue authority under ADR-0003.
- **Not trusted.** Control Plane to police Data Plane operations after
  handoff; anything not carried by the handoff.
- **Data crossing.** Handoff artifact; bounded return-flow events.
- **Authority crossing.** Capability grant, lifetime, revocation,
  invalidation.
- **Secrets / artifacts.** Handoff artifact and its bound parameters.
- **Metadata visible.** As defined by ADR-0003.
- **Failure.** Loss of correct authority; the Data Plane must fail
  closed.
- **Required protections.** Independent validation; least-privilege
  consumption.
- **Unknowns / deferred.** Validation mechanism; return-flow transport.

### Local device/OS ↔ Data Plane

- **Trusted.** OS to enforce process isolation and to expose only
  user-authorized capture/input surfaces.
- **Not trusted.** Other processes; unauthorized surfaces.
- **Data crossing.** Local input events, capture data, clipboard,
  audio/video.
- **Authority crossing.** OS-level permissions distinct from session
  capabilities.
- **Secrets / artifacts.** In-memory session material.
- **Failure.** Leakage into other processes; capture of unauthorized
  surfaces.
- **Required protections.** Minimize in-memory secret lifetime; refuse
  to operate on surfaces outside session scope.
- **Deferred.** Platform-specific mechanisms.

### Remote device/OS ↔ Data Plane

- Same structure as local; mirror obligations.
- **Deferred.** Remote-side platform mechanisms.

### Data Plane ↔ transport

- **Trusted.** Transport to move bytes with the properties negotiated by
  the transport contract.
- **Not trusted.** Transport to preserve confidentiality, integrity, or
  identity on its own.
- **Data crossing.** All session content and control events.
- **Failure.** Content interception, tampering, replay attempts.
- **Required protections.** End-to-end confidentiality and integrity
  above the transport; replay resistance; downgrade resistance.
- **Deferred.** Transport protocol; primitives.

### Data Plane ↔ relay or intermediary

- **Trusted.** Only for reachability and delivery.
- **Not trusted.** For content confidentiality, integrity, identity,
  authorization, or long-term correlation.
- **Failure.** Traffic-analysis linkage; denial of service.
- **Required protections.** No dependence on relay identity;
  minimization of visible metadata; no relay-based capability decisions.
- **Deferred.** Relay architecture.

### Data Plane ↔ codec/media-processing component

- **Trusted.** To transform authorized media without generating
  side-channel leakage beyond documented behavior.
- **Not trusted.** To make authorization decisions or to produce
  identity-carrying metadata.
- **Failure.** Media-format side channels; embedded metadata leakage.
- **Required protections.** Bounded interface; metadata stripping.
- **Deferred.** Codec selection.

### Data Plane ↔ clipboard

- **Trusted.** OS clipboard for currently-active content only.
- **Not trusted.** Clipboard history; other apps' clipboard content.
- **Failure.** Cross-app leakage; unauthorized-direction transfer.
- **Required protections.** Per-direction capability enforcement;
  refusal to persist clipboard history.
- **Deferred.** Format handling.

### Data Plane ↔ audio/video capture

- **Trusted.** OS capture surface for the specific device and permission
  granted.
- **Not trusted.** To scope beyond user-granted permissions.
- **Failure.** Capture of unauthorized sources.
- **Required protections.** Fail closed; explicit per-capability scope.
- **Deferred.** Capture implementation.

### Data Plane ↔ future file-transfer component

- **Trusted.** To operate only within granted capability, path, and
  direction.
- **Not trusted.** To decide policy.
- **Failure.** Unauthorized read/write; partial-write leakage.
- **Required protections.** Capability enforcement; visible failure of
  partial transfers.
- **Deferred.** Transfer protocol.

## Data Plane metadata and observer model

The following observer categories are identified without assuming a
specific transport. Later mechanism-selection ADRs must define the
concrete observer model and measure it.

For each observer, what may be observable in principle includes some
combination of: session existence, timing, duration, direction, volume,
endpoint information, relay changes, reconnect behavior, capability use,
and lifecycle events. This ADR does not claim any observer sees
everything, and does not claim any observer sees nothing.

- **Local network observer.** May observe existence, timing, duration,
  volume, and possibly endpoint reachability from the initiator side.
- **Remote network observer.** Symmetric on the remote side.
- **Relay operator.** May observe existence, timing, duration, direction,
  volume, and relay-change patterns; must not be relied on for content
  confidentiality or identity.
- **Transport provider.** May observe metadata typical of the transport;
  the exact set depends on the future transport selection.
- **Infrastructure operator.** May observe deployment-shape metadata;
  depends on future deployment decisions.
- **Compromised endpoint.** May observe all content that endpoint's role
  is authorized to see; capability scoping limits blast radius.
- **Data Plane operator.** May observe operational health and bounded
  return-flow events; must not be relied on for content access.
- **Colluding infrastructure parties.** May combine metadata across
  points to increase linkability; assumed possible and evaluated in
  later ADRs.

Universal observability and universal non-correlation are both rejected
as claims. Later mechanism-selection ADRs must define and measure the
actual observer model, chosen mitigations, and residual exposure.

## Required security and privacy properties

Stated as requirements, not achievements. Mechanism selection is
deferred; no cryptographic primitive is chosen.

- Confidentiality of session content end-to-end across untrusted
  intermediaries.
- Integrity of session content and control events end-to-end.
- Peer or endpoint authenticity as required by the session type.
- Secure handoff validation (as defined in Handoff validation).
- Authorization enforcement (as defined in Capability enforcement).
- Replay resistance across all session artifacts and events.
- Revocation handling.
- Session invalidation handling.
- Downgrade resistance across negotiated properties.
- Cryptographic agility: no assumption that any primitive is permanent.
- Evidence-supported post-quantum posture: any post-quantum claim must
  be backed by later mechanism evidence, not asserted here.
- Secure failure behavior: uncertainty about authority terminates the
  session.
- Bounded authority; least privilege.
- Memory and secret minimization.
- No unnecessary persistence of session material.
- No custom cryptography without separate justification in a dedicated
  ADR.

Privacy properties additionally include: minimization of return-flow
metadata; prohibition of promoting transient Data Plane metadata into
long-lived identity or account state; and explicit acknowledgement that
some metadata (session existence, coarse timing, coarse volume) is
observable and requires later mitigation analysis.

## Performance requirements

All numeric thresholds below are labeled as architectural estimates,
provisional requirements, or future measurement targets. No unlabeled
numeric threshold is asserted.

- **Interactive latency.** Provisional requirement: perceived
  round-trip latency must be low enough for continuous mouse and
  keyboard interaction; specific numeric thresholds are future
  measurement targets tied to the chosen transport.
- **Jitter.** Provisional requirement: jitter must be bounded such that
  input and pointer feel is not perceptibly disrupted under normal
  network conditions; numeric bounds are future measurement targets.
- **Responsiveness.** Architectural estimate: input events must be
  delivered with priority over bulk streams.
- **Throughput.** Provisional requirement: sufficient for interactive
  screen updates at usable fidelity; specific throughputs are future
  measurement targets.
- **Loss recovery.** Provisional requirement: loss recovery must not
  reorder or corrupt reliable event streams; mechanism deferred.
- **Ordering.** Requirement: strict per-stream ordering where declared
  (input); relaxed where declared (screen frames).
- **Congestion handling.** Requirement: the Data Plane must degrade
  gracefully under congestion rather than collapse.
- **Relay use.** Architectural estimate: relays may be traversed and
  must not be assumed to improve or preserve latency.
- **Reconnect behavior.** Requirement: reconnect must not silently
  extend authority beyond expiry or revocation.
- **Graceful degradation.** Requirement: fidelity, frame rate, and
  optional streams (audio, video, file transfer) may be degraded or
  paused to preserve interactivity of primary input and screen.

## Privacy / performance tension

Low latency, continuous high-volume traffic, relay use, batching,
padding, cover traffic, transport replacement, and usability are in
genuine tension. This ADR does not resolve the tension and does not
claim a universal answer.

- **Required properties.** Confidentiality; integrity; capability
  enforcement; revocation; minimized return flow; no direct or derivable
  stable identity across the seam.
- **Expected exposures.** Session existence, coarse timing, coarse
  duration, and coarse volume are observable to at least some observer
  categories in most plausible transports; per-session activity shape
  may be inferable.
- **Unmeasured facts.** The concrete observability set under any
  specific transport, relay architecture, or padding scheme has not been
  measured. The magnitude of behavioral fingerprinting over many
  sessions has not been measured. The user-visible cost of privacy
  mitigations (batching, delay, padding, cover traffic) on interactivity
  has not been measured.
- **Trade-offs requiring later evidence.** Choice of transport;
  presence/shape of padding or cover traffic; batching windows for
  return-flow events; relay topology; congestion strategy.

The Data Plane is neither fully unlinkable nor unavoidably correlated;
which side of that line it lands on depends on later mechanism selection
and measurement.

## Conflicts and proof-of-concept routing

At the time of drafting, no ADR-0003 obligation is judged architecturally
infeasible for the Data Plane. The following are recorded as areas of
tension that may become conflicts under later analysis and, if so, must
be routed to bounded proofs of concept rather than used to supersede
ADR-0003 by opinion:

1. **Non-transferability vs. transport pragmatism.** Some transport
   choices make non-transferability harder to enforce end-to-end. If a
   preferred transport cannot support proof-of-possession semantics for
   the handoff artifact, that is a documented conflict; the resolution
   is a proof-of-concept exercising proof-of-possession in that
   transport, not silent adoption of bearer semantics. Evidence that
   would resolve it: a working, measured non-transferable handoff in the
   candidate transport, or a documented demonstration that a different
   transport is required.
2. **Return-flow minimization vs. operational needs.** Some operational
   or safety needs may push toward richer return-flow content. If a
   specific operational need cannot be met with a minimized return-flow
   vocabulary, that is a documented conflict; the resolution is a
   proof-of-concept showing the minimum viable event set for that need,
   not expansion by default. Evidence: measured operational adequacy of
   a minimized vocabulary.
3. **Prohibition of derivable stable identity vs. session routing.**
   Routing sessions across relays without any stable-identity-derivable
   metadata is not obviously trivial. If a candidate mechanism cannot
   avoid derivable stable identity, that is a documented conflict; the
   resolution is a proof-of-concept of an ephemeral routing model.
   Evidence: measured routing feasibility under ephemeral-only
   identifiers.

An unresolved conflict does not automatically block ADR-0004 acceptance,
but does block acceptance of any later mechanism-selection ADR that
depends on that conflict being resolved.

## Deferred decisions

Deferred, and to be handled by later ADRs backed by product evidence.
This list is sufficiently explicit for the current bounded question but
is not claimed to be exhaustive:

- Transport selection.
- cMixx use or non-use.
- xxDK use or non-use.
- QUIC, TLS, Noise, WireGuard, WebRTC, WebSocket, or any other protocol
  selection.
- Screen-capture mechanism.
- Input-injection mechanism.
- Codec selection.
- Media pipeline.
- Clipboard format.
- Audio and video implementation.
- File-transfer protocol.
- Relay architecture.
- Congestion control.
- Numeric latency targets.
- Numeric throughput targets.
- Token or capability format.
- Cryptographic primitives.
- Post-quantum mechanism.
- Programming language.
- Framework.
- Database.
- Storage.
- Deployment.
- Platform-specific implementation.
- Browser support mechanism.
- Mobile support mechanism.

Newly discovered decisions must be added through later ADRs.

## Evidence discipline

- The architectural assignments in this ADR are current design
  judgments.
- Performance and metadata assessments are not measured operational
  evidence.
- No transport, privacy, unlinkability, security, latency, post-quantum,
  or reliability property is proven by this ADR.
- Later mechanism-selection ADRs must justify or revise these judgments
  with product evidence.
- Unresolved feasibility conflicts may be routed to bounded proofs of
  concept.
- An accepted ADR is superseded only when evidence justifies it, not by
  architectural opinion alone.
- Structural acceptance of ADR-0004 establishes an architectural
  baseline for Data Plane work; it does not prove feasibility or
  optimality of any specific mechanism.

## Consequences

- The Data Plane has a bounded, mechanism-neutral shape that can be
  reasoned about independently of any transport, codec, or primitive.
- ADR-0003's obligations are anchored on the Data Plane side, so the
  Control Plane's guarantees are not silently voided by later Data Plane
  choices.
- Mechanism-selection ADRs (transport, codec, cryptography, relay,
  platform) now have a specified target to conform to and be measured
  against.
- Some responsibilities (video, file transfer) are named as future
  work; the architecture reserves room for them without committing to
  their design today.
- Some tensions (privacy vs. performance; non-transferability vs.
  transport pragmatism) are named but not resolved; later work must
  resolve them with evidence.

## Alternatives considered

- **Fold the Data Plane into the Control Plane.** Rejected: violates
  ADR-0002 plane separation and dissolves the trust boundary that
  ADR-0003 depends on.
- **Assume the transport provides all confidentiality, integrity, and
  authenticity.** Rejected: transports change, are sometimes hostile,
  and require the Data Plane to have end-to-end properties independent
  of them.
- **Trust relays for identity or authorization.** Rejected: contradicts
  ADR-0003's trust assumptions.
- **Design return-flow richness for operational convenience.** Rejected:
  contradicts minimization; if operational needs cannot be met by a
  minimized vocabulary, that is routed to proof-of-concept, not
  expansion by default.
- **Pick a transport now to make analysis concrete.** Rejected:
  ADR-0004's bounded question explicitly excludes mechanism selection.
  Choosing a transport here would foreclose measurement-driven choice
  later.

## Non-goals

- Not the system-wide trust-boundary ADR.
- Not a mechanism-selection ADR for any transport, codec, primitive,
  token, language, framework, or deployment.
- Not a redesign of the Control Plane.
- Not a renegotiation of ADR-0003's handoff contract.
- Not a proof of feasibility, optimality, latency, unlinkability, or
  post-quantum posture.
- Not an implementation.

## Verification

Structural verification of this ADR requires that:

- Every named Data Plane responsibility is covered with property-level
  content.
- ADR-0003 obligations are carried as conformance requirements, not
  softened.
- No seam contract from ADR-0003 is silently weakened.
- Handoff validation duties are explicit and independent.
- Capability enforcement is defined as required properties, including
  deny-by-default and prohibition of escalation from Data Plane
  content or metadata.
- Return-flow metadata is analyzed for lifecycle needs, timing,
  batching, normalization, omission, residual correlation, and
  cross-session fingerprinting.
- Cross-session event-pattern fingerprinting is explicitly addressed.
- Local trust boundaries touched by the Data Plane are documented.
- Observer categories are explicit and neither claim universal
  observability nor universal non-correlation.
- Privacy/performance tensions are stated honestly.
- Conflicts with ADR-0003 are documented and routed to proof-of-concept
  rather than used to supersede accepted ADRs by opinion.
- No transport, codec, cryptographic primitive, token format, language,
  framework, or deployment is selected.
- Deferred decisions are explicit and not claimed to be exhaustive.
- Architectural judgments are not presented as measured facts.
- Structural acceptance is not treated as proof of feasibility or
  optimality.
- Rollback is possible through a superseding ADR backed by evidence.

## Rollback

ADR-0004 can be superseded by a later ADR that:

- Identifies a specific Data Plane requirement or responsibility whose
  architectural framing is judged incorrect on the basis of product
  evidence.
- States the replacement framing.
- Explicitly addresses the ADR-0003 obligations affected.
- Preserves or explicitly retires any prior commitments.
- Is subject to the same four-gate review.

No partial in-place edit of this ADR is a valid rollback path once
Accepted; the mechanism is a superseding ADR.
