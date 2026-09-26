# Design Direction Guide

Fill this only after the PRD is confirmed.

Purpose: define the product's visual personality. `DESIGN_SYSTEM.md` defines reusable component structure; this file defines how this specific product should look and feel.

## Product Context

- Product / theme: _____
- Primary user: _____
- Main environment of use: _____
- Desired feeling: _____

## Visual Reference

Reference image/link (optional): `_____`

Extract attributes such as palette, typography, density, hierarchy, and mood. Do not clone another product's branded identity 1:1.

## Color

Choose a compact palette with named roles.

- Primary: `#_____`
- Secondary: `#_____`
- Accent: `#_____`
- Background: `#_____`
- Surface: `#_____`
- Text: `#_____`

Use color deliberately. Avoid making every card/button compete for attention.

## Typography

- Heading/display font: _____
- Body font: _____
- Type scale: _____

Maximum two font families. Prefer readability over novelty for dense app screens.

## Layout

- Primary alignment: _____
- Density: compact / balanced / spacious
- Main content width behavior: _____
- Navigation model: _____

Key screen wireframe:

```text
+--------------------------------------+
|                                      |
|                                      |
|                                      |
+--------------------------------------+
```

## Signature Element

The one bold visual idea: `_____`

Spend visual risk here. Keep supporting UI quiet and disciplined.

## Motion

Primary purposeful motion: `_____`

Do not add animation to every section or hover state. Respect `prefers-reduced-motion`.

## Avoid Generic AI/SaaS Defaults

Do not automatically fall back to:

- neon-on-black AI dashboard styling
- endless rounded cards with identical shadows
- gradient blobs with no product meaning
- eyebrow labels above every heading
- numbered sections when content is not sequential
- decorative arrow symbols on every CTA
- fade/slide animation on every block

The design should derive from the confirmed user/problem, not from a generic startup template.

## Quality Floor

Non-negotiable:

- Responsive at common mobile and desktop widths
- Usable keyboard navigation
- Visible focus indicators
- WCAG AA contrast for essential text/actions
- Reduced-motion preference respected
- Clear loading, empty, success, and error states
- Body copy remains comfortably readable

## Copy Style

- Plain, active language
- Sentence case
- CTA labels describe the action (`Save changes`, not `Submit`)
- Errors explain what happened and the next useful action
- Empty states guide the user toward a next step
