namespace ProxiBetApp.Components
{
    public partial class ValidationMessage : ContentView
    {
        public static readonly BindableProperty TextProperty = BindableProperty.Create(
            nameof(Text),
            typeof(string),
            typeof(ValidationMessage),
            string.Empty,
            propertyChanged: OnTextChanged);

        public string Text
        {
            get => (string)GetValue(TextProperty);
            set => SetValue(TextProperty, value);
        }

        public ValidationMessage()
        {
            InitializeComponent();
        }

        private static void OnTextChanged(BindableObject bindable, object oldValue, object newValue)
        {
            var control = (ValidationMessage)bindable;
            var text = (string?)newValue ?? string.Empty;

            control.MessageLabel.Text = text;
            control.IsVisible = !string.IsNullOrWhiteSpace(text);
        }
    }
}
