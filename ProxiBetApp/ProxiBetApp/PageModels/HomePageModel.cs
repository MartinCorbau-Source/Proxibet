using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;
using ProxiBetApp.Services.Theme;

namespace ProxiBetApp.PageModels
{
    public partial class HomePageModel : PageModelBase
    {
        private readonly IAuthService _authService;
        private readonly IThemeService _themeService;
        private readonly CurrentUserStore _currentUserStore;

        [ObservableProperty]
        private bool _isDarkMode;

        public AuthenticatedUser? User => _currentUserStore.CurrentUser;

        public bool IsDevGalleryButtonVisible =>
#if DEBUG
            true;
#else
            false;
#endif

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
            await ExecuteWithErrorHandlingAsync(
                () => _authService.GetMeAsync(),
                onAuthError: async _ =>
                {
                    await _authService.LogoutAsync();
                    await Shell.Current.GoToAsync("//login");
                });
        }

        [RelayCommand]
        private static async Task GoToProfileAsync()
        {
            await Shell.Current.GoToAsync("me");
        }

        [RelayCommand]
        private static async Task GoToDevGalleryAsync()
        {
            await Shell.Current.GoToAsync("dev/gallery");
        }

        [RelayCommand]
        private async Task ToggleThemeAsync()
        {
            await _themeService.ToggleThemeAsync();
            IsDarkMode = _themeService.CurrentTheme == AppTheme.Dark;
        }
    }
}
