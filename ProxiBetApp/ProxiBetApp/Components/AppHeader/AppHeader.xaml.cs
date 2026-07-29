using System.Windows.Input;

namespace ProxiBetApp.Components
{
    public partial class AppHeader : ContentView
    {
        public static readonly BindableProperty HeaderTitleProperty = BindableProperty.Create(
            nameof(HeaderTitle),
            typeof(string),
            typeof(AppHeader),
            string.Empty);

        public static readonly BindableProperty ProfileCommandProperty = BindableProperty.Create(
            nameof(ProfileCommand),
            typeof(ICommand),
            typeof(AppHeader));

        public static readonly BindableProperty ShowProfileIconProperty = BindableProperty.Create(
            nameof(ShowProfileIcon),
            typeof(bool),
            typeof(AppHeader),
            true);

        public static readonly BindableProperty ToggleThemeCommandProperty = BindableProperty.Create(
            nameof(ToggleThemeCommand),
            typeof(ICommand),
            typeof(AppHeader));

        public static readonly BindableProperty IsDarkModeProperty = BindableProperty.Create(
            nameof(IsDarkMode),
            typeof(bool),
            typeof(AppHeader),
            false);

        public string HeaderTitle
        {
            get => (string)GetValue(HeaderTitleProperty);
            set => SetValue(HeaderTitleProperty, value);
        }

        public ICommand? ProfileCommand
        {
            get => (ICommand?)GetValue(ProfileCommandProperty);
            set => SetValue(ProfileCommandProperty, value);
        }

        public bool ShowProfileIcon
        {
            get => (bool)GetValue(ShowProfileIconProperty);
            set => SetValue(ShowProfileIconProperty, value);
        }

        public ICommand? ToggleThemeCommand
        {
            get => (ICommand?)GetValue(ToggleThemeCommandProperty);
            set => SetValue(ToggleThemeCommandProperty, value);
        }

        public bool IsDarkMode
        {
            get => (bool)GetValue(IsDarkModeProperty);
            set => SetValue(IsDarkModeProperty, value);
        }

        public AppHeader()
        {
            InitializeComponent();
        }
    }
}
