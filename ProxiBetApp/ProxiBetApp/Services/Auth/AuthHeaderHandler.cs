using System.Net;
using System.Net.Http.Json;

namespace ProxiBetApp.Services.Auth
{
    /// <summary>
    /// Attaches the bearer token to outgoing requests and refreshes the session once on a 401,
    /// mirroring proxifront's auth.interceptor.ts. Talks to ITokenStorage and a raw, unauthenticated
    /// HttpClient directly (not IAuthService) to avoid a DI cycle, since AuthService's own HttpClient
    /// pipeline includes this handler.
    /// </summary>
    public sealed class AuthHeaderHandler : DelegatingHandler
    {
        private static readonly string[] AuthEndpoints =
        [
            "/api/v1/auth/login",
            "/api/v1/auth/register",
            "/api/v1/auth/refresh",
        ];

        private readonly ITokenStorage _tokenStorage;
        private readonly IHttpClientFactory _httpClientFactory;
        private Task<bool>? _refreshInFlight;
        private readonly SemaphoreSlim _refreshLock = new(1, 1);

        public AuthHeaderHandler(ITokenStorage tokenStorage, IHttpClientFactory httpClientFactory)
        {
            _tokenStorage = tokenStorage;
            _httpClientFactory = httpClientFactory;
        }

        protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            var path = request.RequestUri?.AbsolutePath ?? string.Empty;
            var isAuthEndpoint = AuthEndpoints.Any(path.EndsWith);

            if (!isAuthEndpoint)
            {
                var session = await _tokenStorage.LoadAsync();
                if (session is not null)
                    request.Headers.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", session.AccessToken);
            }

            var response = await base.SendAsync(request, cancellationToken);

            if (isAuthEndpoint || response.StatusCode != HttpStatusCode.Unauthorized)
                return response;

            var refreshed = await RefreshOnceAsync();
            if (!refreshed)
                return response;

            var retryRequest = await CloneRequestAsync(request);
            var newSession = await _tokenStorage.LoadAsync();
            if (newSession is not null)
                retryRequest.Headers.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", newSession.AccessToken);

            return await base.SendAsync(retryRequest, cancellationToken);
        }

        private async Task<bool> RefreshOnceAsync()
        {
            await _refreshLock.WaitAsync();
            try
            {
                _refreshInFlight ??= DoRefreshAsync();
                return await _refreshInFlight;
            }
            finally
            {
                _refreshInFlight = null;
                _refreshLock.Release();
            }
        }

        private async Task<bool> DoRefreshAsync()
        {
            var session = await _tokenStorage.LoadAsync();
            if (session is null)
                return false;

            var client = _httpClientFactory.CreateClient("proxibet-unauthenticated");
            var response = await client.PostAsJsonAsync("/api/v1/auth/refresh", new RefreshRequest { RefreshToken = session.RefreshToken });
            if (!response.IsSuccessStatusCode)
            {
                await _tokenStorage.ClearAsync();
                return false;
            }

            var body = await response.Content.ReadFromJsonAsync<LoginResponse>();
            if (body is null)
                return false;

            var expiresAtUtc = JwtUtils.TryGetExpiryUtc(body.AccessToken) ?? DateTimeOffset.FromUnixTimeSeconds(body.ExpiresIn);
            await _tokenStorage.SaveAsync(new StoredSession
            {
                AccessToken = body.AccessToken,
                RefreshToken = body.RefreshToken,
                ExpiresAtUtc = expiresAtUtc,
                User = body.User,
            });

            return true;
        }

        private static async Task<HttpRequestMessage> CloneRequestAsync(HttpRequestMessage request)
        {
            var clone = new HttpRequestMessage(request.Method, request.RequestUri);
            if (request.Content is not null)
            {
                var bytes = await request.Content.ReadAsByteArrayAsync();
                clone.Content = new ByteArrayContent(bytes);
                foreach (var header in request.Content.Headers)
                    clone.Content.Headers.Add(header.Key, header.Value);
            }
            foreach (var header in request.Headers)
                clone.Headers.TryAddWithoutValidation(header.Key, header.Value);

            return clone;
        }
    }
}
