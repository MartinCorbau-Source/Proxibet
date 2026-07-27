using System.ComponentModel;
using System.ComponentModel.DataAnnotations;
using System.Linq;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;

namespace ProxiBetApp.PageModels
{
    public partial class LoginPageModel : ObservableValidator
    {
        private readonly IAuthService _authService;

        [ObservableProperty]
        [NotifyDataErrorInfo]
        [Required(ErrorMessage = "L'email est requis.")]
        [EmailAddress(ErrorMessage = "Format d'email invalide.")]
        private string _email = string.Empty;

        [ObservableProperty]
        [NotifyDataErrorInfo]
        [Required(ErrorMessage = "Le mot de passe est requis.")]
        [MinLength(8, ErrorMessage = "Le mot de passe doit contenir au moins 8 caractères.")]
        private string _password = string.Empty;

        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string? _errorMessage;

        public string EmailError => GetErrors(nameof(Email)).FirstOrDefault()?.ErrorMessage ?? string.Empty;

        public string PasswordError => GetErrors(nameof(Password)).FirstOrDefault()?.ErrorMessage ?? string.Empty;

        public bool IsDevGalleryButtonVisible =>
#if DEBUG
            true;
#else
            false;
#endif

        public LoginPageModel(IAuthService authService)
        {
            _authService = authService;
        }

        protected override void OnPropertyChanged(PropertyChangedEventArgs e)
        {
            base.OnPropertyChanged(e);
            if (e.PropertyName == nameof(Email))
                OnPropertyChanged(nameof(EmailError));
            if (e.PropertyName == nameof(Password))
                OnPropertyChanged(nameof(PasswordError));
        }

        [RelayCommand]
        private async Task LoginAsync()
        {
            if (IsBusy)
                return;

            ValidateAllProperties();
            if (HasErrors)
                return;

            IsBusy = true;
            ErrorMessage = null;

            try
            {
                await _authService.LoginAsync(Email, Password);
                await AppShell.DisplayToastAsync("Connexion réussie");
                await Shell.Current.GoToAsync("//home");
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

        [RelayCommand]
        private static async Task GoToDevGalleryAsync()
        {
            await Shell.Current.GoToAsync("dev/gallery");
        }
    }
}
