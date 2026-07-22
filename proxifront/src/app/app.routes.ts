import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'components',
    loadComponent: () =>
      import('./pages/components/components.component').then((m) => m.ComponentsComponent),
  },
];
