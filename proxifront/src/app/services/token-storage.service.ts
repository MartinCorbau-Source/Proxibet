import { Injectable } from '@angular/core';

import { AuthenticatedUser } from './auth.models';

export interface StoredAuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  user: AuthenticatedUser;
}

const STORAGE_KEY = 'proxibet_auth';

@Injectable({ providedIn: 'root' })
export class TokenStorageService {
  load(): StoredAuthTokens | null {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }

    try {
      return JSON.parse(raw) as StoredAuthTokens;
    } catch {
      return null;
    }
  }

  save(tokens: StoredAuthTokens): void {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(tokens));
  }

  clear(): void {
    localStorage.removeItem(STORAGE_KEY);
  }
}
