using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;

namespace ProxiBetApp.PageModels
{
    public partial class LoginPageModel : ObservableObject
    {
        private readonly IAuthService _authService;

        [ObservableProperty]
        private string _email = string.Empty;

        [ObservableProperty]
        private string _password = string.Empty;

        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string? _errorMessage;

        public LoginPageModel(IAuthService authService)
        {
            _authService = authService;
        }

        [RelayCommand]
        private async Task LoginAsync()
        {
            if (IsBusy)
                return;

            IsBusy = true;
            ErrorMessage = null;

            try
            {
                await _authService.LoginAsync(Email, Password);
                await AppShell.DisplayToastAsync("Connexion réussie");
                await Shell.Current.GoToAsync("/me");
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
        private static async Task GoToRegisterAsync()
        {
            await Shell.Current.GoToAsync("register");
        }
    }
}
