# ADR-0003: Control Plane architecture

- Status: Accepted
- Type: Operational
- Expires: Does not expire
- Date: 2026-07-14
- Foundry generation: 1
- Product: Daang Remote
- Supersedes: none
- Superseded by: none
- Related: ADR-0002 (Control Plane / Data Plane Separation); Daang Remote product charter

## Context

ADR-0002 established that Daang Remote separates a Control Plane from a Data
Plane and defined the separation invariants that any concrete architecture must
preserve. ADR-0002 did not, and deliberately could not, define what the Control
Plane actually is: what it owns, what it exposes, what it must hand across the
plane seam, and what it is trusted (or not trusted) to do.

Before any mechanism can be evaluated — transports, identity systems, credential
formats, cryptographic primitives, relay topologies, session bindings — the
Control Plane must exist as a bounded architectural object with named
responsibilities, logical interfaces expressed as property contracts, an
explicit plane-crossing handoff shape, and a stated local trust model. Without
that object, later mechanism-selection ADRs have nothing to be evaluated
against and cannot produce independently verifiable evidence.

This ADR defines only that architectural object. It answers exactly one bounded
question:

> What responsibilities, interfaces, required properties, and local trust
> assumptions define the Daang Remote Control Plane, without selecting its
> transport, implementation stack, identity mechanism, or cryptographic
> protocol?

It does not design the product, does not design the Data Plane, and does not
choose any implementation mechanism. Every architectural assignment below is a
current design judgment or a stated requirement, not measured operational
evidence.

## Decision

Adopt a Control Plane defined by:

1. A fixed set of responsibilities the Control Plane owns end-to-end.
2. Logical interfaces expressed only as property contracts, not wire formats.
3. A minimized plane-crossing handoff whose required properties are stated
   without selecting a token, primitive, protocol, or storage mechanism.
4. Bidirectional trust assumptions between the Control Plane, the Data Plane,
   relays, and future identity/transport services.
5. A local trust-boundary analysis limited to boundaries the Control Plane
   directly touches.
6. Required architectural properties the Control Plane must satisfy or
   demonstrably approximate, distinct from claims of achievement.
7. An explicit deferred-decision list enumerating every mechanism choice that
   this ADR does not make and that later ADRs must make with product evidence.

This ADR makes ADR-0002 separation invariant 2 concrete by prohibiting
cross-plane re-linking through identity-bearing or correlatable material.
ADR-0002 separation invariant 3 separately requires every shared component,
including identity, key material, relay pools, and session-binding state, to be
analyzed as a correlation surface in the ADR that introduces the sharing.
ADR-0003 adds its own handoff requirements: the handoff must be minimized,
single-session, single-purpose, short-lived, scoped, and resistant to replay,
without selecting a mechanism.

## Responsibility model

The Control Plane owns the following responsibilities. Each responsibility
lists purpose, inputs, outputs, required properties, local trust assumptions,
metadata handled, identifiers handled, and what is deliberately deferred.

### 1. Discovery

- **Purpose.** Allow a client to locate a set of Control Plane endpoints
  sufficient to attempt session establishment.
- **Inputs.** A minimal client-side bootstrap reference (its exact form is
  deferred) and the current reachability state visible to the client.
- **Outputs.** A candidate set of Control Plane contact points usable for
  rendezvous, together with any properties required to evaluate them (e.g.
  freshness, minimum required guarantees).
- **Required properties.** Discovery must not require the client to disclose
  long-lived identity, must tolerate partial reachability, and must not embed
  session-specific identifiers in the discovery step.
- **Local trust assumptions.** The bootstrap reference source is trusted to
  supply Control Plane contact information but is not trusted with session
  content.
- **Metadata handled.** The fact that a client is attempting to discover
  Control Plane endpoints, plus network-observable characteristics inherent
  to the chosen transport (deferred).
- **Identifiers handled.** None specific to the user, account, device, or
  session at this stage.
- **Deferred.** Transport, addressing scheme, directory model, freshness
  mechanism, anti-enumeration protections.

### 2. Rendezvous

- **Purpose.** Bring two parties (or a client and infrastructure) into a
  shared context sufficient to begin session negotiation.
- **Inputs.** Discovery output; any rendezvous coordinates required by the
  peer relationship (form deferred).
- **Outputs.** A shared ephemeral rendezvous context suitable for negotiation.
- **Required properties.** Rendezvous coordinates must not require or expose
  long-lived identifiers; the rendezvous context must be single-use.
- **Local trust assumptions.** Rendezvous infrastructure may be relied upon
  for liveness and coordination, but must be treated as potentially capable of
  observing, retaining, or correlating rendezvous events. The architecture
  must minimize what it can observe and must not rely on operator promises to
  prevent linkage.
- **Metadata handled.** Timing of rendezvous attempts and any coordinates
  required to complete them.
- **Identifiers handled.** Only ephemeral rendezvous coordinates.
- **Deferred.** Coordinate format, discoverability model, anti-abuse
  mechanisms, relay selection; observer model, retention constraints,
  cross-session correlation controls.

### 3. Session negotiation

- **Purpose.** Establish the parameters that define a single session,
  including its purpose, its participants (by role, not by identity), and its
  bounds.
- **Inputs.** Rendezvous context; capability declarations from participants;
  applicable policy inputs.
- **Outputs.** A negotiated session context sufficient for handoff to the
  Data Plane.
- **Required properties.** Negotiation must be bounded to a single session,
  must not embed long-lived identifiers in the resulting context, and must
  fail closed if required properties cannot be established. No authenticated
  principal automatically receives session or Data Plane authority. Authority
  exists only after separate authorization and capability negotiation.
- **Local trust assumptions.** Participants are trusted only to the extent
  established by authentication; the negotiation channel is trusted for
  confidentiality and integrity as required properties (mechanism deferred).
- **Metadata handled.** Session purpose, requested capabilities, negotiated
  policy.
- **Identifiers handled.** Session-scoped, ephemeral role identifiers only.
- **Deferred.** Negotiation protocol, capability schema, policy language.

### 4. Authentication negotiation

- **Purpose.** Agree on and execute the authentication method(s) required for
  this session, at the level of assurance appropriate to its purpose.
- **Inputs.** Session context; available authentication factors and
  credentials (form deferred).
- **Outputs.** An authentication result usable by session negotiation and
  usable to derive session-scoped authority — not a long-lived credential
  usable outside the session.
- **Required properties.** Mutual authentication where the session requires
  it; freshness; resistance to replay; no leakage of long-lived credential
  material into session context. Authentication establishes claims and an
  assurance level. It does not itself grant Data Plane authority.
  Authorization is a separate decision that maps those authenticated claims to
  a bounded, session-scoped capability.
- **Local trust assumptions.** The authentication verifier is trusted to
  enforce the negotiated method; the credential source is trusted only for
  the authenticated party's claims.
- **Metadata handled.** Authentication method used, assurance level, outcome.
- **Identifiers handled.** Authenticated principal reference (its form is
  deferred) used only within the Control Plane.
- **Deferred.** Authentication factors, credential formats, verifier
  implementation, assurance model.

### 5. Identity exchange

- **Purpose.** Exchange the minimum identity assertions required for the
  session — role, capability, or attribute — without exporting long-lived
  identifiers into the session context or across the plane seam.
- **Inputs.** Authentication result; policy requirements.
- **Outputs.** Session-scoped identity assertions suitable for use inside the
  Control Plane and for deriving handoff authority, not for correlation.
- **Required properties.** Identity exchange must produce session-scoped
  artifacts, must not export account, device, or long-lived route identifiers
  into the Data Plane, and must be revocable.
- **Local trust assumptions.** The identity source is trusted for the claims
  it makes at the assurance level agreed during authentication.
- **Metadata handled.** Claims required for authorization and capability
  negotiation.
- **Identifiers handled.** Session-scoped identity references only.
- **Deferred.** Identity provider, claim schema, credential format, revocation
  channel.

### 6. Relay coordination

- **Purpose.** Where the Control Plane participates in selecting or arranging
  relays, coordinate that selection so that the resulting session can proceed
  under the negotiated properties.
- **Inputs.** Session context; available relay information (form deferred).
- **Outputs.** Relay arrangement inputs necessary for the Data Plane to
  operate the session, expressed only as capabilities, not as durable
  identifiers.
- **Required properties.** Relay coordination must not require the Control
  Plane to hold a stable mapping between users and long-lived relay
  identities; selection may be reproducible under audit only where the audit
  evidence is event-scoped, purpose-bounded, minimally retained, separated from
  long-lived user identity, and reviewed for its own cross-session linkage
  risk. Auditability must not create a durable user-to-relay or
  session-to-session correlation store.
- **Local trust assumptions.** Relay infrastructure is trusted for transport
  duties but is not trusted with session content or identity.
- **Metadata handled.** Relay capability and eligibility information.
- **Identifiers handled.** Ephemeral, session-scoped relay references.
- **Deferred.** Relay system, selection algorithm, reputation model,
  reachability probing.

### 7. Metadata protection

- **Purpose.** Ensure that metadata handled by the Control Plane is minimized
  at collection, minimized in transmission across interfaces, and minimized
  in what is passed across the plane seam.
- **Inputs.** All metadata produced by other responsibilities.
- **Outputs.** Minimized metadata records, minimized cross-interface
  disclosures, minimized handoff payload.
- **Required properties.** Metadata minimization, resistance to unnecessary
  correlation, unlinkability where feasible, and explicit disclosure of
  unavoidable linkability. This ADR does not claim any of these are achieved;
  they are required and later mechanism ADRs must supply evidence.
- **Local trust assumptions.** Every component that receives Control Plane
  metadata is treated as capable of misuse until scoped by a policy.
- **Metadata handled.** All metadata produced elsewhere in this list.
- **Identifiers handled.** All identifiers produced elsewhere in this list,
  reviewed against minimization.
- **Deferred.** Retention model, logging model, observability strategy, redaction
  and aggregation methods.

### 8. Capability negotiation

- **Purpose.** Agree on the set of session capabilities that will be granted,
  bounded to this session's purpose and lifetime.
- **Inputs.** Session context; authenticated identity claims; policy inputs.
- **Outputs.** A capability set expressed as required properties, not as
  transport-level entitlements.
- **Required properties.** Session-scoped authority; least-privilege
  information flow; explicit denial of capabilities not required for the
  session; ability to revoke capability mid-session.
- **Local trust assumptions.** Participants receive only capabilities agreed
  during negotiation; capability enforcement is a shared responsibility
  between Control Plane and Data Plane, as stated below.
- **Metadata handled.** Requested and granted capabilities.
- **Identifiers handled.** Session-scoped capability references only.
- **Deferred.** Capability language, enforcement protocol, revocation channel.

## Logical interfaces

Each interface below is defined only as a contract expressed in required
properties. No wire formats, endpoints, frameworks, APIs, programming
languages, or libraries are selected.

### Client ↔ Control Plane

- **Contract.** The client presents a bounded request for a session of a
  stated purpose; the Control Plane responds with either a negotiated session
  context (leading to handoff) or a refusal.
- **Required properties.** Mutual authentication where the session requires
  it; confidentiality and integrity of the exchange; freshness; no export of
  long-lived Control Plane identifiers into client-visible state beyond what
  the session requires; failure closed on unmet properties.
- **Metadata surface.** Only the metadata required for the session; the
  contract must not require the client to disclose more.

### Control Plane ↔ Data Plane

- **Contract.** The Control Plane hands to the Data Plane a single-session,
  single-purpose, short-lived session context sufficient to operate the
  session, and nothing more. The Data Plane returns only minimum session
  state and does not push data back that would re-link the session to
  Control Plane identity.
- **Required properties.** Session-scoped authority; replay resistance;
  revocation representable; no cross-session correlatable identifiers; no
  long-lived identity, account, device, route, or relay identifiers on the
  handoff.
- **Metadata surface.** The minimized handoff payload defined below.

### Control Plane ↔ relay infrastructure

- **Contract.** The Control Plane conveys to relay infrastructure only the
  capability and coordination information necessary for the relay to perform
  its transport role for this session.
- **Required properties.** No stable Control-Plane-to-relay identity map for
  end users; ephemeral, session-scoped relay references; ability to withdraw
  relay authority; explicit disclosure of any linkability that is unavoidable
  given the deferred transport choice.
- **Metadata surface.** Relay capability and eligibility inputs only.

### Control Plane ↔ future identity/key services

- **Contract.** The Control Plane consumes identity assertions and key
  material from external identity/key services strictly for the duration and
  purpose of authentication, identity exchange, and session-scoped authority
  derivation. It does not export session context back to those services in a
  form that would allow them to correlate sessions.
- **Required properties.** Assurance level must be statable and independently
  verifiable where feasible; trust must be scoped per claim and assurance
  level; requests and returned claims must be least-privilege; long-lived
  private key material must not be exported unless a later ADR proves it
  necessary and safe; compromise, revocation, and recovery behavior must be
  explicit; cryptographic agility must be preserved; and the service must not
  be treated as privacy-preserving merely because it is trusted for identity
  assertions.
- **Metadata surface.** Only claims required for authorization and
  capability negotiation.

### Control Plane ↔ future transport mechanisms

- **Contract.** The Control Plane treats transport as a replaceable
  substrate. Any transport must be able to carry Control Plane interfaces
  while preserving the required properties above.
- **Required properties.** Transport replaceability; the Control Plane may
  depend only on explicitly declared transport-contract properties required by
  the interface, such as confidentiality, integrity, peer authentication,
  delivery, ordering, freshness, or availability behavior. It must not depend
  on undocumented, vendor-specific, or accidental behavior of a particular
  transport; the Control Plane must not embed transport-specific identifiers
  in session context.
- **Metadata surface.** Whatever the deferred transport unavoidably exposes;
  such exposure must be documented, not silently accepted, by any later
  mechanism-selection ADR.

## Plane-crossing handoff requirements

This section makes ADR-0002 invariant 2 concrete at the handoff seam and
applies invariant 3 by treating every shared session artifact, key reference,
relay reference, and binding input as a correlation surface. The minimized,
single-session, single-purpose, short-lived handoff requirements below are new
ADR-0003 requirements.

### Minimum handoff payload

The Control Plane may pass to the Data Plane only the minimum information
required for the Data Plane to operate exactly one session for exactly one
purpose for a bounded lifetime. The strictly necessary fields, expressed as
properties rather than formats, are:

- A session-scoped authorization artifact proving that the session was
  negotiated and granting only the bounded authority required to operate it.
  Possession alone must not automatically create transferable bearer authority;
  the artifact must support verifiable binding to the intended session,
  negotiated purpose, capability set, and intended Data Plane recipient or
  endpoint role.
- The negotiated session purpose, bounded and enumerable.
- The negotiated capability set for this session.
- A bounded session lifetime.
- Freshness and replay-resistance inputs sufficient for the later selected
  mechanism to detect and reject replay. A nonce may be one input, but this ADR
  does not claim that a nonce alone is sufficient.
- The minimum coordination inputs the Data Plane requires to actually run
  the session (their content is deferred to the Data Plane ADR and to
  transport-selection ADRs, but must remain session-scoped).
- An opaque, ephemeral, session-scoped target or connection capability where
  required for the Data Plane to reach the authorized session endpoint. It must
  not identify a durable user, account, device, route, or relay; must not be
  reusable across sessions; must reveal only what is required to establish
  this session; and must not be resolvable by the Data Plane into long-lived
  Control Plane identity.

Every permitted handoff field is metadata, including:

- session purpose;
- capability set;
- lifetime;
- freshness inputs;
- target or connection capability shape;
- relay or coordination inputs;
- revocation state;
- and return-flow events.

Later mechanism-selection ADRs must evaluate whether these fields require
minimization, coarsening, normalization, padding, batching, omission, or
explicit justification. This ADR does not mandate a mitigation.

### Prohibited on the handoff

The following identifiers, references, or derivations must not cross the
plane seam under any circumstance:

- account identifiers;
- user identifiers;
- long-lived or durable device identifiers;
- long-lived route or relay identifiers, except opaque ephemeral
  session-scoped target or connection capabilities satisfying the requirements
  above;
- authentication credentials or long-lived credential material;
- persistent identity claims beyond those required for this session's
  authority;
- any Control-Plane-internal correlation key.

Derivations that would allow reconstruction of the above from handoff
content are equally prohibited.

### Required properties of the handoff artifact

The handoff token or session capability, regardless of mechanism, must:

- bind to exactly one session;
- bind to exactly one negotiated purpose;
- carry a bounded lifetime after which it is inert;
- provide replay resistance as a required property through validated
  freshness, scoped binding, state, or equivalent mechanism-specific controls;
  this ADR does not declare any one input sufficient;
- be revocable as a required property, whether by expiry, explicit
  revocation, or session-context invalidation;
- carry no long-lived identifier and no derivation that yields one;
- contain no stable, user-specific, account-specific, device-specific, or
  cross-session distinguishing feature beyond session-scoped attributes that
  are strictly necessary for the negotiated purpose, capability set, lifetime,
  and connection establishment;
- be verifiably bound to the intended Data Plane recipient, instance, endpoint
  role, or equivalent scoped recipient property;
- support proof-of-possession or an equivalent non-transferability property
  where later threat analysis requires it;
- not become usable merely because it was copied, logged, intercepted, or
  delivered to the wrong recipient.

### Bounding to one session and one purpose

The handoff is bounded to a single session by binding it to a session-scoped
context that ceases to be valid outside that session; to a single purpose by
including the negotiated purpose in the artifact's authority so that
Data-Plane use for a different purpose is rejected; and to a limited lifetime
by an explicit expiry that both the Control Plane and the Data Plane treat as
authoritative. Extension of the lifetime, if permitted at all, must be a new
negotiation, not a refresh of the same artifact.

### Replay prevention and revocation as required properties

Replay prevention is a required property of the artifact and of the
Control-Plane-to-Data-Plane interface. The mechanism is deferred; the
property is not.

Expiry, explicit revocation, and session-state invalidation are distinct
events:

- expiry is scheduled termination at the end of the bounded lifetime;
- explicit revocation is an active withdrawal of authority before expiry;
- session-state invalidation occurs when negotiated state changes or becomes
  invalid.

Later mechanism-selection ADRs must define propagation, freshness,
acknowledgement, race handling, unavailable-status behavior, and fail-closed
semantics for explicit revocation and invalidation. The specific channel and
mechanism remain deferred.

### Preventing possession from enabling correlation

The handoff artifact must contain no direct or derivable stable identity and
must not introduce additional avoidable linkage between Data Plane activity and
Control Plane identity. This ADR does not claim universal non-correlation for
the Data Plane, its operator, or network observers before an observer model and
mechanism are selected.

- the artifact carries no long-lived identifier and no derivation that yields
  one;
- the artifact's structure and content, on inspection, do not distinguish one
  user, account, or device from another beyond what the session's purpose
  unavoidably requires.

Later mechanism-selection ADRs must identify:

- relevant observer positions;
- auxiliary information available to each observer;
- timing and traffic-analysis channels;
- infrastructure-collusion assumptions;
- identity-service and relay visibility;
- cross-session and cross-device correlation risks;
- mitigations;
- and residual correlation that remains after mitigation.

Any unavoidable residual linkability introduced by the eventual mechanism
must be explicitly disclosed by the later mechanism-selection ADR and
justified against product evidence.

### Return flow from Data Plane to Control Plane

Information flowing back from the Data Plane to the Control Plane is
restricted to the minimum session state required for the Control Plane to
manage the session's lifecycle (for example: session ended, capability
exhausted, revocation acknowledged). The return flow must not:

- carry Data-Plane-observed identifiers that could re-link the session to
  Control-Plane identity;
- carry session content or content-derived material;
- promote transient session state into long-lived Control-Plane records.

Session-ended, capability-exhausted, revocation-acknowledged, and similar
return-flow events are metadata and may create timing or behavioral linkage.

## Bidirectional trust assumptions

### What the Data Plane trusts the Control Plane to have already done

- authenticated the participants at the assurance level required by the
  session's purpose;
- authorized the session under applicable policy;
- negotiated the capability set granted to the session;
- established session policy, including lifetime and purpose;
- performed relay or route selection where applicable and conveyed only the
  minimum coordination inputs needed.

The Data Plane does not repeat user authentication, but it must independently
validate that the handoff:

- originated from an authorized Control Plane source;
- has valid integrity and provenance;
- is intended for this Data Plane recipient or endpoint role;
- is bound to this session, purpose, and capability set;
- is fresh and unexpired;
- is not revoked or invalidated under the later selected mechanism;
- and does not grant transferable authority solely through possession.

The Data Plane does not re-perform the Control Plane's responsibilities; it
enforces the bounded context it was handed.

### What the Control Plane trusts the Data Plane to do or not do

The Control Plane trusts the Data Plane to:

- enforce the bounded session context and reject use outside it;
- avoid leaking or persisting correlatable session identifiers;
- avoid reusing session artifacts across sessions;
- report only the minimum required session state back to the Control Plane;
- avoid promoting Data-Plane metadata into long-lived identity or account
  state.

These are explicit architectural trust assumptions and must later be converted
into enforceable and testable properties. They are not assurances that the Data
Plane is inherently trusted or already compliant. They are to be verified by
later Data-Plane ADRs and by product evidence; this ADR does not claim they
are already enforced.

## Local trust-boundary analysis

Only boundaries the Control Plane directly touches are listed. Systemwide
consolidation is deferred until both Control Plane and Data Plane ADRs exist.

### Client ↔ Control Plane boundary

- **Trusted party:** the Control Plane, within its stated responsibilities.
- **Untrusted party:** the client, until authenticated to the assurance level
  the session requires.
- **Data crossing:** session request, negotiated context.
- **Secrets/credentials:** authentication material handled per authentication
  negotiation; not persisted beyond session need.
- **Metadata visible:** session request timing, purpose, and negotiated
  capabilities.
- **Failure consequences:** unauthorized session establishment, disclosure of
  more Control Plane metadata than required.
- **Required protections:** mutual authentication where required,
  confidentiality and integrity of the exchange, fail-closed behavior.
- **Unknown:** the transport substrate's own observability characteristics
  (deferred).

### Control Plane ↔ Data Plane boundary

- **Trusted party:** neither is trusted with the other's internals; each is
  trusted only to honor the interface contract.
- **Untrusted party:** each side, with respect to the other's internal state.
- **Data crossing:** the minimized handoff payload; the minimized return
  flow.
- **Secrets/credentials:** the handoff artifact only, bounded as above.
- **Metadata visible:** session purpose, capabilities, lifetime.
- **Failure consequences:** cross-plane correlation; over-scoped Data Plane
  authority; replay.
- **Required protections:** the handoff requirements above; replay
  resistance; revocation.
- **Unknown:** the exact residual linkability introduced by the eventual
  handoff mechanism (deferred).

### Control Plane ↔ relay infrastructure boundary

- **Trusted party:** relays for transport duties only.
- **Untrusted party:** relays with respect to session content and identity.
- **Data crossing:** relay capability and coordination inputs.
- **Secrets/credentials:** none beyond what the relay role requires
  (deferred).
- **Metadata visible:** the fact of relay use, timing, and coarse capability
  fit.
- **Failure consequences:** relay-side correlation across sessions; relay
  outages leaking session existence.
- **Required protections:** no stable user-to-relay identity map, ephemeral
  references, ability to withdraw relay authority.
- **Unknown:** the deferred transport's inherent exposure at relay observation
  points.

### Control Plane ↔ future identity/key service boundary

- **Trusted party:** the identity/key service, with trust scoped only to
  particular claims or key operations at a stated assurance level.
- **Untrusted party:** the identity/key service with respect to session
  content and session-derived state.
- **Data crossing:** authentication requests, claims, key material references
  (all deferred in form).
- **Secrets/credentials:** authentication material; long-lived credentials
  must not be exported into session context.
- **Metadata visible:** authentication events and their assurance level.
- **Failure consequences:** identity-service-side correlation of sessions;
  credential leakage into session context.
- **Required protections:** assurance statement and verification;
  cryptographic agility; least-privilege information flow.
- **Unknown:** the eventual identity system; deferred.

The service may observe authentication events and must be analyzed as a
potential correlation point. Privacy must not depend only on operator promises.

## Required architectural properties

Requirements the Control Plane must satisfy or demonstrably approximate; not
claims of achievement.

- Metadata minimization.
- Resistance to unnecessary correlation.
- Unlinkability where feasible.
- Explicit disclosure of unavoidable linkability, including the observer
  model, the conditions, and the residual exposure.
- Mutual authentication as a required property where the session's purpose
  requires it.
- Confidentiality and integrity of Control Plane interfaces.
- Replay resistance.
- Revocation capability.
- Cryptographic agility.
- Evidence-supported post-quantum posture; no post-quantum property is
  claimed by this ADR and no primitive is selected here.
- Infrastructure replaceability.
- Transport replaceability.
- Least-privilege information flow.
- Session-scoped authority.
- Secure failure behavior (fail closed).

## Deferred decisions

The following decisions are deliberately deferred and must be made by later
ADRs backed by product evidence. Naming a candidate below does not select it.

- cMixx selection.
- xxDK selection.
- Control-plane transport (candidates include, without selection or
  endorsement: QUIC, TLS, Noise, WireGuard, WebRTC, WebSocket, and others).
- Identity system.
- Credential format.
- Key lifecycle.
- Token format.
- Session binding mechanism.
- Relay system.
- Backend topology.
- Storage system.
- API shape.
- Programming language.
- Framework.
- Deployment.
- Recovery model.
- Device enrollment.
- Account model.

## Evidence discipline

The architectural assignments in this ADR are current design judgments. They
are not measured operational evidence. Any later mechanism-selection ADR that
relies on these assignments must justify or revise them with product
evidence. No security, privacy, unlinkability, or post-quantum property is
considered proven by this ADR alone. Requirements listed above are
requirements, not achievements.

## Consequences

- Later mechanism-selection ADRs have a bounded object to be evaluated
  against, with named required properties and an explicit deferred-decision
  list.
- Independent evaluators can verify structural compliance of proposals against
  the responsibility model, the interface contracts, and the handoff
  requirements without needing to reason about mechanisms.
- The Control Plane's correctness is deliberately decoupled from any single
  transport, identity system, or cryptographic primitive; that decoupling
  imposes a design cost on later ADRs, which must preserve the required
  properties.
- Because no mechanism is selected here, this ADR alone does not enable any
  code, deployment, or user-facing capability.

## Alternatives considered

- **Defining Control Plane and Data Plane in a single ADR.** Rejected:
  violates Foundry bounded-scope discipline and prevents independent
  evaluation.
- **Selecting a transport or identity system inside this ADR.** Rejected:
  collapses the deferred-decision list, prevents evidence-based selection
  later, and would require this ADR to make claims it cannot support.
- **Defining wire formats or APIs alongside contracts.** Rejected: locks
  mechanism into architecture and blocks transport and infrastructure
  replaceability.
- **Enumerating the full systemwide trust-boundary model here.** Rejected:
  premature until the Data Plane ADR exists; this ADR analyzes only local
  boundaries.

## Non-goals

- Selecting or rejecting cMixx, xxDK, QUIC, TLS, Noise, WireGuard, WebRTC,
  WebSocket, or any other transport, identity provider, cryptographic
  primitive, token format, database, programming language, framework, or
  cloud provider.
- Designing the Data Plane.
- Designing the product.
- Producing code, APIs, schemas, endpoints, transport implementations,
  identity implementations, authentication implementations, storage,
  deployment files, CI, or dependencies.
- Making any claim of achieved security, privacy, unlinkability, or
  post-quantum property.

## Verification

This ADR is structurally verifiable by confirming that:

- every Control Plane responsibility listed in the bounded question is
  covered by the responsibility model;
- every logical interface listed is defined as a property contract, with no
  wire format, endpoint, framework, API, language, or library selected;
- the plane-crossing handoff payload is minimized and prohibited identity or
  linkage material is explicitly listed;
- handoff token requirements are stated as required properties without
  selecting a mechanism;
- trust assumptions are stated in both directions between Control Plane and
  Data Plane;
- local trust boundaries touched by the Control Plane are documented, and
  systemwide consolidation is deferred;
- no implementation mechanism is selected;
- all architectural assignments are labeled as design judgments or
  requirements, not proven facts;
- deferred decisions are sufficiently explicit for the current bounded
  question; newly discovered decisions must be added through later ADRs rather
  than being assumed covered here;
- rollback is achievable through a superseding ADR without code changes.

Structural verification accepts ADR-0003 as an architectural baseline only.
It does not prove:

- that the handoff is operationally minimal;
- that the eight responsibilities are optimally grouped;
- that the trust assumptions are implementable;
- that the privacy and security properties are achievable;
- that latency, availability, recovery, abuse prevention, and usability remain
  acceptable;
- or that the architecture is superior to later alternatives.

Those claims require later ADRs, implementation, threat modeling, measurement,
and product evidence.

## Rollback

This ADR contains no code, no deployment, and no runtime configuration.
Rollback is performed by a superseding ADR that either revises the
responsibility model, revises the interface contracts, revises the handoff
requirements, revises the trust assumptions, revises the local
trust-boundary analysis, or revises the required architectural properties.
Any superseding ADR must preserve ADR-0002's separation invariants or itself
supersede ADR-0002 through the Foundry process.
