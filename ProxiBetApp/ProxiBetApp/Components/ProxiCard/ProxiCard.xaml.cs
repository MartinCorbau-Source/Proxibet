using System.Collections.ObjectModel;

namespace ProxiBetApp.Components
{
    /// <summary>
    /// ContentProperty pointe vers CardContent, même mécanisme que CenteredFormLayout.FormContent
    /// / AuthenticatedPageLayout.PageContent. Le Border interne (voir .xaml) doit rester qualifié
    /// via &lt;ContentView.Content&gt; : sans ça, ce ContentProperty capterait aussi le Border et
    /// le re-parenterait dans InnerLayout (gel de layout sans exception). Gardé par le même
    /// IsAncestorOf que les deux composants existants.
    /// </summary>
    [ContentProperty(nameof(CardContent))]
    public partial class ProxiCard : ContentView
    {
        public static readonly BindableProperty ColorProperty = BindableProperty.Create(
            nameof(Color),
            typeof(CardColor),
            typeof(ProxiCard),
            CardColor.Default,
            propertyChanged: OnStyleAffectingPropertyChanged);

        public static readonly BindableProperty HasShadowProperty = BindableProperty.Create(
            nameof(HasShadow),
            typeof(bool),
            typeof(ProxiCard),
            true,
            propertyChanged: OnStyleAffectingPropertyChanged);

        // Défaut -1 ("non défini") : applique CardMaxWidth (Resources/Styles/Spacing.xaml)
        // tant que la page consommatrice ne fournit pas explicitement une valeur >= 0, même
        // logique de non-écrasement que ProxiButton.FontSizeProperty.
        public static readonly BindableProperty MaxWidthProperty = BindableProperty.Create(
            nameof(MaxWidth),
            typeof(double),
            typeof(ProxiCard),
            -1d,
            propertyChanged: OnMaxWidthChanged);

        public CardColor Color
        {
            get => (CardColor)GetValue(ColorProperty);
            set => SetValue(ColorProperty, value);
        }

        public bool HasShadow
        {
            get => (bool)GetValue(HasShadowProperty);
            set => SetValue(HasShadowProperty, value);
        }

        public double MaxWidth
        {
            get => (double)GetValue(MaxWidthProperty);
            set => SetValue(MaxWidthProperty, value);
        }

        public ObservableCollection<View> CardContent { get; } = [];

        public ProxiCard()
        {
            InitializeComponent();
            UpdateStyle();
            UpdateMaxWidth();

            CardContent.CollectionChanged += (_, _) =>
            {
                InnerLayout.Children.Clear();
                foreach (var child in CardContent)
                {
                    if (IsAncestorOf(child, InnerLayout))
                        throw new InvalidOperationException(
                            $"ProxiCard: '{child.GetType().Name}' fait partie de la structure " +
                            "interne du composant et ne peut pas être ajouté à CardContent (vérifier que " +
                            "le Border du XAML interne est bien qualifié via <ContentView.Content>).");
                    InnerLayout.Children.Add(child);
                }
            };
        }

        static void OnStyleAffectingPropertyChanged(BindableObject bindable, object oldValue, object newValue)
        {
            if (bindable is ProxiCard card)
            {
                card.UpdateStyle();
            }
        }

        static void OnMaxWidthChanged(BindableObject bindable, object oldValue, object newValue)
        {
            if (bindable is ProxiCard card)
            {
                card.UpdateMaxWidth();
            }
        }

        // Résout Color vers CardStyle/CardPrimaryStyle (Resources/Styles/Cards.xaml), qui
        // portent déjà StrokeShape/Padding/Background/Shadow. HasShadow=false retire le Shadow
        // posé par le style plutôt que de le reconstruire : pas besoin de détecter le thème
        // actif en code-behind, l'AppThemeBinding du Style suffit tant qu'on ne fait que le
        // retirer/laisser tel quel.
        void UpdateStyle()
        {
            var styleKey = Color switch
            {
                CardColor.Primary => "CardPrimaryStyle",
                _ => "CardStyle"
            };

            InnerBorder.Style = (Style)Application.Current!.Resources[styleKey];

            if (!HasShadow)
            {
                InnerBorder.Shadow = null!;
            }
        }

        void UpdateMaxWidth()
        {
            InnerBorder.MaximumWidthRequest = MaxWidth >= 0
                ? MaxWidth
                : (double)Application.Current!.Resources["CardMaxWidth"];
        }

        static bool IsAncestorOf(Element candidate, Element node)
        {
            for (var current = node.Parent; current is not null; current = current.Parent)
            {
                if (ReferenceEquals(current, candidate))
                    return true;
            }
            return false;
        }
    }
}
