namespace ProxiBetApp.Services.Auth
{
    public sealed class AuthApiException : Exception
    {
        public string Code { get; }

        public AuthApiException(string code, string message) : base(message)
        {
            Code = code;
        }
    }
}
