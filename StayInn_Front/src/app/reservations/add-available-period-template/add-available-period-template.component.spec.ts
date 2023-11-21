import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AddAvailablePeriodTemplateComponent } from './add-available-period-template.component';

describe('AddAvailablePeriodTemplateComponent', () => {
  let component: AddAvailablePeriodTemplateComponent;
  let fixture: ComponentFixture<AddAvailablePeriodTemplateComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AddAvailablePeriodTemplateComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(AddAvailablePeriodTemplateComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
