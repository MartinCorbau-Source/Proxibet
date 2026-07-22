import { TestBed } from '@angular/core/testing';

import { AuthenticatedUser } from './auth.models';
import { CurrentUserStore } from './current-user.store';

describe('CurrentUserStore', () => {
  let store: CurrentUserStore;

  const user: AuthenticatedUser = { id: '1', display_name: 'Jean', email: 'jean@test.fr' };

  beforeEach(() => {
    TestBed.configureTestingModule({});
    store = TestBed.inject(CurrentUserStore);
  });

  it('starts with no user', () => {
    expect(store.user()).toBeNull();
    expect(store.isAuthenticated()).toBe(false);
  });

  it('set updates user and isAuthenticated', () => {
    store.set(user);

    expect(store.user()).toEqual(user);
    expect(store.isAuthenticated()).toBe(true);
  });

  it('clear resets to unauthenticated', () => {
    store.set(user);
    store.clear();

    expect(store.user()).toBeNull();
    expect(store.isAuthenticated()).toBe(false);
  });
});
