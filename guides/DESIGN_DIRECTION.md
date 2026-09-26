# Tilik Design Direction

Tilik serves grades 4–9, prioritizing grade 4. Student screens should feel welcoming, readable, and calm; teacher screens emphasize evidence and next actions.

## Reference and palette

The supplied Gamified Learning App screenshot informs personal greetings, compact statistics, dark green surfaces, and a lime primary action. It is inspiration rather than a clone. Spin wheels, coins, and invented engagement statistics are outside the current scope. The Figma page could not be inspected directly.

Use forest green `#284d3a`, lime `#d8ef62`, a light canvas `#f6f7f3`, white surfaces, text `#17261e`, and muted text `#58645b`. The invitation panel is the main dark surface; learning questions retain a light canvas for mathematical readability. Tokens live in `apps/web/src/App.css`.

## Hierarchy and layout

Student home: welcome and explicit start/continue action, then profile statistics, then recent history. Login and reload always open home; entering an assessment requires an action. Statistics describe persisted evidence, with honest empty and failure states.

Use a system font, sentence case, left alignment, and balanced spacing. Desktop content has a maximum width of 68.75rem. Below 600px the welcome and profile stack vertically; history actions remain reachable. At 360px and narrower profile statistics stack as well.

## Interaction and accessibility

Use one primary learning action per screen, minimum 44px touch targets, visible keyboard focus, named controls, and explicit loading/retry states. State labels include text, not color alone. Respect reduced motion. Avoid decorative gradients, emoji, fabricated progress, and automatic transitions into questions.
