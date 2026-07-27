using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;
using ProxiBetApp.Services.Theme;

namespace ProxiBetApp.PageModels
{
    public partial class HomePageModel : ObservableObject
    {
        private readonly IAuthService _authService;
        private readonly IThemeService _themeService;
        private readonly CurrentUserStore _currentUserStore;

        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string? _errorMessage;

        [ObservableProperty]
        private bool _isDarkMode;

        public AuthenticatedUser? User => _currentUserStore.CurrentUser;

        public HomePageModel(IAuthService authService, IThemeService themeService, CurrentUserStore currentUserStore)
        {
            _authService = authService;
            _themeService = themeService;
            _currentUserStore = currentUserStore;

            _currentUserStore.PropertyChanged += (_, e) =>
            {
                if (e.PropertyName == nameof(Services.Auth.CurrentUserStore.CurrentUser))
                    OnPropertyChanged(nameof(User));
            };

            IsDarkMode = _themeService.CurrentTheme == AppTheme.Dark;
        }

        [RelayCommand]
        private async Task AppearingAsync()
        {
            IsBusy = true;
            ErrorMessage = null;

            try
            {
                await _authService.GetMeAsync();
            }
            catch (AuthApiException)
            {
                await _authService.LogoutAsync();
                await Shell.Current.GoToAsync("//login");
            }
            finally
            {
                IsBusy = false;
            }
        }

        [RelayCommand]
        private static async Task GoToProfileAsync()
        {
            await Shell.Current.GoToAsync("me");
        }

        [RelayCommand]
        private static async Task GoToHomeAsync()
        {
            await Shell.Current.GoToAsync("//home");
        }

        [RelayCommand]
        private async Task ToggleThemeAsync()
        {
            await _themeService.ToggleThemeAsync();
            IsDarkMode = _themeService.CurrentTheme == AppTheme.Dark;
        }
    }
}
