using ProxiBetApp.Models.Auth;
using Refit;

namespace ProxiBetApp.Services.Auth
{
    public sealed class AuthApiException : Exception
    {
        public string Code { get; }

        public AuthApiException(string code, string message) : base(message)
        {
            Code = code;
        }

        public static async Task<AuthApiException> FromRefitExceptionAsync(ApiException refitException)
        {
            ApiErrorResponse? error;
            try
            {
                error = await refitException.GetContentAsAsync<ApiErrorResponse>();
            }
            catch (Exception)
            {
                error = null;
            }

            if (error is not null && !string.IsNullOrEmpty(error.Message))
                return new AuthApiException(error.Error, error.Message);

            return new AuthApiException("INTERNAL_ERROR", "Une erreur inattendue est survenue. Veuillez réessayer.");
        }
    }
}
