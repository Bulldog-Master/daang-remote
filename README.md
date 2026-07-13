# Daang Remote

Daang Remote is the first product authorized under Foundry.

Its initial purpose is to let the founder securely, privately, and reliably interact with his own computing devices across platforms and networks.

The product is governed by Foundry 1 and must preserve:

- metadata minimization;
- metadata resistance;
- prevention of avoidable correlation;
- explicit disclosure of unavoidable linkability;
- cryptographic agility;
- a present-day post-quantum posture where supported by evidence;
- infrastructure and platform replaceability;
- and an experience good enough to replace the founder's current commercial remote-access product.

## Current status

Repository foundation only.

No application architecture, implementation language, framework, transport, identity mechanism, control plane, data plane, relay system, cMixx integration, xxDK integration, authentication model, storage model, deployment model, or update system has been selected in this repository.

Those decisions must be made through bounded work and Architecture Decision Records under Foundry 1.

## Authoritative governance

The authoritative Foundry repository is:

https://github.com/Bulldog-Master/foundry

The authoritative documents are:

- Foundry Constitution  
  https://github.com/Bulldog-Master/foundry/blob/main/CONSTITUTION.md
- Foundry generation  
  https://github.com/Bulldog-Master/foundry/blob/main/VERSION.md
- Foundry 1 operating ADR  
  https://github.com/Bulldog-Master/foundry/blob/main/adrs/ADR-0001-model-neutral-foundry-boundaries.md
- Daang Remote product charter  
  https://github.com/Bulldog-Master/foundry/blob/main/product-charters/daang-remote.md

This repository must not silently diverge from those documents.

If a conflict exists, the Foundry Constitution and accepted ADRs govern.

## Foundry 1 review contract

Every substantive change is reviewed against four mandatory gates:

1. Architecture
2. Security and Cryptography
3. Privacy and Metadata
4. Quality and Verification

The producer of a change cannot be its sole evaluator.

Independent evaluation may come from:

- a different human reviewer;
- a separate AI model or intelligence;
- a separate evaluation session that did not produce the change;
- deterministic verification;
- or an explicit combination.

The founder remains the final approval authority.

A failed mandatory gate blocks merge by default.

Only an explicit, recorded founder override may allow a failed gate to proceed.

An override does not convert a failed gate into a pass.

## Property versus mechanism

The Daang Remote charter defines required properties.

It does not preselect every implementation mechanism.

The product requires:

- minimized metadata;
- resistance to unnecessary correlation;
- prevention of avoidable linkage;
- explicit disclosure of unavoidable linkability;
- cryptographic agility;
- evidence-supported post-quantum posture;
- platform independence;
- infrastructure replaceability.

cMixx and xxDK are candidate mechanisms where they provide measurable value.

They are not hard-coded product doctrine.

A later Architecture ADR must decide whether they belong in:

- discovery;
- rendezvous;
- session negotiation;
- identity exchange;
- control-plane messaging;
- relay coordination;
- or another bounded role.

The latency-sensitive interactive data path must be evaluated separately and must not be forced through a mechanism merely because that mechanism is associated with the product's privacy goals.

## Next valid work

The next pull request after this repository foundation must be one bounded Daang Remote change packet.

It must:

- define one concrete problem;
- declare exact scope;
- define observable acceptance criteria;
- identify the architectural layers touched;
- state what is known and assumed;
- list privacy, metadata, security, and cryptographic risks;
- describe rollback;
- run through all four mandatory gates;
- record whether each gate added value.

Do not begin with broad application scaffolding, a complete architecture, or an agent framework.

Begin with one bounded architecture decision or implementation slice that can produce real evidence.
