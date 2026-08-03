using System.Net;
using Moq;
using ProxiBetApp.Models.Auth;
using ProxiBetApp.Services.Auth;
using Xunit;

namespace ProxiBetApp.Tests
{
    public class AuthHeaderHandlerTests
    {
        [Fact]
        public async Task SendAsync_NoStoredSession_DoesNotAttachAuthorizationHeader()
        {
            var tokenStorageMock = new Mock<ITokenStorage>();
            tokenStorageMock.Setup(s => s.LoadAsync()).ReturnsAsync((StoredSession?)null);

            var innerHandler = FakeHttpMessageHandler.ReturningJson(HttpStatusCode.OK, "{}");
            var sut = CreateHandler(tokenStorageMock.Object, Mock.Of<IHttpClientFactory>(), innerHandler);

            var client = new HttpClient(sut) { BaseAddress = new Uri("http://localhost:8080") };
            await client.GetAsync("/api/v1/me");

            Assert.Null(innerHandler.Requests[0].Headers.Authorization);
        }

        [Fact]
        public async Task SendAsync_WithStoredSession_AttachesBearerHeader()
        {
            var session = new StoredSession { AccessToken = "access-token" };
            var tokenStorageMock = new Mock<ITokenStorage>();
            tokenStorageMock.Setup(s => s.LoadAsync()).ReturnsAsync(session);

            var innerHandler = FakeHttpMessageHandler.ReturningJson(HttpStatusCode.OK, "{}");
            var sut = CreateHandler(tokenStorageMock.Object, Mock.Of<IHttpClientFactory>(), innerHandler);

            var client = new HttpClient(sut) { BaseAddress = new Uri("http://localhost:8080") };
            await client.GetAsync("/api/v1/me");

            Assert.Equal("Bearer", innerHandler.Requests[0].Headers.Authorization?.Scheme);
            Assert.Equal("access-token", innerHandler.Requests[0].Headers.Authorization?.Parameter);
        }

        [Theory]
        [InlineData("/api/v1/auth/login")]
        [InlineData("/api/v1/auth/register")]
        [InlineData("/api/v1/auth/refresh")]
        public async Task SendAsync_AuthEndpoint_NeverAttachesAuthorizationHeader(string path)
        {
            var session = new StoredSession { AccessToken = "access-token" };
            var tokenStorageMock = new Mock<ITokenStorage>();
            tokenStorageMock.Setup(s => s.LoadAsync()).ReturnsAsync(session);

            var innerHandler = FakeHttpMessageHandler.ReturningJson(HttpStatusCode.OK, "{}");
            var sut = CreateHandler(tokenStorageMock.Object, Mock.Of<IHttpClientFactory>(), innerHandler);

            var client = new HttpClient(sut) { BaseAddress = new Uri("http://localhost:8080") };
            await client.GetAsync(path);

            Assert.Null(innerHandler.Requests[0].Headers.Authorization);
        }

        [Fact]
        public async Task SendAsync_UnauthorizedResponse_RefreshesSessionAndRetriesRequest()
        {
            var expiredSession = new StoredSession { AccessToken = "expired-token", RefreshToken = "refresh-token" };
            var tokenStorageMock = new Mock<ITokenStorage>();
            var loadCallCount = 0;
            tokenStorageMock
                .Setup(s => s.LoadAsync())
                .ReturnsAsync(() =>
                {
                    loadCallCount++;
                    return loadCallCount == 1
                        ? expiredSession
                        : new StoredSession { AccessToken = "new-token", RefreshToken = "refresh-token" };
                });

            var refreshResponseJson = """
                {
                    "access_token": "new-token",
                    "refresh_token": "refresh-token",
                    "expires_in": 1234567890,
                    "user": { "id": "u1", "display_name": "Alice", "email": "alice@example.com" }
                }
                """;
            var refreshHandler = FakeHttpMessageHandler.ReturningJson(HttpStatusCode.OK, refreshResponseJson);
            var refreshClient = new HttpClient(refreshHandler) { BaseAddress = new Uri("http://localhost:8080") };

            var httpClientFactoryMock = new Mock<IHttpClientFactory>();
            httpClientFactoryMock.Setup(f => f.CreateClient("proxibet-unauthenticated")).Returns(refreshClient);

            var requestCount = 0;
            var innerHandler = new FakeHttpMessageHandler(_ =>
            {
                requestCount++;
                return requestCount == 1
                    ? new HttpResponseMessage(HttpStatusCode.Unauthorized)
                    : new HttpResponseMessage(HttpStatusCode.OK) { Content = new StringContent("{}") };
            });

            var sut = CreateHandler(tokenStorageMock.Object, httpClientFactoryMock.Object, innerHandler);
            var client = new HttpClient(sut) { BaseAddress = new Uri("http://localhost:8080") };

            var response = await client.GetAsync("/api/v1/me");

            Assert.Equal(HttpStatusCode.OK, response.StatusCode);
            Assert.Equal(2, innerHandler.Requests.Count);
            Assert.Equal("new-token", innerHandler.Requests[1].Headers.Authorization?.Parameter);
            tokenStorageMock.Verify(s => s.SaveAsync(It.Is<StoredSession>(session => session.AccessToken == "new-token")), Times.Once);
        }

        private static AuthHeaderHandler CreateHandler(ITokenStorage tokenStorage, IHttpClientFactory httpClientFactory, HttpMessageHandler innerHandler)
        {
            var handler = new AuthHeaderHandler(tokenStorage, httpClientFactory)
            {
                InnerHandler = innerHandler,
            };
            return handler;
        }
    }
}
