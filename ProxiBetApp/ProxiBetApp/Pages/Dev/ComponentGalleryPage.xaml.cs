using System.Windows.Input;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;

namespace ProxiBetApp.Pages.Dev
{
    public partial class ComponentGalleryPage : ContentPage
    {
        public ComponentGalleryPage()
        {
            InitializeComponent();
            BindingContext = new ComponentGalleryPageModel();
        }
    }

    // BindingContext minimal, propre à cette page de démo (pas de service/DI requis
    // pour exercer visuellement AppHeader, contrairement aux vraies pages authentifiées).
    public partial class ComponentGalleryPageModel : ObservableObject
    {
        [ObservableProperty]
        private bool _isDarkMode;

        public ICommand NoOpCommand { get; } = new Command(() => { });

        [RelayCommand]
        private void ToggleTheme()
        {
            IsDarkMode = !IsDarkMode;
        }
    }
}
