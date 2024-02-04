import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RatingsViewComponent } from './ratings-view.component';

describe('RatingsViewComponent', () => {
  let component: RatingsViewComponent;
  let fixture: ComponentFixture<RatingsViewComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [RatingsViewComponent]
    });
    fixture = TestBed.createComponent(RatingsViewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
