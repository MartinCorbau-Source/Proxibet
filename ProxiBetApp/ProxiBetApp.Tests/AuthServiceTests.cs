using System.Net;
using Moq;
using ProxiBetApp.Models.Auth;
using ProxiBetApp.Services.Auth;
using Xunit;

namespace ProxiBetApp.Tests
{
    public class AuthServiceTests
    {
        [Fact]
        public async Task LoginAsync_Success_SavesSessionAndUpdatesCurrentUser()
        {
            var loginResponseJson = """
                {
                    "access_token": "header.payload.signature",
                    "refresh_token": "refresh-token",
                    "expires_in": 1234567890,
                    "user": { "id": "u1", "display_name": "Alice", "email": "alice@example.com" }
                }
                """;
            var handler = FakeHttpMessageHandler.ReturningJson(HttpStatusCode.OK, loginResponseJson);
            var httpClient = new HttpClient(handler) { BaseAddress = new Uri("http://localhost:8080") };

            var tokenStorageMock = new Mock<ITokenStorage>();
            StoredSession? savedSession = null;
            tokenStorageMock
                .Setup(s => s.SaveAsync(It.IsAny<StoredSession>()))
                .Callback<StoredSession>(session => savedSession = session)
                .Returns(Task.CompletedTask);

            var currentUserStore = new CurrentUserStore();
            var sut = new AuthService(httpClient, tokenStorageMock.Object, currentUserStore);

            await sut.LoginAsync("alice@example.com", "password");

            Assert.NotNull(savedSession);
            Assert.Equal("refresh-token", savedSession!.RefreshToken);
            Assert.Equal("alice@example.com", currentUserStore.CurrentUser?.Email);
            Assert.True(currentUserStore.IsAuthenticated);
        }

        [Fact]
        public async Task LoginAsync_ErrorWithBody_ThrowsAuthApiExceptionWithBackendMessage()
        {
            var errorJson = """{ "error": "INVALID_CREDENTIALS", "message": "Email ou mot de passe invalide." }""";
            var handler = FakeHttpMessageHandler.ReturningJson(HttpStatusCode.Unauthorized, errorJson);
            var httpClient = new HttpClient(handler) { BaseAddress = new Uri("http://localhost:8080") };
            var sut = new AuthService(httpClient, Mock.Of<ITokenStorage>(), new CurrentUserStore());

            var ex = await Assert.ThrowsAsync<AuthApiException>(() => sut.LoginAsync("alice@example.com", "wrong"));

            Assert.Equal("INVALID_CREDENTIALS", ex.Code);
            Assert.Equal("Email ou mot de passe invalide.", ex.Message);
        }

        [Fact]
        public async Task LoginAsync_ErrorWithoutJsonBody_ThrowsGenericAuthApiException()
        {
            var handler = new FakeHttpMessageHandler(_ => new HttpResponseMessage(HttpStatusCode.InternalServerError)
            {
                Content = new StringContent("not json"),
            });
            var httpClient = new HttpClient(handler) { BaseAddress = new Uri("http://localhost:8080") };
            var sut = new AuthService(httpClient, Mock.Of<ITokenStorage>(), new CurrentUserStore());

            var ex = await Assert.ThrowsAsync<AuthApiException>(() => sut.LoginAsync("alice@example.com", "password"));

            Assert.Equal("INTERNAL_ERROR", ex.Code);
        }
    }
}
