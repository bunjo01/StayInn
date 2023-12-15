import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RatingsPopupComponent } from './ratings-popup.component';

describe('RatingsPopupComponent', () => {
  let component: RatingsPopupComponent;
  let fixture: ComponentFixture<RatingsPopupComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [RatingsPopupComponent]
    });
    fixture = TestBed.createComponent(RatingsPopupComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
