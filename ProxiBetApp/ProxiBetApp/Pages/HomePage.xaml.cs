namespace ProxiBetApp.Pages
{
    public partial class HomePage : ContentPage
    {
        public HomePage(HomePageModel model)
        {
            InitializeComponent();
            BindingContext = model;
        }
    }
}
