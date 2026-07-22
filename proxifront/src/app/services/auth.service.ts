import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { RegisterRequest, RegisterResponse } from './auth.models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);

  register(displayName: string, email: string, password: string): Observable<RegisterResponse> {
    const request: RegisterRequest = {
      display_name: displayName,
      email,
      password,
    };

    return this.http.post<RegisterResponse>(`${environment.apiUrl}/api/v1/auth/register`, request);
  }
}
