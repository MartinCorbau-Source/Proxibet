namespace ProxiBetApp.Services.Auth
{
    public static class ApiConfig
    {
        // "localhost" reaches the host machine directly on the iOS simulator,
        // MacCatalyst, and Windows. On Android it also works, but only because of
        // "adb reverse tcp:8080 tcp:8080" (see proxiback README / dev notes) - that
        // command must be re-run after every USB reconnect, for both the emulator
        // (whose 10.0.2.2 alias this project no longer uses) and physical devices.
        //
        // No production URL exists yet, so Release builds throw instead of silently
        // shipping with a dev-only address - set it here once one is available.
        public static string BaseUrl =>
#if DEBUG
            "http://localhost:8080";
#else
            throw new InvalidOperationException(
                "ApiConfig.BaseUrl has no production value configured yet. Set the production API URL before shipping a Release build.");
#endif
    }
}
