using System.Collections.ObjectModel;
using System.Windows.Input;
using ProxiBetApp.Utilities;

namespace ProxiBetApp.Components
{
    /// <summary>
    /// ContentProperty pointe vers PageContent, même mécanisme que CenteredFormLayout.FormContent.
    /// Le Grid interne (voir .xaml) doit rester qualifié via &lt;ContentView.Content&gt; : sans ça,
    /// ce ContentProperty capterait aussi le Grid racine et le re-parenterait dans InnerLayout
    /// (gel de layout sans exception). Gardé par le même IsAncestorOf que CenteredFormLayout.
    /// </summary>
    [ContentProperty(nameof(PageContent))]
    public partial class AuthenticatedPageLayout : ContentView
    {
        public static readonly BindableProperty ProfileCommandProperty = BindableProperty.Create(
            nameof(ProfileCommand), typeof(ICommand), typeof(AuthenticatedPageLayout));

        public static readonly BindableProperty ShowProfileIconProperty = BindableProperty.Create(
            nameof(ShowProfileIcon), typeof(bool), typeof(AuthenticatedPageLayout), true);

        public static readonly BindableProperty ToggleThemeCommandProperty = BindableProperty.Create(
            nameof(ToggleThemeCommand), typeof(ICommand), typeof(AuthenticatedPageLayout));

        public static readonly BindableProperty IsDarkModeProperty = BindableProperty.Create(
            nameof(IsDarkMode), typeof(bool), typeof(AuthenticatedPageLayout), false);

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

        public ObservableCollection<View> PageContent { get; } = [];

        public AuthenticatedPageLayout()
        {
            InitializeComponent();

            BottomNav.GoToHomeCommand = new Command(() => GoToHomeAsync().FireAndForgetSafeAsync());
            BottomNav.GoToGalleryCommand = new Command(() => GoToGalleryAsync().FireAndForgetSafeAsync());

            PageContent.CollectionChanged += (_, _) =>
            {
                InnerLayout.Children.Clear();
                foreach (var child in PageContent)
                {
                    if (IsAncestorOf(child, InnerLayout))
                        throw new InvalidOperationException(
                            $"AuthenticatedPageLayout: '{child.GetType().Name}' fait partie de la structure " +
                            "interne du composant et ne peut pas être ajouté à PageContent (vérifier que " +
                            "le Grid du XAML interne est bien qualifié via <ContentView.Content>).");
                    InnerLayout.Children.Add(child);
                }
            };

            // Les insets système (zone de geste Android en edge-to-edge) sont dispatchés à la
            // Page hôte, pas à un ContentView imbriqué : positionner SafeAreaEdges ici plutôt
            // que dans BottomNavBar, qui ne les reçoit jamais. Bord bas uniquement, pour que le
            // fond de BottomNavBar (Grid.Row="1") s'étende jusqu'au bord physique de l'écran.
            ParentChanged += (_, _) =>
            {
                if (FindHostPage(this) is ContentPage hostPage)
                {
                    hostPage.SafeAreaEdges = new SafeAreaEdges(SafeAreaRegions.None, SafeAreaRegions.None, SafeAreaRegions.None, SafeAreaRegions.All);
                }
            };
        }

        private static async Task GoToHomeAsync()
        {
            if (Shell.Current is not null)
                await Shell.Current.GoToAsync("//home");
        }

        private static async Task GoToGalleryAsync()
        {
            if (Shell.Current is not null)
                await Shell.Current.GoToAsync("dev/gallery");
        }

        private static bool IsAncestorOf(Element candidate, Element node)
        {
            for (var current = node.Parent; current is not null; current = current.Parent)
            {
                if (ReferenceEquals(current, candidate))
                    return true;
            }
            return false;
        }

        private static Page? FindHostPage(Element node)
        {
            for (var current = node.Parent; current is not null; current = current.Parent)
            {
                if (current is Page page)
                    return page;
            }
            return null;
        }
    }
}
