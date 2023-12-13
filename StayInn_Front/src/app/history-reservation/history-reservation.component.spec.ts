import { ComponentFixture, TestBed } from '@angular/core/testing';

import { HistoryReservationComponent } from './history-reservation.component';

describe('HistoryReservationComponent', () => {
  let component: HistoryReservationComponent;
  let fixture: ComponentFixture<HistoryReservationComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [HistoryReservationComponent]
    });
    fixture = TestBed.createComponent(HistoryReservationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
