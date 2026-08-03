namespace ProxiBetApp.Pages
{
    public partial class MePage : ContentPage
    {
        public MePage(MePageModel model)
        {
            InitializeComponent();
            BindingContext = model;
        }
    }
}
