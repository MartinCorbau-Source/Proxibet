namespace ProxiBetApp.Components
{
    public partial class FormField : ContentView
    {
        public static readonly BindableProperty FieldLabelProperty = BindableProperty.Create(
            nameof(FieldLabel),
            typeof(string),
            typeof(FormField),
            string.Empty);

        public static readonly BindableProperty TextProperty = BindableProperty.Create(
            nameof(Text),
            typeof(string),
            typeof(FormField),
            string.Empty,
            BindingMode.TwoWay);

        public static readonly BindableProperty PlaceholderProperty = BindableProperty.Create(
            nameof(Placeholder),
            typeof(string),
            typeof(FormField),
            string.Empty);

        public static readonly BindableProperty KeyboardProperty = BindableProperty.Create(
            nameof(Keyboard),
            typeof(Keyboard),
            typeof(FormField),
            Keyboard.Default);

        public static readonly BindableProperty IsPasswordProperty = BindableProperty.Create(
            nameof(IsPassword),
            typeof(bool),
            typeof(FormField),
            false);

        public static readonly BindableProperty ErrorMessageProperty = BindableProperty.Create(
            nameof(ErrorMessage),
            typeof(string),
            typeof(FormField),
            string.Empty);

        public string FieldLabel
        {
            get => (string)GetValue(FieldLabelProperty);
            set => SetValue(FieldLabelProperty, value);
        }

        public string Text
        {
            get => (string)GetValue(TextProperty);
            set => SetValue(TextProperty, value);
        }

        public string Placeholder
        {
            get => (string)GetValue(PlaceholderProperty);
            set => SetValue(PlaceholderProperty, value);
        }

        public Keyboard Keyboard
        {
            get => (Keyboard)GetValue(KeyboardProperty);
            set => SetValue(KeyboardProperty, value);
        }

        public bool IsPassword
        {
            get => (bool)GetValue(IsPasswordProperty);
            set => SetValue(IsPasswordProperty, value);
        }

        public string ErrorMessage
        {
            get => (string)GetValue(ErrorMessageProperty);
            set => SetValue(ErrorMessageProperty, value);
        }

        public FormField()
        {
            InitializeComponent();
        }
    }
}
