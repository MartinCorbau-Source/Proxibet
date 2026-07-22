import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { Subject, throwError } from 'rxjs';

import { environment } from '../../environments/environment';
import { AuthService } from '../services/auth.service';
import { authInterceptor, resetAuthInterceptorState } from './auth.interceptor';

describe('authInterceptor', () => {
  let http: HttpClient;
  let httpMock: HttpTestingController;
  let authServiceMock: {
    getAccessToken: ReturnType<typeof vi.fn>;
    refresh: ReturnType<typeof vi.fn>;
    logout: ReturnType<typeof vi.fn>;
  };
  let router: Router;

  beforeEach(() => {
    resetAuthInterceptorState();

    authServiceMock = {
      getAccessToken: vi.fn(() => 'access-token'),
      refresh: vi.fn(),
      logout: vi.fn(),
    };

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: AuthService, useValue: authServiceMock },
      ],
    });

    http = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
    vi.spyOn(router, 'navigate').mockResolvedValue(true);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('attaches the bearer token to protected requests', () => {
    http.get('/api/v1/me').subscribe();

    const req = httpMock.expectOne('/api/v1/me');
    expect(req.request.headers.get('Authorization')).toBe('Bearer access-token');
    req.flush({});
  });

  it('does not attach a token to the login endpoint', () => {
    http.post(`${environment.apiUrl}/api/v1/auth/login`, {}).subscribe();

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/auth/login`);
    expect(req.request.headers.has('Authorization')).toBe(false);
    req.flush({});
  });

  it('refreshes once and retries both requests on concurrent 401s', () => {
    const refresh$ = new Subject<{
      access_token: string;
      refresh_token: string;
      expires_in: number;
      user: unknown;
    }>();
    authServiceMock.refresh.mockReturnValue(refresh$);

    http.get('/api/v1/me').subscribe();
    http.get('/api/v1/other').subscribe();

    const req1 = httpMock.expectOne('/api/v1/me');
    const req2 = httpMock.expectOne('/api/v1/other');
    req1.flush({}, { status: 401, statusText: 'Unauthorized' });
    req2.flush({}, { status: 401, statusText: 'Unauthorized' });

    expect(authServiceMock.refresh).toHaveBeenCalledTimes(1);

    refresh$.next({ access_token: 'new-token', refresh_token: 'r', expires_in: 1, user: {} });
    refresh$.complete();

    const retry1 = httpMock.expectOne('/api/v1/me');
    const retry2 = httpMock.expectOne('/api/v1/other');
    expect(retry1.request.headers.get('Authorization')).toBe('Bearer new-token');
    expect(retry2.request.headers.get('Authorization')).toBe('Bearer new-token');
    retry1.flush({});
    retry2.flush({});
  });

  it('logs out and redirects to /login when refresh itself fails', () => {
    authServiceMock.refresh.mockReturnValue(throwError(() => new Error('refresh failed')));

    let receivedError: unknown;
    http.get('/api/v1/me').subscribe({ error: (err) => (receivedError = err) });

    const req = httpMock.expectOne('/api/v1/me');
    req.flush({}, { status: 401, statusText: 'Unauthorized' });

    expect(authServiceMock.logout).toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
    expect(receivedError).toBeTruthy();
  });
});
