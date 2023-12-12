import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ProfileMenuComponent } from './profile-menu.component';

describe('ProfileMenuComponent', () => {
  let component: ProfileMenuComponent;
  let fixture: ComponentFixture<ProfileMenuComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ProfileMenuComponent]
    });
    fixture = TestBed.createComponent(ProfileMenuComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
