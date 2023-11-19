import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AccommodationsComponent } from './accommodations.component';

describe('AccommodationsComponent', () => {
  let component: AccommodationsComponent;
  let fixture: ComponentFixture<AccommodationsComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [AccommodationsComponent]
    });
    fixture = TestBed.createComponent(AccommodationsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
