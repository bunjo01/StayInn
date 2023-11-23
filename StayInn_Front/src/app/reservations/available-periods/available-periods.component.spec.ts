import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AvailablePeriodsComponent } from './available-periods.component';

describe('AvailablePeriodsComponent', () => {
  let component: AvailablePeriodsComponent;
  let fixture: ComponentFixture<AvailablePeriodsComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [AvailablePeriodsComponent]
    });
    fixture = TestBed.createComponent(AvailablePeriodsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
