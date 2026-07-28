using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;
using ProxiBetApp.Services.Theme;

namespace ProxiBetApp.PageModels
{
    public partial class MePageModel : PageModelBase
    {
        private readonly IAuthService _authService;
        private readonly IThemeService _themeService;

        [ObservableProperty]
        private AuthenticatedUser? _user;

        [ObservableProperty]
        private bool _isDarkMode;

        public MePageModel(IAuthService authService, IThemeService themeService)
        {
            _authService = authService;
            _themeService = themeService;
            IsDarkMode = _themeService.CurrentTheme == AppTheme.Dark;
        }

        [RelayCommand]
        private async Task ToggleThemeAsync()
        {
            await _themeService.ToggleThemeAsync();
            IsDarkMode = _themeService.CurrentTheme == AppTheme.Dark;
        }

        [RelayCommand]
        private async Task AppearingAsync()
        {
            await ExecuteWithErrorHandlingAsync(
                async () => User = await _authService.GetMeAsync(),
                onAuthError: async _ =>
                {
                    await _authService.LogoutAsync();
                    await Shell.Current.GoToAsync("//login");
                });
        }

        [RelayCommand]
        private async Task LogoutAsync()
        {
            await _authService.LogoutAsync();
            await Shell.Current.GoToAsync("//login");
        }

    }
}
