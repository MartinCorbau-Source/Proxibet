using System.Net.Http.Json;

namespace ProxiBetApp.Services.Auth
{
    public sealed class AuthService : IAuthService
    {
        private readonly HttpClient _httpClient;
        private readonly ITokenStorage _tokenStorage;
        private readonly CurrentUserStore _currentUserStore;

        public AuthService(HttpClient httpClient, ITokenStorage tokenStorage, CurrentUserStore currentUserStore)
        {
            _httpClient = httpClient;
            _tokenStorage = tokenStorage;
            _currentUserStore = currentUserStore;
        }

        public async Task<AuthenticatedUser> RegisterAsync(string displayName, string email, string password)
        {
            var response = await _httpClient.PostAsJsonAsync("/api/v1/auth/register", new RegisterRequest
            {
                DisplayName = displayName,
                Email = email,
                Password = password,
            });

            await EnsureSuccessAsync(response);
            var body = await response.Content.ReadFromJsonAsync<AuthenticatedUser>();
            return body ?? throw new AuthApiException("INTERNAL_ERROR", "Une erreur inattendue est survenue. Veuillez réessayer.");
        }

        public async Task LoginAsync(string email, string password)
        {
            var response = await _httpClient.PostAsJsonAsync("/api/v1/auth/login", new LoginRequest
            {
                Email = email,
                Password = password,
            });

            await EnsureSuccessAsync(response);
            var body = await response.Content.ReadFromJsonAsync<LoginResponse>();
            if (body is null)
                throw new AuthApiException("INTERNAL_ERROR", "Une erreur inattendue est survenue. Veuillez réessayer.");

            await ApplySessionAsync(body);
        }

        public async Task LogoutAsync()
        {
            await _tokenStorage.ClearAsync();
            _currentUserStore.CurrentUser = null;
        }

        public async Task<AuthenticatedUser> GetMeAsync()
        {
            var response = await _httpClient.GetAsync("/api/v1/me");
            await EnsureSuccessAsync(response);
            var body = await response.Content.ReadFromJsonAsync<AuthenticatedUser>();
            if (body is null)
                throw new AuthApiException("INTERNAL_ERROR", "Une erreur inattendue est survenue. Veuillez réessayer.");

            _currentUserStore.CurrentUser = body;
            return body;
        }

        public async Task<AuthenticatedUser> UpdateMeAsync(string? displayName, string? email)
        {
            var response = await _httpClient.PatchAsJsonAsync("/api/v1/me", new UpdateMeRequest
            {
                DisplayName = displayName,
                Email = email,
            });

            await EnsureSuccessAsync(response);
            var body = await response.Content.ReadFromJsonAsync<AuthenticatedUser>();
            if (body is null)
                throw new AuthApiException("INTERNAL_ERROR", "Une erreur inattendue est survenue. Veuillez réessayer.");

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

        private static async Task EnsureSuccessAsync(HttpResponseMessage response)
        {
            if (response.IsSuccessStatusCode)
                return;

            ApiErrorResponse? error;
            try
            {
                error = await response.Content.ReadFromJsonAsync<ApiErrorResponse>();
            }
            catch (Exception)
            {
                error = null;
            }

            if (error is not null && !string.IsNullOrEmpty(error.Message))
                throw new AuthApiException(error.Error, error.Message);

            throw new AuthApiException("INTERNAL_ERROR", "Une erreur inattendue est survenue. Veuillez réessayer.");
        }
    }
}
