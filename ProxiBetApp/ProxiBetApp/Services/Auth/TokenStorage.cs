using System.Text.Json;

namespace ProxiBetApp.Services.Auth
{
    public sealed class TokenStorage : ITokenStorage
    {
        private const string StorageKey = "proxibet_auth";

        public async Task SaveAsync(StoredSession session)
        {
            var json = JsonSerializer.Serialize(session);
            await SecureStorage.Default.SetAsync(StorageKey, json);
        }

        public async Task<StoredSession?> LoadAsync()
        {
            try
            {
                var json = await SecureStorage.Default.GetAsync(StorageKey);
                if (string.IsNullOrEmpty(json))
                    return null;

                return JsonSerializer.Deserialize<StoredSession>(json);
            }
            catch (Exception)
            {
                // Corrupt/undecryptable entry (e.g. Android debug keystore rotated) - treat as no session.
                return null;
            }
        }

        public Task ClearAsync()
        {
            SecureStorage.Default.Remove(StorageKey);
            return Task.CompletedTask;
        }
    }
}
