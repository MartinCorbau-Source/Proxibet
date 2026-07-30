using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;

namespace ProxiBetApp.PageModels
{
    public partial class HomePageModel : PageModelBase
    {
        private readonly IAuthService _authService;
        private readonly CurrentUserStore _currentUserStore;

        public AuthenticatedUser? User => _currentUserStore.CurrentUser;

        public bool IsDevGalleryButtonVisible =>
#if DEBUG
            true;
#else
            false;
#endif

        public HomePageModel(IAuthService authService, CurrentUserStore currentUserStore)
        {
            _authService = authService;
            _currentUserStore = currentUserStore;

            _currentUserStore.PropertyChanged += (_, e) =>
            {
                if (e.PropertyName == nameof(Services.Auth.CurrentUserStore.CurrentUser))
                    OnPropertyChanged(nameof(User));
            };
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
        private static async Task GoToDevGalleryAsync()
        {
            await Shell.Current.GoToAsync("dev/gallery");
        }
    }
}
