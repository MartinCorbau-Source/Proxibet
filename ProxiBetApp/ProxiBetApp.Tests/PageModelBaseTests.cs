using ProxiBetApp.PageModels;
using ProxiBetApp.Services;
using ProxiBetApp.Services.Auth;
using Xunit;

namespace ProxiBetApp.Tests
{
    public sealed partial class TestPageModel : PageModelBase
    {
        public Task RunAsync(Func<Task> operation, Func<AuthApiException, Task>? onAuthError = null) =>
            ExecuteWithErrorHandlingAsync(operation, onAuthError);
    }

    public class PageModelBaseTests
    {
        [Fact]
        public async Task ExecuteWithErrorHandlingAsync_NetworkException_SetsErrorMessageWithoutThrowing()
        {
            var sut = new TestPageModel();

            await sut.RunAsync(() => throw new NetworkException("Impossible de contacter le serveur.", new Exception("dns failure")));

            Assert.Equal("Impossible de contacter le serveur.", sut.ErrorMessage);
            Assert.False(sut.IsBusy);
        }

        [Fact]
        public async Task ExecuteWithErrorHandlingAsync_AuthApiException_WithoutHandler_SetsErrorMessage()
        {
            var sut = new TestPageModel();

            await sut.RunAsync(() => throw new AuthApiException("INVALID_CREDENTIALS", "Email ou mot de passe invalide."));

            Assert.Equal("Email ou mot de passe invalide.", sut.ErrorMessage);
            Assert.False(sut.IsBusy);
        }

        [Fact]
        public async Task ExecuteWithErrorHandlingAsync_AuthApiException_WithHandler_CallsHandlerInsteadOfSettingErrorMessage()
        {
            var sut = new TestPageModel();
            AuthApiException? received = null;

            await sut.RunAsync(
                () => throw new AuthApiException("SESSION_EXPIRED", "Session expirée."),
                onAuthError: ex =>
                {
                    received = ex;
                    return Task.CompletedTask;
                });

            Assert.NotNull(received);
            Assert.Equal("SESSION_EXPIRED", received!.Code);
            Assert.Null(sut.ErrorMessage);
            Assert.False(sut.IsBusy);
        }

        [Fact]
        public async Task ExecuteWithErrorHandlingAsync_Success_DoesNotSetErrorMessage()
        {
            var sut = new TestPageModel();

            await sut.RunAsync(() => Task.CompletedTask);

            Assert.Null(sut.ErrorMessage);
            Assert.False(sut.IsBusy);
        }
    }
}
