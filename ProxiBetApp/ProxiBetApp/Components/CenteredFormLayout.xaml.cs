using System.Collections.ObjectModel;

namespace ProxiBetApp.Components
{
    /// <summary>
    /// ContentProperty pointe vers FormContent pour permettre la syntaxe
    /// &lt;components:CenteredFormLayout&gt;&lt;Label/&gt;&lt;Entry/&gt;...&lt;/components:CenteredFormLayout&gt;
    /// (plusieurs enfants directs, comme sur les VerticalStackLayout qu'elle remplace).
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
                    InnerLayout.Children.Add(child);
            };
        }
    }
}
