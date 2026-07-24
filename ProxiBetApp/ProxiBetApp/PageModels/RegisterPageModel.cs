using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;

namespace ProxiBetApp.PageModels
{
    public partial class RegisterPageModel : ObservableObject
    {
        private readonly IAuthService _authService;

        [ObservableProperty]
        private string _displayName = string.Empty;

        [ObservableProperty]
        private string _email = string.Empty;

        [ObservableProperty]
        private string _password = string.Empty;

        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string? _errorMessage;

        public RegisterPageModel(IAuthService authService)
        {
            _authService = authService;
        }

        [RelayCommand]
        private async Task RegisterAsync()
        {
            if (IsBusy)
                return;

            IsBusy = true;
            ErrorMessage = null;

            try
            {
                await _authService.RegisterAsync(DisplayName, Email, Password);
                await AppShell.DisplayToastAsync("Compte créé");
                await Shell.Current.GoToAsync("//login");
            }
            catch (AuthApiException ex)
            {
                ErrorMessage = ex.Message;
            }
            finally
            {
                IsBusy = false;
            }
        }

        [RelayCommand]
        private static async Task GoToLoginAsync()
        {
            await Shell.Current.GoToAsync("//login");
        }
    }
}
