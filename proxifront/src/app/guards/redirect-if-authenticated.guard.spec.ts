import { TestBed } from '@angular/core/testing';
import { Router, UrlTree } from '@angular/router';

import { AuthService } from '../services/auth.service';
import { redirectIfAuthenticatedGuard } from './redirect-if-authenticated.guard';

describe('redirectIfAuthenticatedGuard', () => {
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
      redirectIfAuthenticatedGuard({} as never, {} as never),
    );
  }

  it('allows navigation when not authenticated', () => {
    authServiceMock.isAuthenticated = () => false;

    expect(runGuard()).toBe(true);
  });

  it('redirects to / when already authenticated', () => {
    authServiceMock.isAuthenticated = () => true;

    const result = runGuard();

    expect(result).toBeInstanceOf(UrlTree);
    expect(router.serializeUrl(result as UrlTree)).toBe('/');
  });
});
