# Design System Guide

Purpose: keep React UI reusable and visually consistent without overbuilding a design system during a hackathon.

## Component Levels

Use a lightweight atomic structure:

```text
components/
  atoms/        # Button, Input, Badge, Icon
  molecules/    # FormField, SearchBar, StatItem
  organisms/    # Navbar, SearchPanel, ResultCard
```

Feature-specific compositions belong inside their feature rather than being forced into the global component library.

## Rules

- Reuse before creating.
- Add a prop/variant before duplicating an existing component for a small visual difference.
- Keep reusable components independent from feature-specific data fetching.
- Use semantic HTML first.
- Keep styling values behind tokens/CSS variables where practical.
- Avoid inline style for static design values; dynamic runtime values are acceptable.
- Do not create abstractions for components used once unless they clarify a meaningful boundary.

## Base Tokens

Fill these after `DESIGN_DIRECTION.md` is defined.

```css
:root {
  --color-primary: #000000;
  --color-secondary: #000000;
  --color-background: #ffffff;
  --color-surface: #ffffff;
  --color-text: #111111;
  --color-muted: #666666;
  --color-danger: #b42318;

  --space-1: 0.25rem;
  --space-2: 0.5rem;
  --space-3: 1rem;
  --space-4: 1.5rem;
  --space-5: 2rem;

  --radius-sm: 0.375rem;
  --radius-md: 0.75rem;
  --radius-lg: 1rem;

  --font-size-sm: 0.875rem;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;
  --font-size-xl: 1.5rem;
}
```

Replace placeholder palette values with the confirmed project direction; do not leave a generic visual identity in the final product.

## Accessibility Floor

Reusable interactive components must support:

- Keyboard interaction
- Visible focus state
- Appropriate labels / accessible names
- Disabled states when applicable
- Sufficient contrast

## Example Button Contract

```tsx
type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'sm' | 'md' | 'lg';
};
```

Keep the component API small. Add variants only when a real product need exists.
