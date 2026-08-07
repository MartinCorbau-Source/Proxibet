# DECISIONS.md — ProxiBetApp

History of non-obvious design decisions: the *why* behind choices that can't be deduced by reading the current code. Complements [CLAUDE.md](CLAUDE.md), which documents the current state of the design system (the rules to apply); this file documents how we got there and why certain alternatives were discarded.

This file is not loaded into context on every task — consult it only when a past decision seems to contradict the current code, or before revisiting a choice documented here.

## Color rebrand (2026-08-04)

Primary moved from the violet seed hue 253° (`#512BD4`) to a violet hue 265° — same luminosity/saturation curve, alias `Primary` → `Primary400` (`#8244D9`, 5.56:1 on white, AA compliant).

Secondary moved from a desaturated violet (~31% sat., muted neutral tone) to a mint green hue 155° that's noticeably more saturated (~55-68%): an assertive accent (success/streaks/highlights), plus a neutral tone.

**Motive**: Proxibet is a bet between friends with no real money at stake ("just fun among friends"), the visual tone sought is playful/modern rather than serious/fintech — which deliberately rules out blue/indigo "financial trust" directions and the teal seed `#36C9A9` (too close to fintech territory, insufficient contrast at equal lightness), both explored and discarded in favor of the violet/mint-green duo.

Two `LinearGradientBrush` were added at this occasion, same model as `HeaderBackgroundGradientLight`/`Dark` (`AppHeader.xaml`): `GradientHero` (Primary400 → Secondary500, solid, for hero/CTA backgrounds) and `GradientSoft` (Primary200 → Secondary300, light, for badges/highlights) — not consumed by any existing style or component yet, to reuse as-is rather than recreate an equivalent.

### Cascading consequences of the rebrand

- **Pill buttons → medium radius**: `BaseButtonStyle.CornerRadius` used to be `{StaticResource CornerRadiusFull}` (pill shape); dropped in favor of `ButtonCornerRadiusMedium` — motive: rendering judged "extremely ugly" on Windows. Direct consequence: `ButtonRoundedStyle`/`SecondaryButtonRoundedStyle` became visually identical to `PrimaryButtonStyle`/`SecondaryButtonStyle` on this one `CornerRadius` criterion — expected, these styles remain distinct for other `Variant`/`Radius` combinations, not a duplicate to remove. `ButtonSquareStyle`/`SecondaryButtonSquareStyle` stay on `ButtonCornerRadiusSmall`, so still visually distinct from non-Square styles. `CornerRadiusFull` remains defined in `Shapes.xaml` but no longer has any consumer in `Buttons.xaml`.
- **`IconButtonStyle.CornerRadius`**: this override originally existed to escape the pill shape then inherited from `BaseButtonStyle` — on a non-full-width button, the Pressed/PointerOver state-layer following a pill `CornerRadius` produced a stretched oval halo, disproportionate to the actual content. Since `BaseButtonStyle` itself now resolves to `ButtonCornerRadiusMedium`, this `Setter` is redundant with the inherited value but stays explicit in the style — not a regression, just a vestige not to be confused with a genuine radius divergence.
- **`SecondaryButtonStyle`: transparent → solid fill**: no longer transparent/link-styled but a solid fill, same structure as `PrimaryButtonStyle` (fill + `VisualState` Pressed/PointerOver/Disabled) recolored on the Secondary scale (mint green). `BackgroundColor` (Normal) = `Secondary700` (`#1B7952`), same tier Light/Dark: first tier of the Secondary scale that reaches AA on white text (5.38:1) — the semantic alias `Secondary` (= `Secondary100`, `#DBF5EA`, too light) doesn't fit here any more than it fit for `Primary`/`PrimaryButtonStyle` back then, same tier-selection motive. Pressed = `Secondary800`, PointerOver = `Secondary600`. Disabled reuses the `Gray200`/`Gray600` (fill) + `Gray950`/`Gray200` (text) pattern, rather than only graying out the text as before (there was no fill to gray out while the style was transparent).
  - **Deliberate tension, not an oversight**: this style is still used by buttons that are "navigation" in nature rather than "success/status" (`Créer un compte` (Sign up) / `LoginPage`, `J'ai déjà un compte` (I already have an account) / `RegisterPage`, `[Debug] Component Gallery` / `LoginPage`), while Secondary (mint green) was otherwise established as the "status/result" axis via `CardSecondaryStyle`. Explicit user decision: `SecondaryButtonStyle` now means "the green button style", without carrying its own status/success semantics — the two usages (card vs. button) of the same color token deliberately diverge in meaning.
- **`IconButtonStyle` did not follow `SecondaryButtonStyle`**: stays deliberately transparent, text color `Primary`/`PrimaryDark`. No longer read "identical to `SecondaryButtonStyle`" as a rendering equivalence since this rebrand — only a structural inheritance (`BasedOn="{StaticResource BaseButtonStyle}"` directly, not via `SecondaryButtonStyle`). The state-layer tints (`Primary50`/`Primary900` on Pressed/PointerOver, `Gray300`/`Gray600` on Disabled) no longer have an equivalent in `SecondaryButtonStyle` since its move to a solid fill.
- **`Headline`/`SubHeadline`** recolored to `Primary700`/`Primary200` (AppThemeBinding) instead of `MidnightBlue` — an off-palette color that had never been contrast-checked — and now carry `FontFamily="OpenSansSemibold"` (they used to inherit the Regular weight from the implicit `Label`).

## Layout padding 5px everywhere, full-width components

`LayoutPadding` is now `5` in both Default **and** Desktop (`OnIdiom` structure kept despite identical values, to stay consistent with the other `OnIdiom` tokens in the file and be able to re-differentiate them later without changing the binding's type) — replaces the previous `Default=15`/`Desktop=30`.

`CardMaxWidth` (formerly 480, consumed by `ProxiCard.MaxWidth`) was removed: no more dedicated max-width token for cards. `ProxiCard.MaxWidth` when not supplied by the page now resolves to `double.PositiveInfinity` — full width by default, no more fallback to 480. `FormMaxWidth` (480) stays unchanged and distinct: still consumed by `CenteredFormLayout`, with no link to `ProxiCard` since this removal.

## XAML typing of shape tokens (Shapes.xaml)

`CornerRadiusSmall` (8), `CornerRadiusMedium` (12), `CornerRadiusLarge` (16) are `x:Double` (consumed by `RoundRectangle.CornerRadius`, which accepts a double). `CornerRadiusFull` (999) is `x:Int32`: `Button.CornerRadius` is an `int` on the MAUI side, an `x:Double` there triggered a runtime warning "Cannot convert X to type System.Int32".

The same type conflict later showed up on `CornerRadiusSmall`/`Medium` themselves as soon as a `Button.CornerRadius` (not just a `RoundRectangle.CornerRadius`) needed to consume them (styles `ButtonRoundedStyle`/`ButtonSquareStyle`/`SecondaryButtonRoundedStyle`/`SecondaryButtonSquareStyle`). Rather than retyping `CornerRadiusSmall`/`Medium` themselves (which would have broken `FormField.xaml`, which consumes them as `RoundRectangle.CornerRadius`), two dedicated tokens `ButtonCornerRadiusSmall` (8) and `ButtonCornerRadiusMedium` (12) were added as `x:Int32`. These are *value* duplicates, not new sources of truth: `CornerRadiusSmall`/`Medium` remain the source of truth (edited on redesign), `ButtonCornerRadiusSmall`/`Medium` exist only to satisfy the `int` type required by `Button.CornerRadius` — keep them manually in sync if `CornerRadiusSmall`/`Medium` ever change.

Same logic on elevation tokens: `Elevation1/2/3Radius` as `x:Double` (`Shadow.Radius` is a double) and `Elevation1/2/3Opacity` as `x:Single`, not `x:Double` (`Shadow.Opacity` is a `float`, same kind of warning otherwise).

No shared `Shadow` object, deliberately: in MAUI, `Shadow` inherits from `Element` (only one `Parent` allowed) — a single instance declared as a `StaticResource` breaks as soon as a second `Border` consumes it. Every style that wants a shadow therefore instantiates its own inline `<Shadow>` reusing the numeric values (see `CardStyle` in `Cards.xaml` for the reference pattern).

**General rule to remember**: a token's XAML type must match the exact CLR type of the consuming MAUI property, not just "a numeric type that compiles" — the error only surfaces at runtime (binding), never at build time.

## Theme toggle migration (AppHeader → MePage)

The theme-toggle button (`ProxiButton Variant="Icon"`) that lived in `AppHeader` was removed when the dark mode toggle moved to `MePage` as a native `Switch` (bound directly to `MePageModel`). The `Utilities.ThemeIconConverter` resource it consumed was removed entirely along with it. The profile icon had already been removed earlier, when "Me" navigation moved from a header button to a native `TabBar` Shell tab.

Cascading effects: `AuthenticatedPageLayout` no longer carries `ToggleThemeCommand`/`IsDarkMode` (it was a passthrough to the internal `AppHeader`) — only remaining bindable property: `PageContent`. `HomePageModel` lost `IsDarkMode`, `ToggleThemeAsync`, the `_themeService` field, and the `IThemeService themeService` constructor parameter. `HomePageModel.GoToProfileCommand` and `MePageModel.GoToHomeAsync` were also removed: tab navigation is handled natively by Shell.

`MePage` is now the only page in the app where the theme can be toggled, via a `Grid` (Label "Mode sombre" + native `Switch` styled by `Controls.xaml`, not a dedicated component) — revisit as a component if a second preference toggle appears elsewhere in the app.

## Removal of the "[Debug] Component Gallery" button on HomePage

Removed from `HomePage.xaml`, with `HomePageModel.IsDevGalleryButtonVisible` and `GoToDevGalleryCommand` removed in mirror, for lack of any remaining consumer on this page. Access to `Pages/Dev/ComponentGalleryPage.xaml` remains unchanged otherwise via the Debug tab of the `TabBar` Shell (`AppShell.AddDebugTabs()`, `dev/gallery` route) — an access path always separate from this button. `LoginPage.xaml` keeps an equivalent button (`LoginPageModel.IsDevGalleryButtonVisible`/`GoToDevGalleryCommand`), unaffected by this removal.

## "Future components" placeholder on HomePage

`HomePage.xaml` contains a 2-column `Grid` (`*,*`) where each column hosts a `ProxiCard` with a placeholder `Body1Strong` `Label` ("Futur composant A"/"Futur composant B") — a throwaway layout illustrating a future component not yet designed, with no connection to the current design system and no missing token/component to document in `CLAUDE.md`.

## Network bug fixed (July 2026)

`ExecuteWithErrorHandlingAsync` exists because a PageModel that forgets `catch (NetworkException ...)` crashes the whole app on the first network outage — a real bug fixed in July 2026 on `LoginPageModel`/`RegisterPageModel`, which only had `catch (AuthApiException ...)`. A transport error isn't an exotic exception on mobile, it's the common case (airplane, tunnel, captive Wi-Fi, server down).

## Known limitation: nothing technically prevents a hand-written try/catch

`ExecuteWithErrorHandlingAsync` reduces friction (one line instead of ten) and gives a single point to check in code review, but can't technically force a future PageModel to use it — C# offers no way to forbid a hand-written try/catch next to it. CommunityToolkit.Mvvm 8.3.2 also doesn't expose a "run a custom delegate on any exception from an `AsyncRelayCommand`" hook: `AsyncRelayCommandOptions.FlowExceptionsToTaskScheduler` routes the exception to `TaskScheduler.UnobservedTaskException` (a process-wide, non-deterministic event, with no way to associate the exception back to the PageModel instance concerned) — unusable for cleanly populating `ErrorMessage`.

The real anti-crash safety net therefore remains the `Services/` layer: every service method that calls a Refit API MUST itself guarantee that no exception other than `AuthApiException`/`NetworkException` escapes it (the `catch (ApiException) : throw ... catch (Exception) : throw new NetworkException(...)` pattern from `AuthService`). `ExecuteWithErrorHandlingAsync` consumes this service contract, it doesn't replace it — if a future service lets an unwrapped exception leak through, the helper will let it propagate and crash, by design (a visible crash in dev/QA beats a lying `ErrorMessage` that would mask a real bug in the service layer).

## EnablePreviewFeatures not enabled

`MauiAllowImplicitXmlnsDeclaration` (which would also allow omitting the root `xmlns`/`xmlns:x`) is not enabled — it's a preview feature, the marginal gain doesn't justify the stability risk on a project that isn't in production yet.

## Pages/ migration history (Components)

- **HomePage**: migrated to `AuthenticatedPageLayout` (no attributes) in place of the old `AppHeader` + content + bottom-nav shell duplicated page by page.
- **MePage**: migrated with a deliberate, acknowledged deviation — `<AuthenticatedPageLayout>` still wraps the internal `VerticalStackLayout`, but with the theme-toggle card inserted in addition (see the dedicated section above). The user-info block is not (yet) a candidate for a dedicated `FormField`/component — revisit if an `InfoRow`/`ReadOnlyField` component appears.
- **ComponentGalleryPage**: `ComponentGalleryPageModel` (which only carried `IsDarkMode`/`ToggleTheme` since the removal of the `AppHeader` demo card) was removed entirely, for lack of any remaining consumer. The page therefore no longer has a `BindingContext` or `x:DataType`.
