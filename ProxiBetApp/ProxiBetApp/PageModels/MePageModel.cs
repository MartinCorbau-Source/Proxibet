using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;

namespace ProxiBetApp.PageModels
{
    public partial class MePageModel : ObservableObject
    {
        private readonly IAuthService _authService;

        [ObservableProperty]
        private AuthenticatedUser? _user;

        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string? _errorMessage;

        public MePageModel(IAuthService authService)
        {
            _authService = authService;
        }

        [RelayCommand]
        private async Task AppearingAsync()
        {
            IsBusy = true;
            ErrorMessage = null;

            try
            {
                User = await _authService.GetMeAsync();
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
        private async Task LogoutAsync()
        {
            await _authService.LogoutAsync();
            await Shell.Current.GoToAsync("//login");
        }
    }
}
