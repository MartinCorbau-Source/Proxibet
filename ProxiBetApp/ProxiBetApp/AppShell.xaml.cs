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
            CheckInitialRouteAsync(authService).FireAndForgetSafeAsync(errorHandler);
        }

        private static void RegisterRoutes()
        {
            Routing.RegisterRoute("register", typeof(RegisterPage));
            Routing.RegisterRoute("me", typeof(MePage));
#if DEBUG
            Routing.RegisterRoute("dev/gallery", typeof(Pages.Dev.ComponentGalleryPage));
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
