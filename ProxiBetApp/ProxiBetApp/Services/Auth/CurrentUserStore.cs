using CommunityToolkit.Mvvm.ComponentModel;

namespace ProxiBetApp.Services.Auth
{
    public partial class CurrentUserStore : ObservableObject
    {
        [ObservableProperty]
        [NotifyPropertyChangedFor(nameof(IsAuthenticated))]
        private AuthenticatedUser? _currentUser;

        public bool IsAuthenticated => CurrentUser is not null;
    }
}
