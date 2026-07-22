import { Injectable, computed, signal } from '@angular/core';

import { AuthenticatedUser } from './auth.models';

@Injectable({ providedIn: 'root' })
export class CurrentUserStore {
  private readonly _user = signal<AuthenticatedUser | null>(null);

  readonly user = this._user.asReadonly();
  readonly isAuthenticated = computed(() => this._user() !== null);

  set(user: AuthenticatedUser | null): void {
    this._user.set(user);
  }

  clear(): void {
    this._user.set(null);
  }
}
