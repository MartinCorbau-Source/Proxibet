import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { environment } from '../../../environments/environment';
import { AuthenticatedUser } from '../../services/auth.models';
import { MeComponent } from './me.component';

describe('MeComponent', () => {
  let component: MeComponent;
  let fixture: ComponentFixture<MeComponent>;
  let httpMock: HttpTestingController;

  const user: AuthenticatedUser = { id: '1', display_name: 'Jean', email: 'jean@test.fr' };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MeComponent],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(MeComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('loads and displays the current user profile', () => {
    fixture.detectChanges();

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/me`);
    req.flush(user);

    expect(component['user']()).toEqual(user);
    expect(component['loading']()).toBe(false);
  });

  it('shows an error message when the profile request fails', () => {
    fixture.detectChanges();

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/me`);
    req.flush({ error: 'INVALID_TOKEN', message: 'ko' }, { status: 401, statusText: 'Unauthorized' });

    expect(component['errorMessage']()).toBe('Impossible de récupérer votre profil.');
    expect(component['loading']()).toBe(false);
  });
});
