import { TestBed } from '@angular/core/testing';

import { AuthenticatedUser } from './auth.models';
import { StoredAuthTokens, TokenStorageService } from './token-storage.service';

describe('TokenStorageService', () => {
  let service: TokenStorageService;

  const user: AuthenticatedUser = { id: '1', display_name: 'Jean', email: 'jean@test.fr' };
  const tokens: StoredAuthTokens = {
    accessToken: 'access',
    refreshToken: 'refresh',
    expiresAt: 1234567890,
    user,
  };

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({});
    service = TestBed.inject(TokenStorageService);
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('returns null when nothing is stored', () => {
    expect(service.load()).toBeNull();
  });

  it('round-trips saved tokens', () => {
    service.save(tokens);
    expect(service.load()).toEqual(tokens);
  });

  it('returns null when stored content is corrupted', () => {
    localStorage.setItem('proxibet_auth', 'not-json');
    expect(service.load()).toBeNull();
  });

  it('removes stored tokens on clear', () => {
    service.save(tokens);
    service.clear();
    expect(service.load()).toBeNull();
  });
});
