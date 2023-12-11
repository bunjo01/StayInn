import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RateAccommodationComponent } from './rate-accommodation.component';

describe('RateAccommodationComponent', () => {
  let component: RateAccommodationComponent;
  let fixture: ComponentFixture<RateAccommodationComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [RateAccommodationComponent]
    });
    fixture = TestBed.createComponent(RateAccommodationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
