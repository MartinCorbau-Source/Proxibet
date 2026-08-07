using System.Windows.Input;

namespace ProxiBetApp.Components
{
    public partial class ProxiButton : ContentView
    {
        public static readonly BindableProperty TextProperty = BindableProperty.Create(
            nameof(Text),
            typeof(string),
            typeof(ProxiButton),
            string.Empty);

        public static readonly BindableProperty CommandProperty = BindableProperty.Create(
            nameof(Command),
            typeof(ICommand),
            typeof(ProxiButton));

        public static readonly BindableProperty CommandParameterProperty = BindableProperty.Create(
            nameof(CommandParameter),
            typeof(object),
            typeof(ProxiButton));

        public static readonly BindableProperty VariantProperty = BindableProperty.Create(
            nameof(Variant),
            typeof(ButtonVariant),
            typeof(ProxiButton),
            ButtonVariant.Primary,
            propertyChanged: OnStyleAffectingPropertyChanged);

        public static readonly BindableProperty RadiusProperty = BindableProperty.Create(
            nameof(Radius),
            typeof(ButtonRadius),
            typeof(ProxiButton),
            ButtonRadius.Default,
            propertyChanged: OnStyleAffectingPropertyChanged);

        public static readonly BindableProperty FontFamilyProperty = BindableProperty.Create(
            nameof(FontFamily),
            typeof(string),
            typeof(ProxiButton));

        // Défaut -1 ("non défini") plutôt que 0 : un double non renseigné ne doit pas
        // écraser le FontSize du style résolu par Variant/Radius (voir UpdateStyle), donc
        // ce n'est appliqué à InnerButton que si explicitement fourni (>= 0).
        public static readonly BindableProperty FontSizeProperty = BindableProperty.Create(
            nameof(FontSize),
            typeof(double),
            typeof(ProxiButton),
            -1d,
            propertyChanged: OnStyleAffectingPropertyChanged);

        public static readonly BindableProperty ImageSourceProperty = BindableProperty.Create(
            nameof(ImageSource),
            typeof(ImageSource),
            typeof(ProxiButton));

        // Défaut Thickness(-1) ("non défini") plutôt que Thickness(0) : même raison que
        // FontSize (voir FontSizeProperty) — un Padding non renseigné ne doit pas écraser
        // celui du style résolu par Variant/Radius (ex. BaseButtonStyle "24,12").
        public static readonly BindableProperty PaddingProperty = BindableProperty.Create(
            nameof(Padding),
            typeof(Thickness),
            typeof(ProxiButton),
            new Thickness(-1),
            propertyChanged: OnStyleAffectingPropertyChanged);

        public static readonly BindableProperty ContentLayoutProperty = BindableProperty.Create(
            nameof(ContentLayout),
            typeof(Button.ButtonContentLayout),
            typeof(ProxiButton),
            new Button.ButtonContentLayout(Button.ButtonContentLayout.ImagePosition.Left, 10));

        public string Text
        {
            get => (string)GetValue(TextProperty);
            set => SetValue(TextProperty, value);
        }

        public ICommand? Command
        {
            get => (ICommand?)GetValue(CommandProperty);
            set => SetValue(CommandProperty, value);
        }

        public object? CommandParameter
        {
            get => GetValue(CommandParameterProperty);
            set => SetValue(CommandParameterProperty, value);
        }

        public ButtonVariant Variant
        {
            get => (ButtonVariant)GetValue(VariantProperty);
            set => SetValue(VariantProperty, value);
        }

        public ButtonRadius Radius
        {
            get => (ButtonRadius)GetValue(RadiusProperty);
            set => SetValue(RadiusProperty, value);
        }

        public string? FontFamily
        {
            get => (string?)GetValue(FontFamilyProperty);
            set => SetValue(FontFamilyProperty, value);
        }

        public double FontSize
        {
            get => (double)GetValue(FontSizeProperty);
            set => SetValue(FontSizeProperty, value);
        }

        public ImageSource? ImageSource
        {
            get => (ImageSource?)GetValue(ImageSourceProperty);
            set => SetValue(ImageSourceProperty, value);
        }

        public new Thickness Padding
        {
            get => (Thickness)GetValue(PaddingProperty);
            set => SetValue(PaddingProperty, value);
        }

        public Button.ButtonContentLayout ContentLayout
        {
            get => (Button.ButtonContentLayout)GetValue(ContentLayoutProperty);
            set => SetValue(ContentLayoutProperty, value);
        }

        public ProxiButton()
        {
            InitializeComponent();
            UpdateStyle();
        }

        static void OnStyleAffectingPropertyChanged(BindableObject bindable, object oldValue, object newValue)
        {
            if (bindable is ProxiButton proxiButton)
            {
                proxiButton.UpdateStyle();
            }
        }

        // Résout Variant+Radius vers l'un des 7 styles nommés de Buttons.xaml plutôt que de
        // composer CornerRadius/couleurs ici : ces styles restent la seule source de vérité
        // pour le rendu (états Pressed/PointerOver/Disabled inclus), ProxiButton ne fait que
        // sélectionner lequel s'applique. FontSize suit la même logique de non-écrasement :
        // seule une valeur explicitement fournie (>= 0) est poussée sur InnerButton, sinon
        // celle du style s'applique (voir note sur FontSizeProperty). Variant.Icon ignore Radius :
        // IconButtonStyle impose son propre CornerRadius fixe (voir Buttons.xaml), un bouton
        // icône n'a pas vocation à varier en Full/Medium/Small comme Primary/Secondary.
        void UpdateStyle()
        {
            var styleKey = (Variant, Radius) switch
            {
                (ButtonVariant.Primary, ButtonRadius.Default) => "PrimaryButtonStyle",
                (ButtonVariant.Primary, ButtonRadius.Medium) => "ButtonRoundedStyle",
                (ButtonVariant.Primary, ButtonRadius.Small) => "ButtonSquareStyle",
                (ButtonVariant.Secondary, ButtonRadius.Default) => "SecondaryButtonStyle",
                (ButtonVariant.Secondary, ButtonRadius.Medium) => "SecondaryButtonRoundedStyle",
                (ButtonVariant.Secondary, ButtonRadius.Small) => "SecondaryButtonSquareStyle",
                (ButtonVariant.Icon, _) => "IconButtonStyle",
                _ => "PrimaryButtonStyle"
            };

            InnerButton.Style = (Style)Application.Current!.Resources[styleKey];

            if (FontSize >= 0)
            {
                InnerButton.FontSize = FontSize;
            }

            if (Padding.Left >= 0 && Padding.Top >= 0 && Padding.Right >= 0 && Padding.Bottom >= 0)
            {
                InnerButton.Padding = Padding;
            }
        }
    }
}
