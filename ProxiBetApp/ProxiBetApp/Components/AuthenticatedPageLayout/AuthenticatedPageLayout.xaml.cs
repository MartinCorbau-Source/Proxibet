using System.Collections.ObjectModel;
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
        public ObservableCollection<View> PageContent { get; } = [];

        public AuthenticatedPageLayout()
        {
            InitializeComponent();

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
    }
}
