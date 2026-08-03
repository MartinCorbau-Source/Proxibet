using CommunityToolkit.Maui;
using Microsoft.Extensions.Logging;
using ProxiBetApp.Pages;
using ProxiBetApp.PageModels;
using ProxiBetApp.Services;
using ProxiBetApp.Services.Auth;
using ProxiBetApp.Services.Theme;
using Refit;
#if DEBUG
using Plugin.Maui.DebugRainbows;
#endif

namespace ProxiBetApp
{
    public static class MauiProgram
    {
        public static MauiApp CreateMauiApp()
        {
            var builder = MauiApp.CreateBuilder();
            builder
                .UseMauiApp<App>()
                .UseMauiCommunityToolkit()
                .ConfigureFonts(fonts =>
                {
                    fonts.AddFont("OpenSans-Regular.ttf", "OpenSansRegular");
                    fonts.AddFont("OpenSans-Semibold.ttf", "OpenSansSemibold");
                    fonts.AddFont("SegoeUI-Semibold.ttf", "SegoeSemibold");
                    fonts.AddFont("FluentSystemIcons-Regular.ttf", FluentUI.FontFamily);
                });

#if DEBUG
            builder.Logging.AddDebug();
            builder.Services.AddLogging(configure => configure.AddDebug());
            // builder.UseDebugRainbows(); // désactivé temporairement pour voir le vrai rendu visuel
#endif

            builder.Services.AddSingleton<IErrorHandler, ModalErrorHandler>();
            builder.Services.AddSingleton<ITokenStorage, TokenStorage>();
            builder.Services.AddSingleton<CurrentUserStore>();
            builder.Services.AddSingleton<IThemeService, ThemeService>();
            builder.Services.AddTransient<AuthHeaderHandler>();

            builder.Services.AddHttpClient("proxibet-unauthenticated", client =>
            {
                client.BaseAddress = new Uri(ApiConfig.BaseUrl);
            });

            builder.Services.AddRefitClient<IAuthApi>()
                .ConfigureHttpClient(client => client.BaseAddress = new Uri(ApiConfig.BaseUrl))
                .AddHttpMessageHandler<AuthHeaderHandler>();

            builder.Services.AddTransient<IAuthService, AuthService>();

            builder.Services.AddTransient<LoginPageModel>();
            builder.Services.AddTransient<RegisterPageModel>();
            builder.Services.AddTransient<MePageModel>();
            builder.Services.AddTransient<HomePageModel>();

            builder.Services.AddTransient<LoginPage>();
            builder.Services.AddTransient<RegisterPage>();
            builder.Services.AddTransient<MePage>();
            builder.Services.AddTransient<HomePage>();

            builder.Services.AddTransient<AppShell>();

            return builder.Build();
        }
    }
}
