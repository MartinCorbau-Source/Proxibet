namespace ProxiBetApp.Services
{
    /// <summary>
    /// Transport failure (DNS resolution, connection refused, timeout...) wrapping any API call.
    /// Lives at the root of Services/ rather than Services/Auth/ because it isn't auth-specific:
    /// any future API domain would hit the same transport failures and should reuse this type.
    /// </summary>
    public sealed class NetworkException : Exception
    {
        public NetworkException(string message, Exception innerException)
            : base(message, innerException)
        {
        }
    }
}
