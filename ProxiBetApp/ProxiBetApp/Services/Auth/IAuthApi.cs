using ProxiBetApp.Models.Auth;
using Refit;

namespace ProxiBetApp.Services.Auth
{
    /// <summary>
    /// Pure network surface for the proxiback auth/me endpoints, consumed by Refit.
    /// Session/JWT/local-only concerns (logout, IsAuthenticatedAsync) live in IAuthService instead.
    /// </summary>
    public interface IAuthApi
    {
        [Post("/api/v1/auth/register")]
        Task<AuthenticatedUser> RegisterAsync([Body] RegisterRequest request);

        [Post("/api/v1/auth/login")]
        Task<LoginResponse> LoginAsync([Body] LoginRequest request);

        [Get("/api/v1/me")]
        Task<AuthenticatedUser> GetMeAsync();

        [Patch("/api/v1/me")]
        Task<AuthenticatedUser> UpdateMeAsync([Body] UpdateMeRequest request);
    }
}
