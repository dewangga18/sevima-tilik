# PRD Guide — Hackathon Execution Mode

Purpose: turn a rough hackathon idea into a confirmed, demoable MVP and a priority-ordered implementation plan without wasting time on low-value work.

Do not generate implementation code, scaffold, deployment config, or final design direction before the product direction is explicitly confirmed.

## Stage 1 — Explore

From the case/theme or initial user idea, propose **2-3 meaningfully different solution angles** when the direction is still open.

Each angle must include:

- Target user
- Specific problem
- Core solution
- Main value
- Primary trade-off
- AI-agent opportunity

An angle is meaningfully different only if it changes the target user, problem emphasis, workflow, or solution model — not just UI or naming.

If the user already provides a clear direction, do not force multiple alternatives. Move directly to Stage 2 and identify only the missing critical information.

### AI-Agent Check

Because a functional AI agent can be a high-value hackathon differentiator, explicitly check whether the solution benefits from an agent that can perform **at least 2 distinct meaningful actions**.

Good pattern:

```text
understand intent
-> choose action/tool
-> execute
-> observe result
-> decide next action or stop
```

Examples of distinct actions:

- search or retrieve data
- create/update a record
- generate a structured plan/output
- compare or rank alternatives
- classify incoming data
- call an external API
- prepare or trigger a workflow
- execute a user-approved operation

A chatbot that only answers questions is not sufficient. Do not force AI into the product when it adds no real value.

If multiple solution angles were proposed, stop after Stage 1 and wait for the user to choose, merge, modify, or reject the direction.

## Stage 2 — Discovery

Ask only questions whose answers materially affect the MVP or implementation plan.

Use at most **3 compact grouped questions**, and skip anything already answered by the user.

The discovery must establish, directly or implicitly:

1. **User + problem** — Who is the primary user and what single problem must the MVP solve?
2. **Demo path + scope** — What must a judge/user be able to do from start to finish? Which capabilities are must-have, nice-to-have, and explicitly out of scope?
3. **Constraints + differentiator** — Time, data/API availability, required stack, external services, and whether AI performs meaningful actions rather than only chat.

Do not conduct a long product interview. Ask follow-ups only when ambiguity would block a correct build decision.

## Stage 3 — Confirm Product Direction

Produce a compact draft:

```markdown
## Chosen Direction
## Problem Statement
## Target User
## Core Demo Flow
## Must-Have Features
## AI Agent Actions (if applicable)
## Nice to Have
## Out of Scope
## Success Criteria
## Constraints
```

Wait for explicit confirmation before implementation planning.

## Stage 4 — Prioritize

After confirmation, classify work by hackathon value.

### P0 — Core Demo Path

Work required for the primary end-to-end demo to function.

Examples:

- minimum app shell/navigation
- minimum auth or guest identity only if required
- essential data model
- critical API flow
- main user input -> processing -> visible result

P0 must be completed before optional work.

### P1 — Differentiator / Challenge Value

Features that create the strongest product value or directly demonstrate the hackathon differentiator.

Examples:

- functional AI agent
- 2+ meaningful agent actions
- core automation
- unique external integration
- feature directly tied to judging criteria

### P2 — Reliability / Quality Floor

Required quality that prevents the demo from feeling broken or unsafe.

Examples:

- loading states
- error states
- empty states
- essential validation
- responsive layout
- accessibility baseline
- security baseline
- recovery from common failures

### P3 — Polish

Non-critical improvements that should never block P0-P2.

Examples:

- animations
- decorative interactions
- secondary screens
- visual refinements
- extra filters
- optional settings

When time is limited, cut scope from P3 first, then non-essential P2/P1 items. Never sacrifice the core P0 demo path for polish.

## Stage 5 — Generate Implementation Plan

Create an implementation plan **before coding**.

The plan must:

- be split into small phases
- execute the highest-priority work first
- use vertical slices instead of finishing an entire technical layer first
- keep the app runnable and demoable after every completed phase
- identify dependencies/blockers early
- explicitly mark optional work that can be dropped when time runs short

### Required Plan Format

```markdown
## Implementation Plan

### Phase 1 — Demoable Core [P0]
Goal: complete the smallest end-to-end user flow.

- [ ] ...
- [ ] ...

Demo checkpoint:
- User can ... -> API/service does ... -> UI shows ...

### Phase 2 — Core Value / Differentiator [P0/P1]
Goal: make the product demonstrate its primary value.

- [ ] ...
- [ ] ...

Demo checkpoint:
- ...

### Phase 3 — Reliability [P2]
Goal: make the demo resilient enough for judges/users.

- [ ] loading/error/empty states
- [ ] critical validation
- [ ] responsive/accessibility/security baseline

### Phase 4 — Polish [P3]
Goal: improve presentation only after the product is already demoable.

- [ ] ...

### Drop First If Time Is Short
- ...
```

## Vertical Slice Rule

Do not plan work like this by default:

```text
Phase 1: build all backend
Phase 2: build all frontend
Phase 3: integrate everything
```

Prefer:

```text
Phase 1: user input -> Go API -> result -> React UI
Phase 2: agent action A -> API/tool -> visible result
Phase 3: agent action B -> API/tool -> visible result
```

A phase is complete only when its behavior is observable through the product or through a meaningful integration checkpoint.

## Execution Rule

Once the implementation plan is confirmed:

1. Start from the highest incomplete priority.
2. Finish the current phase's demo checkpoint before moving to lower-priority work.
3. Do not start P3 while critical P0/P1 work remains incomplete.
4. If a blocker appears, choose the smallest fallback that preserves the core demo flow.
5. Keep the project runnable after each meaningful change.
6. Re-plan when requirements materially change; do not silently expand scope.

## PRD Quality Bar

A confirmed PRD + implementation plan must answer:

- Who is this for?
- What one problem matters most?
- What exact workflow must work in the demo?
- What is P0, P1, P2, and P3?
- What is deliberately excluded?
- What proves the MVP works?
- If AI is used, what real actions does the agent perform?
- What phase should be implemented first?
- What can be dropped safely if time runs short?
