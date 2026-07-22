import { TestBed } from '@angular/core/testing';
import { Router, UrlTree } from '@angular/router';

import { AuthService } from '../services/auth.service';
import { authGuard } from './auth.guard';

describe('authGuard', () => {
  let authServiceMock: { isAuthenticated: () => boolean };
  let router: Router;

  beforeEach(() => {
    authServiceMock = { isAuthenticated: () => false };

    TestBed.configureTestingModule({
      providers: [{ provide: AuthService, useValue: authServiceMock }],
    });

    router = TestBed.inject(Router);
  });

  function runGuard() {
    return TestBed.runInInjectionContext(() =>
      authGuard({} as never, {} as never),
    );
  }

  it('allows navigation when authenticated', () => {
    authServiceMock.isAuthenticated = () => true;

    expect(runGuard()).toBe(true);
  });

  it('redirects to /login when not authenticated', () => {
    authServiceMock.isAuthenticated = () => false;

    const result = runGuard();

    expect(result).toBeInstanceOf(UrlTree);
    expect(router.serializeUrl(result as UrlTree)).toBe('/login');
  });
});
