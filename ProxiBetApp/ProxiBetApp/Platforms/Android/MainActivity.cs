using Android.App;
using Android.Content.PM;
using Android.OS;
using AndroidX.Core.View;
using MauiApplication = Microsoft.Maui.Controls.Application;

namespace ProxiBetApp
{
    [Activity(Theme = "@style/Maui.SplashTheme", MainLauncher = true, LaunchMode = LaunchMode.SingleTop, ConfigurationChanges = ConfigChanges.ScreenSize | ConfigChanges.Orientation | ConfigChanges.UiMode | ConfigChanges.ScreenLayout | ConfigChanges.SmallestScreenSize | ConfigChanges.Density)]
    public class MainActivity : MauiAppCompatActivity
    {
        // MAUI ne synchronise pas la couleur des icônes de la status bar OS avec UserAppTheme :
        // sans ça, les icônes système restent sombres et deviennent illisibles en dark mode.
        protected override void OnCreate(Bundle? savedInstanceState)
        {
            base.OnCreate(savedInstanceState);

            ApplyStatusBarAppearance(MauiApplication.Current?.RequestedTheme ?? AppTheme.Light);

            if (MauiApplication.Current is not null)
            {
                MauiApplication.Current.RequestedThemeChanged += (_, e) => ApplyStatusBarAppearance(e.RequestedTheme);
            }
        }

        private void ApplyStatusBarAppearance(AppTheme theme)
        {
            var window = Window;
            if (window is null)
            {
                return;
            }

            var insetsController = WindowCompat.GetInsetsController(window, window.DecorView);
            if (insetsController is not null)
            {
                insetsController.AppearanceLightStatusBars = theme != AppTheme.Dark;
            }
        }
    }
}
