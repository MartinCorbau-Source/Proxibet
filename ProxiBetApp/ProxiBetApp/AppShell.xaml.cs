using CommunityToolkit.Maui.Alerts;
using CommunityToolkit.Maui.Core;
using ProxiBetApp.Services.Theme;
using Font = Microsoft.Maui.Font;

namespace ProxiBetApp
{
    public partial class AppShell : Shell
    {
        public AppShell(IAuthService authService, IErrorHandler errorHandler)
        {
            InitializeComponent();
            RegisterRoutes();
            AddDebugTabs();
            CheckInitialRouteAsync(authService).FireAndForgetSafeAsync(errorHandler);
        }

        private static void RegisterRoutes()
        {
            Routing.RegisterRoute("register", typeof(RegisterPage));
        }

        private void AddDebugTabs()
        {
#if DEBUG
            var galleryTab = new Tab
            {
                Title = "Gallery",
                Route = "gallery",
                Icon = new FontImageSource
                {
                    FontFamily = "FluentUI",
                    Glyph = Fonts.FluentUI.beaker_24_regular
                }
            };
            galleryTab.Items.Add(new ShellContent
            {
                ContentTemplate = new DataTemplate(typeof(Pages.Dev.ComponentGalleryPage)),
                Route = "dev/gallery"
            });
            MainTabBar.Items.Add(galleryTab);
#endif
        }

        private async Task CheckInitialRouteAsync(IAuthService authService)
        {
            if (await authService.IsAuthenticatedAsync())
                await this.GoToAsync("//home");
        }

        public static async Task DisplaySnackbarAsync(string message)
        {
            CancellationTokenSource cancellationTokenSource = new CancellationTokenSource();

            var snackbarOptions = new SnackbarOptions
            {
                BackgroundColor = Color.FromArgb("#FF3300"),
                TextColor = Colors.White,
                ActionButtonTextColor = Colors.Yellow,
                CornerRadius = new CornerRadius(0),
                Font = Font.SystemFontOfSize(18),
                ActionButtonFont = Font.SystemFontOfSize(14)
            };

            var snackbar = Snackbar.Make(message, visualOptions: snackbarOptions);

            await snackbar.Show(cancellationTokenSource.Token);
        }

        public static async Task DisplayToastAsync(string message)
        {
            // Toast is currently not working in MCT on Windows
            if (OperatingSystem.IsWindows())
                return;

            var toast = Toast.Make(message, textSize: 18);

            var cts = new CancellationTokenSource(TimeSpan.FromSeconds(5));
            await toast.Show(cts.Token);
        }
    }
}
