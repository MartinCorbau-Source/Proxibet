# CLAUDE.md — ProxiBetApp (.NET MAUI)

Ce fichier documente le design system de l'app mobile MAUI (`ProxiBetApp/ProxiBetApp/`). Il complète le `CLAUDE.md` racine du monorepo, qui couvre `proxiback/` et `proxifront/`.

## Principe d'acceptation du design system

**Un redesign visuel (couleurs, rayons, espacements, typographie) doit pouvoir se faire en éditant uniquement `Resources/Styles/` et/ou `Components/` — jamais `Pages/*.xaml`.**

Si un changement de design oblige à toucher une page, c'est le signe qu'un token ou un composant manque : il faut l'ajouter à la couche appropriée plutôt que de patcher la page.

Avant tout changement de token, valider le rendu sur `Pages/Dev/ComponentGalleryPage.xaml` (route `dev/gallery`, disponible uniquement en build Debug) : elle instancie chaque composant/variante et sert de banc de test visuel, plus rapide que de naviguer Login → Register → Me.

## Couche de tokens (`Resources/Styles/`)

Fichiers mergés dans `App.xaml`, dans cet ordre (l'ordre compte : les styles implicites plus tardifs dans la liste peuvent s'appuyer sur les précédents) :

1. `Colors.xaml` — échelles Primary/Secondary (10 niveaux, générées depuis un seed HSL), alias sémantiques (`Primary`, `Error`, `ErrorDark`...), tokens de fond Dark/Light, grayscale, brushes assortis.
2. `Spacing.xaml` — échelle numérique `size20`-`size560`, `IconSize(Small)` (OnIdiom), `LayoutPadding`/`LayoutSpacing` (OnIdiom), `FormMaxWidth`, `FormSpacing`.
3. `Typography.xaml` — style `Label` implicite + échelle keyée façon Fluent 2 (`Caption2/1/1Strong`, `Body2/2Strong/1/1Strong`, `Title3/2/1`, `LargeTitle`, `Display`).
4. `Buttons.xaml` — `BaseButtonStyle` (keyé, base commune), style `Button` implicite (BasedOn), `SecondaryButtonStyle` (BasedOn, variante transparente/lien).
5. `Inputs.xaml` — styles implicites `Entry`/`Editor`/`Picker`.
6. `Cards.xaml` — style `Border` implicite + `CardStyle` (keyé).
7. `Shell.xaml` — styles `Page`/`Shell`/`NavigationPage`/`TabbedPage`.
8. `Controls.xaml` — styles implicites pour les contrôles natifs restants (ActivityIndicator, Switch, CheckBox...).

**Règle dure** : toute couleur, taille, espacement ou rayon utilisé dans `Components/*.xaml` ou `Pages/*.xaml` doit référencer un `{StaticResource}` (éventuellement via `{AppThemeBinding}` pour le Dark/Light). Si la valeur nécessaire n'a pas encore de token, l'ajouter d'abord dans le fichier de tokens concerné (avec un commentaire expliquant sa provenance/son usage, voir `FormMaxWidth`/`FormSpacing` en exemple), puis le consommer — ne jamais coder une valeur en dur "en attendant".

## Couche de composants (`Components/`)

Composants composites réutilisables, au-dessus des styles implicites/keyés. Convention :

- Chaque composant = un `ContentView` en XAML + code-behind, exactement comme les pages (`ComponentName.xaml` + `ComponentName.xaml.cs`, `x:Class="ProxiBetApp.Components.ComponentName"`). Pas de C# markup.
- API publique = uniquement des `BindableProperty`. Une page consommatrice ne doit jamais overrider les enfants internes d'un composant depuis son propre XAML.
- Variantes visuelles exposées via une propriété bindable (enum si plusieurs valeurs), qui sélectionne en interne un `Style` nommé — jamais via des setters ad hoc côté appelant. C'est le principe déjà appliqué par `SecondaryButtonStyle` (un seul point d'édition pour tous les appelants).
- Composants "dumb"/présentationnels : pas de référence à un `PageModel`. Le binding avec la logique métier se fait uniquement via les `BindableProperty` exposées par le composant.
- Rester à plat dans `Components/` tant qu'il y a moins de ~8-10 fichiers ; sous-dossiers (`Components/Forms/`, `Components/Layout/`...) seulement quand une catégorie en justifie 3+.

Composants existants :
- `ValidationMessage` — message d'erreur utilisant les tokens `Error`/`ErrorDark`, se masque automatiquement quand `Text` est vide.
- `FormField` — `Label` + `Entry` + `ValidationMessage`, propriétés bindables `FieldLabel`, `Text` (TwoWay), `Placeholder`, `Keyboard`, `IsPassword`, `ErrorMessage`.
- `CenteredFormLayout` — shell `ScrollView > VerticalStackLayout` centré/largeur max pour les pages de formulaire ; accepte plusieurs enfants directs en XAML via `ContentProperty` (MAUI n'a pas de `ContentPresenter` hors `ControlTemplate`, d'où ce pattern par collection observable, voir le code-behind).
- `AppHeader` — en-tête des pages authentifiées : `Label` titre (style `Title2`) + bouton bascule thème (glyph FluentUI choisi via `Utilities.ThemeIconConverter` selon `IsDarkMode`) + bouton icône profil (glyph FluentUI `person_circle_24_regular`). Propriétés bindables `HeaderTitle` (string), `ProfileCommand` (`ICommand`), `ShowProfileIcon` (bool, défaut `true`, masque le bouton profil), `ToggleThemeCommand` (`ICommand`), `IsDarkMode` (bool). Embarqué en premier enfant du layout racine de chaque page authentifiée (`HomePage`, `MePage`) plutôt que via `Shell.TitleView`, qui est par-page/par-`ShellContent` et ne se partage pas automatiquement entre tabs Shell.

`CardStyle` reste volontairement un style `Border` (pas de `ContentView` dédié) tant qu'aucun usage n'a besoin de comportement (tap, overlay de chargement...) — à revisiter le jour où ce besoin apparaît, pas avant.

## Conventions de nommage

- PascalCase, pas de préfixe `Custom`/`My`.
- Paire de fichiers systématique `Nom.xaml` + `Nom.xaml.cs`.
- Commentaires de rationale en français dans le XAML (styles et composants), comme déjà pratiqué dans `Colors.xaml`/`Buttons.xaml`/`Spacing.xaml` — expliquer le *pourquoi* (quelle duplication ça remplace, quelle contrainte ça respecte), pas le *quoi*.
- `BindableProperty` : convention MAUI standard (`NomProperty` statique + accesseur `Nom`), pas de déviation.

## Couche de services (`Services/`)

Organisation feature-first : un domaine métier = un sous-dossier de `Services/` (ex. `Services/Auth/`), en miroir avec son équivalent `Models/<Domaine>/` (ex. `Models/Auth/`). À l'intérieur d'un domaine, tous les fichiers (interfaces, implémentations, exceptions, handlers, utilitaires) restent à plat — même règle de seuil que `Components/` (rester à plat tant qu'il y a moins de ~8-10 fichiers ; sous-dossiers par rôle technique, ex. `Auth/Network/`, `Auth/Session/`, seulement quand une catégorie en justifie 3+ **et** que le dossier dépasse ce seuil).

Les fichiers transverses à toute l'app (config partagée par plusieurs domaines comme `ApiConfig`, gestion d'erreur globale comme `IErrorHandler`/`ModalErrorHandler`) restent à la racine de `Services/`, jamais dans un sous-dossier de domaine — même si un seul domaine les consomme aujourd'hui, ils sont conceptuellement partagés.

**Pas de vertical slice** (`Features/<Domaine>/{PageModels,Services,Models}`) : les couches horizontales `Pages/` → `PageModels/` → `Services/`/`Models/` restent la structure de référence. Un vertical slice casserait cette frontière pour un bénéfice nul tant qu'un seul domaine métier a une taille significative — à reconsidérer seulement si plusieurs domaines deviennent chacun assez gros pour avoir leurs propres `PageModels`/`Models`/`Services` nombreux, pas préventivement.

## Convention xmlns

Tous les fichiers XAML (`Components/*.xaml`, `Pages/*.xaml`, `App.xaml`) utilisent le xmlns global .NET MAUI 10 `http://schemas.microsoft.com/dotnet/maui/global` comme xmlns par défaut, au lieu de l'ancien `http://schemas.microsoft.com/dotnet/2021/maui` + préfixes `xmlns:components`/`xmlns:pageModels` par fichier. `ProxiBetApp.Components` et `ProxiBetApp.PageModels` sont agrégés dans ce schema via `GlobalXmlns.cs` (à la racine du projet, pendant de `GlobalUsings.cs` mais pour les xmlns XAML plutôt que les `using` C#) ; `ProxiBetApp` et `ProxiBetApp.Pages` y sont déjà inclus par défaut. Conséquence : dans le XAML, `FormField`, `CenteredFormLayout`, `ValidationMessage`, `LoginPageModel`, etc. se référencent sans préfixe (`<FormField .../>`, `x:DataType="LoginPageModel"`).

`xmlns:x` reste déclaré partout (requis pour `x:Class`/`x:Name`/`x:DataType`). Les xmlns de packages tiers non couverts par le schema global (ex. `xmlns:toolkit` pour CommunityToolkit.Maui sur `MePage.xaml`) restent déclarés explicitement avec leur préfixe.

Un nouveau namespace projet référencé depuis le XAML doit être ajouté à `GlobalXmlns.cs`, pas déclaré en xmlns local dans le fichier consommateur.

**Choix délibéré** : `EnablePreviewFeatures`/`MauiAllowImplicitXmlnsDeclaration` (qui permettrait d'omettre aussi `xmlns`/`xmlns:x` racine) n'est pas activé — c'est une feature preview, le gain marginal ne justifie pas le risque de stabilité sur un projet qui n'est pas encore en prod.

## Migration en cours

- `Pages/LoginPage.xaml` — migrée : `CenteredFormLayout` + `FormField` + `ValidationMessage`.
- `Pages/RegisterPage.xaml` — migrée : `CenteredFormLayout` + `FormField` + `ValidationMessage`.
- `Pages/HomePage.xaml` — partiellement migrée : `AppHeader` en en-tête + `ValidationMessage`, mais garde un `VerticalStackLayout`/`ScrollView` natif pour le corps (pas de composant de layout de type formulaire applicable ici, ce n'est pas un écran de saisie) ; le binding `EventToCommandBehavior` sur `x:Reference HomeRoot` (Appearing → `AppearingCommand`) impose aussi de garder `x:Name`/`x:Reference` sur la page.
- `Pages/MePage.xaml` — partiellement migrée : `AppHeader` en en-tête (`ShowProfileIcon="False"`, déjà sur la page profil) + `ValidationMessage`, mais garde son `VerticalStackLayout` natif à cause du binding `EventToCommandBehavior` sur `x:Reference MeRoot`. N'est pas (encore) un `CardStyle`/`FormField` candidat pour la section infos utilisateur — à revisiter si un composant `InfoRow`/`ReadOnlyField` apparaît.
- `Pages/Dev/ComponentGalleryPage.xaml` — hors périmètre de la règle (page de dev, jamais en Release), tenue à jour manuellement en miroir de `Components/`.

Toute nouvelle page de formulaire doit utiliser `CenteredFormLayout` + `FormField` dès sa création plutôt que de dupliquer le pattern `ScrollView > VerticalStackLayout` à la main. Toute nouvelle page authentifiée (post-login) doit démarrer avec `AppHeader` en en-tête plutôt que de recréer un titre de page + icône profil/toggle de thème à la main.
