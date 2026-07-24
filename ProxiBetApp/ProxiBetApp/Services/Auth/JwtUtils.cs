using System.Text;
using System.Text.Json;

namespace ProxiBetApp.Services.Auth
{
    public static class JwtUtils
    {
        public static DateTimeOffset? TryGetExpiryUtc(string jwt)
        {
            var parts = jwt.Split('.');
            if (parts.Length != 3)
                return null;

            try
            {
                var payloadJson = Encoding.UTF8.GetString(Base64UrlDecode(parts[1]));
                using var document = JsonDocument.Parse(payloadJson);
                if (document.RootElement.TryGetProperty("exp", out var expElement) && expElement.TryGetInt64(out var exp))
                    return DateTimeOffset.FromUnixTimeSeconds(exp);
            }
            catch (Exception ex) when (ex is FormatException or JsonException)
            {
                return null;
            }

            return null;
        }

        private static byte[] Base64UrlDecode(string input)
        {
            var base64 = input.Replace('-', '+').Replace('_', '/');
            switch (base64.Length % 4)
            {
                case 2: base64 += "=="; break;
                case 3: base64 += "="; break;
            }
            return Convert.FromBase64String(base64);
        }
    }
}
