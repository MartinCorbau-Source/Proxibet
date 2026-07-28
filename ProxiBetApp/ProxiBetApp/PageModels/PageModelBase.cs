using CommunityToolkit.Mvvm.ComponentModel;

namespace ProxiBetApp.PageModels
{
    /// <summary>
    /// Base commune aux PageModels dont une commande appelle un service consommant une API
    /// distante. Centralise IsBusy/ErrorMessage (dupliqués à l'identique dans les 4 PageModels
    /// avant cette classe) et ExecuteWithErrorHandlingAsync, qui applique le pattern attendu :
    /// catch (AuthApiException) pour une erreur "métier" renvoyée par le serveur, catch
    /// (NetworkException) pour une panne de transport (DNS, connexion refusée, timeout).
    ///
    /// Dérive de ObservableValidator (pas ObservableObject) car LoginPageModel/RegisterPageModel
    /// ont besoin de [NotifyDataErrorInfo] ; ObservableValidator sans attribut de validation posé
    /// se comporte comme un ObservableObject ordinaire, donc HomePageModel/MePageModel n'ont
    /// aucun effet de bord à en hériter aussi.
    ///
    /// Limite assumée : ce helper ne peut pas empêcher un futur PageModel d'écrire son propre
    /// try/catch à côté au lieu de l'appeler — C# n'a pas de mécanisme pour rendre l'appel
    /// obligatoire. Le vrai filet de sécurité anti-crash reste la couche Services/ : chaque
    /// méthode de service qui appelle une API doit garantir qu'aucune exception autre que
    /// AuthApiException/NetworkException (ou équivalent pour un futur domaine) n'en sort (voir
    /// AuthService). Ce que ce helper apporte réellement : une ligne à écrire au lieu de dix,
    /// et un point unique à vérifier en revue de code. Voir ProxiBetApp/CLAUDE.md pour le détail.
    /// </summary>
    public abstract partial class PageModelBase : ObservableValidator
    {
        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string? _errorMessage;

        protected async Task ExecuteWithErrorHandlingAsync(
            Func<Task> operation,
            Func<AuthApiException, Task>? onAuthError = null)
        {
            IsBusy = true;
            ErrorMessage = null;

            try
            {
                await operation();
            }
            catch (AuthApiException ex)
            {
                if (onAuthError is not null)
                    await onAuthError(ex);
                else
                    ErrorMessage = ex.Message;
            }
            catch (NetworkException ex)
            {
                ErrorMessage = ex.Message;
            }
            finally
            {
                IsBusy = false;
            }
        }
    }
}
