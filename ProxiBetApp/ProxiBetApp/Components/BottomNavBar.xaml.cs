using System.Windows.Input;

namespace ProxiBetApp.Components
{
    public partial class BottomNavBar : ContentView
    {
        public static readonly BindableProperty GoToHomeCommandProperty = BindableProperty.Create(
            nameof(GoToHomeCommand),
            typeof(ICommand),
            typeof(BottomNavBar));

        public ICommand? GoToHomeCommand
        {
            get => (ICommand?)GetValue(GoToHomeCommandProperty);
            set => SetValue(GoToHomeCommandProperty, value);
        }

        public BottomNavBar()
        {
            InitializeComponent();
        }
    }
}
