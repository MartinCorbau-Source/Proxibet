using System.Collections.ObjectModel;

namespace ProxiBetApp.Components
{
    /// <summary>
    /// ContentProperty pointe vers FormContent pour permettre la syntaxe
    /// &lt;components:CenteredFormLayout&gt;&lt;Label/&gt;&lt;Entry/&gt;...&lt;/components:CenteredFormLayout&gt;
    /// (plusieurs enfants directs, comme sur les VerticalStackLayout qu'elle remplace).
    /// Le ScrollView interne (voir .xaml) doit rester qualifié via &lt;ContentView.Content&gt; :
    /// sans ça, ce même ContentProperty capte aussi le ScrollView du composant et le fait
    /// ajouter dans InnerLayout, qu'il contient déjà — cycle parent/enfant qui gèle le layout
    /// sans exception (bug déjà rencontré). La garde ci-dessous transforme une régression de
    /// ce type en erreur explicite au lieu d'un gel silencieux.
    /// </summary>
    [ContentProperty(nameof(FormContent))]
    public partial class CenteredFormLayout : ContentView
    {
        public ObservableCollection<View> FormContent { get; } = [];

        public CenteredFormLayout()
        {
            InitializeComponent();
            FormContent.CollectionChanged += (_, _) =>
            {
                InnerLayout.Children.Clear();
                foreach (var child in FormContent)
                {
                    if (IsAncestorOf(child, InnerLayout))
                        throw new InvalidOperationException(
                            $"CenteredFormLayout: '{child.GetType().Name}' fait partie de la structure " +
                            "interne du composant et ne peut pas être ajouté à FormContent (vérifier que " +
                            "le ScrollView du XAML interne est bien qualifié via <ContentView.Content>).");
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
