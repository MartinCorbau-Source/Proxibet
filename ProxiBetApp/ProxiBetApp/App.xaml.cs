using ProxiBetApp.Services.Theme;

namespace ProxiBetApp
{
    public partial class App : Application
    {
        private readonly AppShell _appShell;
        private readonly IThemeService _themeService;

        public App(AppShell appShell, IThemeService themeService)
        {
            InitializeComponent();
            _appShell = appShell;
            _themeService = themeService;
        }

        protected override Window CreateWindow(IActivationState? activationState)
        {
            // Application.Current n'est garanti assigné qu'à partir d'ici : appeler
            // ApplyPersistedTheme() plus tôt (ex. constructeur d'AppShell, résolu par DI
            // avant celui-ci) lève un NullReferenceException sur Application.Current.
            _themeService.ApplyPersistedTheme();
            return new Window(_appShell);
        }
    }
}