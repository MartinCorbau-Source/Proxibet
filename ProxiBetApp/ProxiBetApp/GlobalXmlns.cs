using Microsoft.Maui.Controls.Xaml;

// Agrège les namespaces projet dans le xmlns global MAUI 10 (http://schemas.microsoft.com/dotnet/maui/global),
// pour éviter de répéter xmlns:components/xmlns:pageModels dans chaque fichier XAML.
// ProxiBetApp et ProxiBetApp.Pages y sont déjà inclus par défaut.
[assembly: XmlnsDefinition("http://schemas.microsoft.com/dotnet/maui/global", "ProxiBetApp.Components")]
[assembly: XmlnsDefinition("http://schemas.microsoft.com/dotnet/maui/global", "ProxiBetApp.PageModels")]
