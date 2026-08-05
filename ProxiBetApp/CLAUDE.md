# CLAUDE.md — ProxiBetApp (.NET MAUI)

This file documents the design system of the MAUI mobile app (`ProxiBetApp/ProxiBetApp/`). It complements the monorepo's root `CLAUDE.md`, which covers `proxiback/` and `proxifront/`. For the *why* behind past decisions (color rebrand, XAML typing choices, page migrations...), see [DECISIONS.md](DECISIONS.md) — this file documents the current state and the rules to apply, not the history.

## Design system acceptance principle

**A visual redesign (colors, radii, spacing, typography) must be doable by editing only `Resources/Styles/` and/or `Components/` — never `Pages/*.xaml`.**

If a design change forces a page edit, that's a sign a token or component is missing: add it to the appropriate layer instead of patching the page.

Before any token change, validate the rendering on `Pages/Dev/ComponentGalleryPage.xaml` (route `dev/gallery`, Debug builds only): it instantiates every component/variant and serves as a visual test bench, faster than navigating Login → Register → Me.

**Never launch/run the app yourself to test a change** (`dotnet build -t:Run`, device/emulator deployment, opening Visual Studio, etc.): the user always tests on their end. A plain `dotnet build` to check compilation remains acceptable, but launching the app is out of scope — don't suggest it or do it on your own initiative.

**XAML comments: only when necessary, otherwise none.** The default is no comments. A comment is only justified for a genuinely non-obvious WHY (a hidden constraint, a specific platform bug) that a reader couldn't deduce from the code itself — never to restate what the code already does. If a comment is needed, one line is enough. No large multi-paragraph blocks at the top of a file, no narrative of past attempts in the file itself — that kind of context belongs in [DECISIONS.md](DECISIONS.md), not in the code.

**Look for the existing .NET MAUI primitive before rolling your own.** MAUI already provides mechanisms for most navigation/UI-state needs (e.g. `Shell` natively handles active-tab highlighting via `TabBar`/`Shell.CurrentItem`, see `AppShell.xaml`). Before implementing a homegrown solution (manual route parsing, custom events, state polling), explicitly check whether the standard MAUI API already covers the need — and if a project constraint prevents using it, document that clearly as a deliberate workaround. When in doubt about the exact behavior of a MAUI API, verify with a targeted test (a small `dotnet build`/log, reading the official docs) rather than guessing.

## Token layer (`Resources/Styles/`)

Files merged in `App.xaml`, in this order (order matters: implicit styles later in the list can rely on earlier ones):

1. **`Colors.xaml`** — Primary/Secondary scales (10 levels, generated from an HSL seed), semantic aliases (`Primary`, `Error`, `ErrorDark`...), Dark/Light background tokens, grayscale, matching brushes. `Outline`/`OutlineVariant` (light) and `OutlineDark`/`OutlineVariantDark` (dark) for outlined form-field borders (consumed by `FormField`). `ShadowColor`/`ShadowColorDark` for the inline `Shadow` of `Shapes.xaml`'s elevation tokens. `GradientHero`/`GradientSoft`: two `LinearGradientBrush` available for hero/CTA backgrounds and badges/highlights, not yet consumed by any style or component.
2. **`Spacing.xaml`** — numeric scale `size20`-`size560`, `IconSize(Small)` (OnIdiom), `LayoutPadding`/`LayoutSpacing` (OnIdiom, `5` in both Default and Desktop), `FormMaxWidth` (480, consumed by `CenteredFormLayout`), `FormSpacing`.
3. **`Shapes.xaml`** — shape system: `CornerRadiusSmall` (8), `CornerRadiusMedium` (12), `CornerRadiusLarge` (16) as `x:Double` (consumed by `RoundRectangle.CornerRadius`). `CornerRadiusFull` (999, pill) as `x:Int32`. `ButtonCornerRadiusSmall`/`Medium` (duplicate the value of `CornerRadiusSmall`/`Medium` as `x:Int32` to satisfy `Button.CornerRadius`) — keep manually in sync if `CornerRadiusSmall`/`Medium` change. Numeric elevation tokens: `Elevation1/2/3Radius` (`x:Double`) and `Elevation1/2/3Opacity` (`x:Single`). No shared `Shadow` object (MAUI: only one `Parent` allowed per `Shadow`) — every style instantiates its own inline `<Shadow>` reusing these values (see `CardStyle` in `Cards.xaml`). **Typing rule**: a token's XAML type must match the exact CLR type of the consuming MAUI property — the error only surfaces at runtime, never at build time. Full reasoning: [DECISIONS.md](DECISIONS.md#xaml-typing-of-shape-tokens-shapesxaml).
4. **`Typography.xaml`** — implicit `Label` style + Fluent 2-style keyed scale (`Caption2/1/1Strong`, `Body2/2Strong/1/1Strong`, `Title3/2/1`, `LargeTitle`, `Display`). `Headline`/`SubHeadline` on `Primary700`/`Primary200` (AppThemeBinding), `FontFamily="OpenSansSemibold"`.
5. **`Buttons.xaml`** — 7 named styles, in 3 inheritance layers:
   - `BaseButtonStyle` (keyed, common base): `FontFamily="OpenSansSemibold"`, `FontSize=16`, `BorderWidth=0`, `CornerRadius={StaticResource ButtonCornerRadiusMedium}`, `Padding`, `Minimum*Request`. Keyed only so it can serve as a `BasedOn` target (MAUI doesn't allow a style to inherit from a non-keyed implicit style).
   - `PrimaryButtonStyle` (keyed, `BasedOn="{StaticResource BaseButtonStyle}"`): source of truth for the "primary button" rendering — color (`Primary`/`White`) + `VisualState` Pressed/PointerOver/Disabled.
   - Implicit `Button` style (non-keyed, `BasedOn="{StaticResource PrimaryButtonStyle}"`): any `Button` without an explicit `Style` inherits `PrimaryButtonStyle`.
   - `SecondaryButtonStyle` (keyed, `BasedOn="{StaticResource BaseButtonStyle}"`): solid `Secondary700` fill/white text, `VisualState` Pressed=`Secondary800`/PointerOver=`Secondary600`, Disabled `Gray200`/`Gray600` (fill) + `Gray950`/`Gray200` (text). Designates "the green button style" with no status/success semantics of its own (see [DECISIONS.md](DECISIONS.md#color-rebrand-2026-08-04) for the deliberate tension with `CardSecondaryStyle`).
   - `IconButtonStyle` (keyed, `BasedOn="{StaticResource BaseButtonStyle}"` **directly**, not via `PrimaryButtonStyle`/`SecondaryButtonStyle`): glyph-only icon buttons, resolved by `ProxiButton` via `Variant="Icon"` (`Radius` ignored). Transparent, text `Primary`/`PrimaryDark`, state-layer `Primary50`/`Primary900` (Pressed/PointerOver), `Gray300`/`Gray600` (Disabled). Explicit `CornerRadius` `ButtonCornerRadiusMedium`, `Padding` reduced to `size80`. No fixed `WidthRequest`/`HeightRequest` (`MinimumWidthRequest`/`MinimumHeightRequest` at 44, inherited from `BaseButtonStyle`).
   - 4 `CornerRadius` variants (`ButtonRoundedStyle`, `ButtonSquareStyle`, `SecondaryButtonRoundedStyle`, `SecondaryButtonSquareStyle`), each `BasedOn` `PrimaryButtonStyle` or `SecondaryButtonStyle` (not `BaseButtonStyle`) to also inherit color + `VisualState`, overriding only `CornerRadius` (`ButtonCornerRadiusMedium`/`Small`).

   These 7 styles are the resolution target of `ProxiButton.UpdateStyle()` (`Variant` × `Radius` → one of the 7 names, `Icon` ignoring `Radius`). Any new button variant must follow the same inheritance scheme — never duplicate color/`VisualState` by hand.
6. **`Inputs.xaml`** — implicit `Entry`/`Editor`/`Picker` styles. The implicit `Entry` stays deliberately minimal (no visual container): the `FormField` component is what carries the outlined container. `PlaceholderColor` = `Gray400`. `FontSize` references `InputFontSize` (OnIdiom, `Spacing.xaml`), `MinimumHeightRequest`/`MinimumWidthRequest` reference `size400` (40). **Known exception**: `Editor` has a hardcoded `FontSize="14"` (component unused in current forms) — not a debt to replicate elsewhere.
7. **`Cards.xaml`** — implicit `Border` style + `CardStyle` (keyed), `CornerRadiusLarge`, inline `Shadow` at Elevation1 level. `CardPrimaryStyle` (`BasedOn="{StaticResource CardStyle}"`) overrides `Background` = `Primary50`/`Primary900`. `CardSecondaryStyle` (same scheme) = `Secondary50`/`Secondary900` — carries the "status/result" axis, distinct from Primary ("action", buttons). Consumed by `ProxiCard` (resolves `CardColor`: Default/Primary/Secondary).
8. **`Shell.xaml`** — `Page`/`Shell`/`NavigationPage`/`TabbedPage` styles.
9. **`Controls.xaml`** — implicit styles for the remaining native controls (ActivityIndicator, Switch, CheckBox...).

**Hard rule**: any color, size, spacing, or radius used in `Components/*.xaml` or `Pages/*.xaml` must reference a `{StaticResource}` (optionally via `{AppThemeBinding}` for Dark/Light). If the needed value doesn't have a token yet, add it first in the relevant token file (with a comment explaining its origin/usage), then consume it — never hardcode a value "for now". Shadow special case: it's the inline `<Shadow>` instantiation that must reuse the `ElevationN Radius/Opacity` tokens — never hardcoded radius/opacity values copy-pasted from one style to another.

This rule also applies **inside `Resources/Styles/` itself**: a `Setter` in a token file must also reference a `{StaticResource}` from a file merged earlier in `App.xaml` rather than a literal, except for the token's own definition (e.g. `<x:Double x:Key="size400">40</x:Double>` in `Spacing.xaml`, which *is* the source of truth). Before adjusting a size/font/spacing value in `Resources/Styles/`, first check whether an existing token (even in a different file) already covers the target value before writing a new one.

## Component layer (`Components/`)

Reusable composite components, on top of the implicit/keyed styles. Convention:

- Each component = a `ContentView` in XAML + code-behind, exactly like pages (`ComponentName.xaml` + `ComponentName.xaml.cs`, `x:Class="ProxiBetApp.Components.ComponentName"`). No C# markup.
- Public API = `BindableProperty` only. A consuming page must never override a component's internal children from its own XAML.
- Visual variants exposed via a bindable property (enum if there are several values), which internally selects a named `Style` — never via ad hoc setters on the caller's side.
- "Dumb"/presentational components: no reference to a `PageModel`. Binding to business logic happens only through the component's exposed `BindableProperty`.
- `Components/` is organized as one subfolder per component (`Components/ProxiButton/`, `Components/ProxiCard/`, `Components/FormField/`...), each grouping its `.xaml`/`.xaml.cs` and associated enum if applicable. No grouping by technical category.

Existing components:
- **`ValidationMessage`** — error message (`Error`/`ErrorDark`), auto-hides when `Text` is empty.
- **`FormField`** — `Label` + `Entry` + `ValidationMessage`, bindable properties `FieldLabel`, `Text` (TwoWay), `Placeholder`, `Keyboard`, `IsPassword`, `ErrorMessage`. Material 3 "outlined" container: the `Entry` is wrapped in a `Border` (`StrokeShape="RoundRectangle {StaticResource CornerRadiusSmall}"`) whose `Stroke` reacts to state via a 3-state `VisualStateGroup FieldStates`: `Resting` (`Outline`/`OutlineDark`, thickness 1), `Focused` (`Primary`/`PrimaryDark`, thickness 2), `Error` (`Error`/`ErrorDark`, thickness 2). Priority: error > focus > resting.
- **`CenteredFormLayout`** — `ScrollView > VerticalStackLayout` shell, centered/max-width, for form pages; accepts multiple direct children in XAML via `ContentProperty` (MAUI has no `ContentPresenter` outside a `ControlTemplate`).
- **`AppHeader`** — a visual fade only: a `Grid` with no interactive content, `Background` as a `LinearGradientBrush` (`HeaderBackgroundGradientLight`/`Dark`) overlaying a gradient above the content scrolling behind it. No `BindableProperty`. Embedded as the first child of each authenticated page's root layout rather than via `Shell.TitleView` (per-page/per-`ShellContent`, doesn't auto-share across Shell tabs).
- **`AuthenticatedPageLayout`** — common shell for authenticated pages: a `Grid` with no `RowDefinitions` holding an overlaid `ScrollView`/projected content and `AppHeader` (`AppHeader` at `ZIndex="1"`), content projected via `ContentProperty`/`ObservableCollection<View>` (`PageContent` property). Bottom navigation is carried by the native `TabBar` Shell (`AppShell.xaml`), not by this component. Only bindable property: `PageContent`. `Shell.NavBarIsVisible="False"` must still be declared by each consuming `ContentPage`.
- **`ProxiButton`** — systematic replacement for any direct `<Button Style="{StaticResource ...}"/>`: no native `<Button>` exists in `Pages/`/`Components/` outside of the internal `InnerButton`. Properties `Text`, `Command`, `CommandParameter` (passthrough); `IsEnabled`/`IsVisible` inherited from `VisualElement`. `Variant` (enum `ButtonVariant`: Primary/Secondary/Icon, default Primary) and `Radius` (enum `ButtonRadius`: Default/Medium/Small, default Default) resolved in code-behind (`UpdateStyle()`) to one of the 7 `Buttons.xaml` styles. `Variant="Icon"` ignores `Radius`. `FontFamily`, `ImageSource`/`ContentLayout` as direct passthrough. `FontSize` (double, sentinel default `-1d`) applied only if `>= 0`, so it never silently overrides the resolved style's `FontSize` (see principle below).
- **`ProxiCard`** — replaces direct use of `<Border Style="{StaticResource CardStyle}">`. `Color` (enum `CardColor`: Default/Primary/Secondary) resolved to `CardStyle`/`CardPrimaryStyle`/`CardSecondaryStyle`. `HasShadow` (bool, default `true`): if `false`, removes the `Shadow` set by the style (`InnerBorder.Shadow = null`) rather than rebuilding it in C#. `MaxWidth` (double, sentinel default `-1d`) resolves to `double.PositiveInfinity` when not supplied. `CardContent` (`ObservableCollection<View>`) uses the same `[ContentProperty]`/`CollectionChanged` mechanism as `CenteredFormLayout.FormContent`/`AuthenticatedPageLayout.PageContent`.

### Principle: component = style selector, not rendering engine

`ProxiButton` and `ProxiCard` both follow the same principle, to be respected by any future reusable component exposing visual variants: **the component never codes its own rendering** (colors, `VisualState`, shadows with `AppThemeBinding`) — that logic lives entirely in `Resources/Styles/`, as named XAML `Style`s, the sole visual source of truth. The component only chooses *which* named style to apply, via a simple `switch` in code-behind (`UpdateStyle()`), assigned via `Application.Current.Resources[styleKey]`.

Additional sub-pattern for properties that should only apply when explicitly supplied by the consuming page (`FontSize` on `ProxiButton`, `MaxWidth` on `ProxiCard`): a sentinel default value (`-1d` for a `double`) applied conditionally in code-behind (`if (value >= 0) InnerControl.Property = value`), rather than a direct XAML binding — a direct binding would systematically apply the `BindableProperty`'s C# default (`0d`) at initialization, silently overriding the resolved default value without any page having asked for it.

## Naming conventions

- PascalCase, no `Custom`/`My` prefix.
- Systematic file pair `Name.xaml` + `Name.xaml.cs`.
- Rationale comments in the XAML (styles and components) in French, matching existing usage: explain the *why*, not the *what*.
- `BindableProperty`: standard MAUI convention (static `NameProperty` + `Name` accessor), no deviation.

## Service layer (`Services/`)

Feature-first organization: one business domain = one `Services/` subfolder (e.g. `Services/Auth/`), mirrored by `Models/<Domain>/`. Inside a domain, all files stay flat — same threshold rule as `Components/` (stay flat under ~8-10 files; technical-role subfolders only once a category justifies 3+ files **and** the folder exceeds that threshold).

Files that are cross-cutting to the whole app (`ApiConfig`, `IErrorHandler`/`ModalErrorHandler`, `NetworkException`) stay at the root of `Services/`, never in a domain subfolder — even if only one domain consumes them today.

**No vertical slice** (`Features/<Domain>/{PageModels,Services,Models}`): the horizontal layers `Pages/` → `PageModels/` → `Services/`/`Models/` remain the reference structure, to be reconsidered only if several domains each grow large enough to have their own sizeable `PageModels`/`Models`/`Services`.

### Network error handling (PageModels)

Any `[RelayCommand]` on a PageModel that calls a service consuming a remote API must go through `PageModelBase.ExecuteWithErrorHandlingAsync` (`PageModels/PageModelBase.cs`) rather than writing its own try/catch. This helper centralizes: `IsBusy = true` / `ErrorMessage = null` before the call, `catch (AuthApiException)` (server-side business error — via `onAuthError` if a specific reaction is needed like logout + redirect, otherwise a default message in `ErrorMessage`), `catch (NetworkException)` (transport failure, generic message), `IsBusy = false` in a `finally`. See `LoginPageModel.LoginAsync`/`RegisterPageModel.RegisterAsync` (simple case) and `HomePageModel.AppearingAsync`/`MePageModel.AppearingAsync` (case with `onAuthError`) as references. `PageModelBase` also carries `IsBusy`/`ErrorMessage` (`[ObservableProperty]`).

Why this convention exists and its known limits: [DECISIONS.md](DECISIONS.md#network-bug-fixed-july-2026).

New API domain (e.g. a future `IOrdersApi`): create its own business exception type on the model of `AuthApiException` (don't reuse `AuthApiException` for a non-auth domain), but reuse `NetworkException` as-is (`Services/NetworkException.cs`, root of `Services/`).

## xmlns convention

All XAML files use the .NET MAUI 10 global xmlns `http://schemas.microsoft.com/dotnet/maui/global` as the default, instead of per-file `xmlns:components`/`xmlns:pageModels` prefixes. `ProxiBetApp.Components` and `ProxiBetApp.PageModels` are aggregated into this schema via `GlobalXmlns.cs` (project root, the XAML-xmlns counterpart to `GlobalUsings.cs`). Consequence: `FormField`, `CenteredFormLayout`, `ValidationMessage`, `LoginPageModel`, etc. are referenced without a prefix.

`xmlns:x` stays declared everywhere (required for `x:Class`/`x:Name`/`x:DataType`). Third-party package xmlns not covered by the global schema (e.g. `xmlns:toolkit` for CommunityToolkit.Maui on `MePage.xaml`) stay explicitly declared.

A new project namespace referenced from XAML must be added to `GlobalXmlns.cs`, not declared as a local xmlns in the consuming file.

## Pages/ migration status

- `Pages/LoginPage.xaml` — migrated: `CenteredFormLayout` + `FormField` + `ValidationMessage`.
- `Pages/RegisterPage.xaml` — migrated: `CenteredFormLayout` + `FormField` + `ValidationMessage`.
- `Pages/HomePage.xaml` — migrated: `<AuthenticatedPageLayout>` root (no attributes). Page-specific content: `VerticalStackLayout` (welcome Label, `ActivityIndicator`, `ValidationMessage`) + 2-column `Grid` with placeholder cards (see [DECISIONS.md](DECISIONS.md#future-components-placeholder-on-homepage)). `x:Name="HomeRoot"` and `Shell.NavBarIsVisible="False"` on the `ContentPage` (required by the `Appearing` `EventToCommandBehavior`'s `x:Reference HomeRoot`).
- `Pages/MePage.xaml` — migrated with a deliberate deviation: `<AuthenticatedPageLayout>` wraps the internal `VerticalStackLayout` (`ActivityIndicator`/`ValidationMessage`/user-info card) + a theme-toggle `ProxiCard` (native `Switch`, see [DECISIONS.md](DECISIONS.md#theme-toggle-migration-appheader--mepage)). `MePage` is the only page where the theme can be toggled — don't reintroduce `IsDarkMode`/`ToggleThemeCommand` on `AuthenticatedPageLayout` without revisiting this choice first. It's the `ShellContent` of the "Profil" tab in `AppShell.xaml`.
- `Pages/Dev/ComponentGalleryPage.xaml` — aligned on the same shell as authenticated pages: `<AuthenticatedPageLayout>` replaces the old bare root `ScrollView`. Accessible in Debug via a dedicated tab of the `TabBar` Shell (`AppShell.AddDebugTabs()`). No `BindingContext`/`x:DataType` (only displays component instances, no dynamic data). Out of scope for the acceptance rule (dev page, never shipped in Release), kept up to date manually in mirror of `Components/`.

Any new form page must use `CenteredFormLayout` + `FormField` from the start. Any new authenticated (post-login) page must start with `AuthenticatedPageLayout` (which already embeds `AppHeader`); navigation between sections goes through the native `TabBar` Shell (`AppShell.xaml`), not through this component. The theme toggle is not a responsibility of `AuthenticatedPageLayout`/`AppHeader`: `MePage` is the only page that exposes it.
