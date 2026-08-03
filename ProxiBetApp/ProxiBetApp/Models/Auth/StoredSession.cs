namespace ProxiBetApp.Models.Auth
{
    public sealed class StoredSession
    {
        public string AccessToken { get; set; } = string.Empty;
        public string RefreshToken { get; set; } = string.Empty;
        public DateTimeOffset ExpiresAtUtc { get; set; }
        public AuthenticatedUser User { get; set; } = new();
    }
}
