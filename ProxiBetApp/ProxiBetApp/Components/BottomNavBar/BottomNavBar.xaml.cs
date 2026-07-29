using System.Windows.Input;

namespace ProxiBetApp.Components
{
    public partial class BottomNavBar : ContentView
    {
        public static readonly BindableProperty GoToHomeCommandProperty = BindableProperty.Create(
            nameof(GoToHomeCommand),
            typeof(ICommand),
            typeof(BottomNavBar));

        public static readonly BindableProperty GoToGalleryCommandProperty = BindableProperty.Create(
            nameof(GoToGalleryCommand),
            typeof(ICommand),
            typeof(BottomNavBar));

        public ICommand? GoToHomeCommand
        {
            get => (ICommand?)GetValue(GoToHomeCommandProperty);
            set => SetValue(GoToHomeCommandProperty, value);
        }

        public ICommand? GoToGalleryCommand
        {
            get => (ICommand?)GetValue(GoToGalleryCommandProperty);
            set => SetValue(GoToGalleryCommandProperty, value);
        }

        public bool IsGalleryButtonVisible =>
#if DEBUG
            true;
#else
            false;
#endif

        public BottomNavBar()
        {
            InitializeComponent();
        }
    }
}
