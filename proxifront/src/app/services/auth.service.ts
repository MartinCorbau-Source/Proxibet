import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { environment } from '../../environments/environment';
import {
  AuthenticatedUser,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
} from './auth.models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);

  private readonly _currentUser = signal<AuthenticatedUser | null>(null);
  readonly currentUser = this._currentUser.asReadonly();
  readonly isAuthenticated = computed(() => this._currentUser() !== null);

  register(displayName: string, email: string, password: string): Observable<RegisterResponse> {
    const request: RegisterRequest = {
      display_name: displayName,
      email,
      password,
    };

    return this.http.post<RegisterResponse>(`${environment.apiUrl}/api/v1/auth/register`, request);
  }

  login(email: string, password: string): Observable<LoginResponse> {
    const request: LoginRequest = { email, password };

    return this.http
      .post<LoginResponse>(`${environment.apiUrl}/api/v1/auth/login`, request)
      .pipe(tap((response) => this._currentUser.set(response.user)));
  }

  logout(): void {
    this._currentUser.set(null);
    // TODO: appeler un futur POST /api/v1/auth/logout pour invalider le cookie côté serveur
  }
}
