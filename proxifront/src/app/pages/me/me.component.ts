import { Component, inject, OnInit, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { Router } from '@angular/router';

import { AuthenticatedUser } from '../../services/auth.models';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-me',
  imports: [MatButtonModule, MatCardModule],
  templateUrl: './me.component.html',
  styleUrl: './me.component.scss',
})
export class MeComponent implements OnInit {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly user = signal<AuthenticatedUser | null>(null);
  protected readonly loading = signal(true);
  protected readonly errorMessage = signal<string | null>(null);

  ngOnInit(): void {
    this.authService.getMe().subscribe({
      next: (user) => {
        this.loading.set(false);
        this.user.set(user);
      },
      error: () => {
        this.loading.set(false);
        this.errorMessage.set('Impossible de récupérer votre profil.');
      },
    });
  }

  protected logout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
