using System.ComponentModel;
using System.ComponentModel.DataAnnotations;
using System.Linq;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using ProxiBetApp.Services.Auth;

namespace ProxiBetApp.PageModels
{
    public partial class RegisterPageModel : ObservableValidator
    {
        private readonly IAuthService _authService;

        [ObservableProperty]
        [NotifyDataErrorInfo]
        [Required(ErrorMessage = "Le nom d'affichage est requis.")]
        [StringLength(100, MinimumLength = 2, ErrorMessage = "Le nom d'affichage doit contenir entre 2 et 100 caractères.")]
        private string _displayName = string.Empty;

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

        public string DisplayNameError => GetErrors(nameof(DisplayName)).FirstOrDefault()?.ErrorMessage ?? string.Empty;

        public string EmailError => GetErrors(nameof(Email)).FirstOrDefault()?.ErrorMessage ?? string.Empty;

        public string PasswordError => GetErrors(nameof(Password)).FirstOrDefault()?.ErrorMessage ?? string.Empty;

        public RegisterPageModel(IAuthService authService)
        {
            _authService = authService;
        }

        protected override void OnPropertyChanged(PropertyChangedEventArgs e)
        {
            base.OnPropertyChanged(e);
            if (e.PropertyName == nameof(DisplayName))
                OnPropertyChanged(nameof(DisplayNameError));
            if (e.PropertyName == nameof(Email))
                OnPropertyChanged(nameof(EmailError));
            if (e.PropertyName == nameof(Password))
                OnPropertyChanged(nameof(PasswordError));
        }

        [RelayCommand]
        private async Task RegisterAsync()
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
