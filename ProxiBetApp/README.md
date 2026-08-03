# ProxiBetApp — lancer l'app mobile

App .NET MAUI (`ProxiBetApp.slnx`) qui consomme l'API `proxiback/`. Ce README couvre le lancement depuis Visual Studio, en local (Windows) ou sur un téléphone Android branché en USB.

## Prérequis

- Visual Studio 2022 (17.14+) avec la charge de travail **.NET Multi-platform App UI development**.
- Le backend doit tourner via Docker avant de lancer l'app : à la racine du repo, `make docker-up` (ou `docker compose up -d`). Vérifier avec `docker compose ps` que `proxibet-api` est `Up`.

Ouvrir `ProxiBetApp/ProxiBetApp.slnx` dans Visual Studio.

## Lancer sur Windows

En haut de Visual Studio, choisir la target **Windows Machine** dans le sélecteur à côté du bouton ▶️, puis lancer (F5 ou ▶️).

`ApiConfig.BaseUrl` (`ProxiBetApp/Services/ApiConfig.cs`) vaut `http://localhost:8080` en Debug — sur Windows, `localhost` atteint directement Docker sur la même machine, aucune config supplémentaire n'est nécessaire.

## Lancer sur un téléphone Android branché en USB

1. Sur le téléphone : activer les **options développeur** puis le **débogage USB**, brancher en USB, accepter l'invite d'autorisation qui apparaît sur l'écran.
2. Dans Visual Studio, le sélecteur de target (à côté de `ProxiBetApp`) doit lister l'appareil détecté (ex. *Google Pixel 8 Pro (Android 16.0 - API 36)*) — le choisir.
3. Lancer (F5 ou ▶️).

Le backend Docker tourne sur la machine, pas sur le téléphone — comme le téléphone est physique, `localhost` dans l'app désignerait le téléphone lui-même, pas la machine. Le `.csproj` (`ProxiBetApp/ProxiBetApp.csproj`) déclenche automatiquement `adb reverse tcp:8080 tcp:8080` à chaque déploiement Android en Debug (target MSBuild `AdbReverseBackendPort`), donc il n'y a rien à faire manuellement pour ça.

### Si l'app n'arrive pas à joindre le backend malgré tout

- Vérifier que Docker tourne : `docker compose ps` à la racine du repo, `proxibet-api` doit être `Up`.
- Vérifier que le téléphone est bien vu par adb : `adb devices` doit le lister avec le statut `device` (pas `unauthorized`/`offline` — dans ce cas, redébloquer l'invite de débogage USB sur le téléphone).
- Redéployer depuis Visual Studio (F5) : ça relance le hook `adb reverse` automatiquement.
- En dernier recours, relancer le mapping à la main : `adb reverse tcp:8080 tcp:8080`.

## Tests

Le projet `ProxiBetApp.Tests` (xUnit) se lance via l'Explorateur de tests de Visual Studio, ou `dotnet test ProxiBetApp/ProxiBetApp.Tests`.
