namespace ProxiBetApp.Services.Theme
{
    public sealed class ThemeService : IThemeService
    {
        private const string ThemePreferenceKey = "app_theme";

        public AppTheme CurrentTheme => Application.Current!.UserAppTheme;

        public void ApplyPersistedTheme()
        {
            var stored = Preferences.Default.Get(ThemePreferenceKey, nameof(AppTheme.Light));
            Application.Current!.UserAppTheme = stored == nameof(AppTheme.Dark) ? AppTheme.Dark : AppTheme.Light;
        }

        public Task ToggleThemeAsync()
        {
            var nextTheme = CurrentTheme == AppTheme.Dark ? AppTheme.Light : AppTheme.Dark;

            Application.Current!.UserAppTheme = nextTheme;
            Preferences.Default.Set(ThemePreferenceKey, nextTheme.ToString());

            return Task.CompletedTask;
        }
    }
}
