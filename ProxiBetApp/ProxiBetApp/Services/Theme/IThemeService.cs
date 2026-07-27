namespace ProxiBetApp.Services.Theme
{
    public interface IThemeService
    {
        AppTheme CurrentTheme { get; }

        void ApplyPersistedTheme();

        Task ToggleThemeAsync();
    }
}
