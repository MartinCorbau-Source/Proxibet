import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, tap, throwError } from 'rxjs';

import { environment } from '../../environments/environment';
import {
  AuthenticatedUser,
  LoginRequest,
  LoginResponse,
  RefreshRequest,
  RefreshResponse,
  RegisterRequest,
  RegisterResponse,
} from './auth.models';
import { CurrentUserStore } from './current-user.store';
import { TokenStorageService } from './token-storage.service';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly tokenStorage = inject(TokenStorageService);
  private readonly currentUserStore = inject(CurrentUserStore);

  readonly currentUser = this.currentUserStore.user;
  readonly isAuthenticated = this.currentUserStore.isAuthenticated;

  constructor() {
    this.currentUserStore.set(this.tokenStorage.load()?.user ?? null);
  }

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
      .pipe(tap((response) => this.applySession(response)));
  }

  refresh(): Observable<RefreshResponse> {
    const stored = this.tokenStorage.load();
    if (!stored) {
      return throwError(() => new Error('No refresh token available'));
    }

    const request: RefreshRequest = { refresh_token: stored.refreshToken };

    return this.http
      .post<RefreshResponse>(`${environment.apiUrl}/api/v1/auth/refresh`, request)
      .pipe(tap((response) => this.applySession(response)));
  }

  logout(): void {
    this.tokenStorage.clear();
    this.currentUserStore.clear();
    // Pas d'endpoint de logout côté backend pour l'instant : le refresh token
    // reste valide côté serveur jusqu'à expiration naturelle ou rotation.
  }

  getAccessToken(): string | null {
    return this.tokenStorage.load()?.accessToken ?? null;
  }

  getMe(): Observable<AuthenticatedUser> {
    return this.http
      .get<AuthenticatedUser>(`${environment.apiUrl}/api/v1/me`)
      .pipe(tap((user) => this.currentUserStore.set(user)));
  }

  private applySession(response: LoginResponse | RefreshResponse): void {
    this.tokenStorage.save({
      accessToken: response.access_token,
      refreshToken: response.refresh_token,
      expiresAt: Date.now() + response.expires_in * 1000,
      user: response.user,
    });
    this.currentUserStore.set(response.user);
  }
}
