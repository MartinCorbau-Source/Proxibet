import {
  HttpErrorResponse,
  HttpEvent,
  HttpHandlerFn,
  HttpInterceptorFn,
  HttpRequest,
} from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, catchError, finalize, share, switchMap, throwError } from 'rxjs';

import { AuthService } from '../services/auth.service';
import { RefreshResponse } from '../services/auth.models';

const AUTH_ENDPOINTS = ['/api/v1/auth/login', '/api/v1/auth/register', '/api/v1/auth/refresh'];

let refreshInFlight$: Observable<RefreshResponse> | null = null;

/** Exposed for tests only, to avoid state leaking between specs. */
export function resetAuthInterceptorState(): void {
  refreshInFlight$ = null;
}

function isAuthEndpoint(url: string): boolean {
  return AUTH_ENDPOINTS.some((path) => url.includes(path));
}

function withBearer<T>(req: HttpRequest<T>, token: string): HttpRequest<T> {
  return req.clone({ setHeaders: { Authorization: `Bearer ${token}` } });
}

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const router = inject(Router);

  const skipAuth = isAuthEndpoint(req.url);
  const accessToken = skipAuth ? null : authService.getAccessToken();

  const authReq = accessToken ? withBearer(req, accessToken) : req;

  return next(authReq).pipe(
    catchError((error: unknown) => {
      if (error instanceof HttpErrorResponse && error.status === 401 && !skipAuth) {
        return handle401(authReq, next, authService, router);
      }

      return throwError(() => error);
    }),
  );
};

function handle401(
  req: HttpRequest<unknown>,
  next: HttpHandlerFn,
  authService: AuthService,
  router: Router,
): Observable<HttpEvent<unknown>> {
  if (!refreshInFlight$) {
    refreshInFlight$ = authService.refresh().pipe(
      finalize(() => (refreshInFlight$ = null)),
      share(),
    );
  }

  return refreshInFlight$.pipe(
    switchMap((response) => next(withBearer(req, response.access_token))),
    catchError((refreshError: unknown) => {
      authService.logout();
      router.navigate(['/login']);
      return throwError(() => refreshError);
    }),
  );
}
