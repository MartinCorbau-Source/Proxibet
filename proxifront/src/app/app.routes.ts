import { Routes } from '@angular/router';

import { authGuard } from './guards/auth.guard';
import { redirectIfAuthenticatedGuard } from './guards/redirect-if-authenticated.guard';

export const routes: Routes = [
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/home/home.component').then((m) => m.HomeComponent),
  },
  {
    path: 'components',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/components/components.component').then((m) => m.ComponentsComponent),
  },
  {
    path: 'me',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/me/me.component').then((m) => m.MeComponent),
  },
  {
    path: 'register',
    canActivate: [redirectIfAuthenticatedGuard],
    loadComponent: () =>
      import('./pages/register/register.component').then((m) => m.RegisterComponent),
  },
  {
    path: 'login',
    canActivate: [redirectIfAuthenticatedGuard],
    loadComponent: () =>
      import('./pages/login/login.component').then((m) => m.LoginComponent),
  },
];
