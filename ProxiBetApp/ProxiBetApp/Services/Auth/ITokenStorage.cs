namespace ProxiBetApp.Services.Auth
{
    public interface ITokenStorage
    {
        Task SaveAsync(StoredSession session);
        Task<StoredSession?> LoadAsync();
        Task ClearAsync();
    }
}
