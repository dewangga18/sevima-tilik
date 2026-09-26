# Tilik Design System

Use existing components before introducing abstractions. The shared stylesheet `apps/web/src/App.css` owns palette, spacing, typography, radii, and content-width tokens; `index.css` owns document defaults and accessibility resets. Feature layout lives in `StudentHome.css`.

## Tokens

Primary/hover: `#284d3a` / `#1c392a`; primary light: `#eef3e9`.
Accent/hover: `#d8ef62` / `#c7df4e`; inverse text/muted: `#ffffff` / `#dce7da`.
Canvas/surface/border: `#f6f7f3` / `#ffffff` / `#dfe5dc`.
Text/muted: `#17261e` / `#58645b`.
Success, warning, and danger have separate semantic tokens; use accompanying text labels.

Spacing tokens 1–8: 0.25, 0.5, 0.75, 1, 1.5, 2, 3, 4rem.
Type scale: 0.875, 1, 1.125, 1.5, 2, 2.5rem. System font stack; no remote font dependency.
Radii: 6, 10, 16px. Content width: 68.75rem.

## Component contracts

- Buttons reuse `.btn` with primary/secondary variants, explicit disabled state, and a minimum 44px height. The invitation uses the accent variant through its existing context.
- Navbar presents identity/logout and the student links Beranda, Profil belajar, and Riwayat; links return from activities to the matching home section.
- Profile statistics use a description list. Empty evidence is never rendered as measured zero mastery.
- History is a semantic list with real dates, status text, and an explicit open action.
- Quiz options remain keyboard-operable buttons. Screen transitions focus the activity heading; errors have alert/status semantics.
- Hidden assessment views remain noninteractive and preserve unsent choices while returning to home.

- Teacher/admin shells share `ManagementDashboard` layout with a sidebar, topbar, and content workspace; navigation differs by role. Curriculum uses a semantic table with local horizontal scrolling. Unimplemented management features show explicit availability states.
- Below 600px dashboard navigation collapses behind an expanded-state menu button; selecting a section closes the menu and focuses its heading. Student links stay in the navbar on a second row below 800px.

## Responsive and accessibility floor

Student layout reflows at 800, 600, and 360px. Verify 320px through desktop without horizontal overflow. Interactive targets are at least 44px. Essential text/actions meet WCAG AA contrast, focus remains visible, and reduced-motion preferences are respected. Keep semantic HTML and reuse components; do not reorganize the repository into an atomic component hierarchy merely for this guide.

## Request states

Remote flows expose loading with status/aria-busy and disable duplicate actions. Errors use alert semantics and a named retry when recoverable; empty data is distinct from failure. Successful curriculum loading and saved assessment results have status text; logout confirms local completion on the login screen. Missing backend features display availability rather than fabricated empty datasets or working forms.
