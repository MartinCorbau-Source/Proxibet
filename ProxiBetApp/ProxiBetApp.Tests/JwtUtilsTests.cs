using System.Text;
using System.Text.Json;
using ProxiBetApp.Services.Auth;
using Xunit;

namespace ProxiBetApp.Tests
{
    public class JwtUtilsTests
    {
        [Fact]
        public void TryGetExpiryUtc_ValidJwtWithExpClaim_ReturnsExpiry()
        {
            var exp = DateTimeOffset.UtcNow.AddHours(1).ToUnixTimeSeconds();
            var jwt = BuildJwt(new { exp });

            var result = JwtUtils.TryGetExpiryUtc(jwt);

            Assert.NotNull(result);
            Assert.Equal(exp, result!.Value.ToUnixTimeSeconds());
        }

        [Fact]
        public void TryGetExpiryUtc_MalformedSegmentCount_ReturnsNull()
        {
            var result = JwtUtils.TryGetExpiryUtc("only.two");

            Assert.Null(result);
        }

        [Fact]
        public void TryGetExpiryUtc_PayloadNotJson_ReturnsNull()
        {
            var header = Base64UrlEncode("header");
            var payload = Base64UrlEncode("not-json");
            var signature = Base64UrlEncode("sig");

            var result = JwtUtils.TryGetExpiryUtc($"{header}.{payload}.{signature}");

            Assert.Null(result);
        }

        [Fact]
        public void TryGetExpiryUtc_ExpClaimMissing_ReturnsNull()
        {
            var jwt = BuildJwt(new { sub = "user-1" });

            var result = JwtUtils.TryGetExpiryUtc(jwt);

            Assert.Null(result);
        }

        private static string BuildJwt(object payload)
        {
            var header = Base64UrlEncode("{\"alg\":\"none\"}");
            var payloadJson = JsonSerializer.Serialize(payload);
            var payloadSegment = Base64UrlEncode(payloadJson);
            var signature = Base64UrlEncode("sig");
            return $"{header}.{payloadSegment}.{signature}";
        }

        private static string Base64UrlEncode(string value)
        {
            return Convert.ToBase64String(Encoding.UTF8.GetBytes(value))
                .Replace('+', '-')
                .Replace('/', '_')
                .TrimEnd('=');
        }
    }
}
