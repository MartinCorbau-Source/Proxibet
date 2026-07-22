import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../environments/environment';
import { AuthenticatedUser, LoginResponse } from './auth.models';
import { AuthService } from './auth.service';
import { TokenStorageService } from './token-storage.service';

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;
  let tokenStorage: TokenStorageService;

  const user: AuthenticatedUser = { id: '1', display_name: 'Jean', email: 'jean@test.fr' };
  const loginResponse: LoginResponse = {
    access_token: 'access-1',
    refresh_token: 'refresh-1',
    expires_in: 1234567890,
    user,
  };

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    tokenStorage = TestBed.inject(TokenStorageService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  it('starts unauthenticated when nothing is stored', () => {
    service = TestBed.inject(AuthService);
    expect(service.isAuthenticated()).toBe(false);
    expect(service.currentUser()).toBeNull();
  });

  it('rehydrates currentUser from storage on construction', () => {
    tokenStorage.save({
      accessToken: 'access',
      refreshToken: 'refresh',
      expiresAt: 1234567890,
      user,
    });

    service = TestBed.inject(AuthService);

    expect(service.isAuthenticated()).toBe(true);
    expect(service.currentUser()).toEqual(user);
  });

  it('login persists tokens and updates currentUser', () => {
    service = TestBed.inject(AuthService);

    service.login('jean@test.fr', 'password123').subscribe();

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/auth/login`);
    req.flush(loginResponse);

    expect(service.currentUser()).toEqual(user);
    expect(tokenStorage.load()?.accessToken).toBe('access-1');
    expect(tokenStorage.load()?.refreshToken).toBe('refresh-1');
  });

  it('refresh applies the rotated tokens the same way login does', () => {
    tokenStorage.save({
      accessToken: 'stale',
      refreshToken: 'refresh-old',
      expiresAt: 1,
      user,
    });
    service = TestBed.inject(AuthService);

    const rotated: LoginResponse = {
      access_token: 'access-2',
      refresh_token: 'refresh-2',
      expires_in: 999,
      user,
    };

    service.refresh().subscribe();

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/auth/refresh`);
    expect(req.request.body).toEqual({ refresh_token: 'refresh-old' });
    req.flush(rotated);

    expect(tokenStorage.load()?.accessToken).toBe('access-2');
    expect(tokenStorage.load()?.refreshToken).toBe('refresh-2');
  });

  it('logout clears storage and currentUser', () => {
    tokenStorage.save({
      accessToken: 'access',
      refreshToken: 'refresh',
      expiresAt: 1234567890,
      user,
    });
    service = TestBed.inject(AuthService);

    service.logout();

    expect(service.currentUser()).toBeNull();
    expect(tokenStorage.load()).toBeNull();
  });

  it('getAccessToken reads the stored access token', () => {
    tokenStorage.save({
      accessToken: 'access',
      refreshToken: 'refresh',
      expiresAt: 1234567890,
      user,
    });
    service = TestBed.inject(AuthService);

    expect(service.getAccessToken()).toBe('access');
  });
});
