using Refit;

namespace ProxiBetApp.Services.Auth
{
    public sealed class AuthService : IAuthService
    {
        private readonly IAuthApi _authApi;
        private readonly ITokenStorage _tokenStorage;
        private readonly CurrentUserStore _currentUserStore;

        public AuthService(IAuthApi authApi, ITokenStorage tokenStorage, CurrentUserStore currentUserStore)
        {
            _authApi = authApi;
            _tokenStorage = tokenStorage;
            _currentUserStore = currentUserStore;
        }

        public async Task<AuthenticatedUser> RegisterAsync(string displayName, string email, string password)
        {
            try
            {
                return await _authApi.RegisterAsync(new RegisterRequest
                {
                    DisplayName = displayName,
                    Email = email,
                    Password = password,
                });
            }
            catch (ApiException ex)
            {
                throw await AuthApiException.FromRefitExceptionAsync(ex);
            }
            catch (Exception ex)
            {
                throw new NetworkException("Impossible de contacter le serveur. Vérifiez votre connexion.", ex);
            }
        }

        public async Task LoginAsync(string email, string password)
        {
            LoginResponse body;
            try
            {
                body = await _authApi.LoginAsync(new LoginRequest
                {
                    Email = email,
                    Password = password,
                });
            }
            catch (ApiException ex)
            {
                throw await AuthApiException.FromRefitExceptionAsync(ex);
            }
            catch (Exception ex)
            {
                throw new NetworkException("Impossible de contacter le serveur. Vérifiez votre connexion.", ex);
            }

            await ApplySessionAsync(body);
        }

        public async Task LogoutAsync()
        {
            await _tokenStorage.ClearAsync();
            _currentUserStore.CurrentUser = null;
        }

        public async Task<AuthenticatedUser> GetMeAsync()
        {
            AuthenticatedUser body;
            try
            {
                body = await _authApi.GetMeAsync();
            }
            catch (ApiException ex)
            {
                throw await AuthApiException.FromRefitExceptionAsync(ex);
            }
            catch (Exception ex)
            {
                throw new NetworkException("Impossible de contacter le serveur. Vérifiez votre connexion.", ex);
            }

            _currentUserStore.CurrentUser = body;
            return body;
        }

        public async Task<AuthenticatedUser> UpdateMeAsync(string? displayName, string? email)
        {
            AuthenticatedUser body;
            try
            {
                body = await _authApi.UpdateMeAsync(new UpdateMeRequest
                {
                    DisplayName = displayName,
                    Email = email,
                });
            }
            catch (ApiException ex)
            {
                throw await AuthApiException.FromRefitExceptionAsync(ex);
            }
            catch (Exception ex)
            {
                throw new NetworkException("Impossible de contacter le serveur. Vérifiez votre connexion.", ex);
            }

            _currentUserStore.CurrentUser = body;
            return body;
        }

        public async Task<bool> IsAuthenticatedAsync()
        {
            var session = await _tokenStorage.LoadAsync();
            return session is not null;
        }

        private async Task ApplySessionAsync(LoginResponse body)
        {
            // The backend's "expires_in" field is actually an absolute Unix timestamp, not a
            // duration (see proxiback/internal/auth/token.go). Decode the JWT's own "exp" claim
            // instead of computing DateTime.Now + expires_in, which would double-apply the offset.
            var expiresAtUtc = JwtUtils.TryGetExpiryUtc(body.AccessToken) ?? DateTimeOffset.FromUnixTimeSeconds(body.ExpiresIn);

            await _tokenStorage.SaveAsync(new StoredSession
            {
                AccessToken = body.AccessToken,
                RefreshToken = body.RefreshToken,
                ExpiresAtUtc = expiresAtUtc,
                User = body.User,
            });

            _currentUserStore.CurrentUser = body.User;
        }
    }
}
