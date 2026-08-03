namespace ProxiBetApp.Services.Auth
{
    public interface IAuthService
    {
        Task<AuthenticatedUser> RegisterAsync(string displayName, string email, string password);
        Task LoginAsync(string email, string password);
        Task LogoutAsync();
        Task<AuthenticatedUser> GetMeAsync();
        Task<AuthenticatedUser> UpdateMeAsync(string? displayName, string? email);
        Task<bool> IsAuthenticatedAsync();
    }
}
