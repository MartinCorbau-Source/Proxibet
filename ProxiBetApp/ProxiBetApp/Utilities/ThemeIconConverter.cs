using System.Globalization;

namespace ProxiBetApp.Utilities
{
    // Bascule le glyph FluentUI affiché selon le thème courant : lune = "passer en sombre",
    // soleil = "passer en clair" (l'icône représente la destination du tap, pas l'état actuel).
    public sealed class ThemeIconConverter : IValueConverter
    {
        public object Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
        {
            var isDarkMode = value is bool b && b;
            return isDarkMode ? FluentUI.brightness_high_24_regular : FluentUI.weather_moon_24_regular;
        }

        public object ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
        {
            throw new NotSupportedException();
        }
    }
}
