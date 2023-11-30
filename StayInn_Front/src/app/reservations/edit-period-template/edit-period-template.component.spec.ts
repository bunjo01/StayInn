import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EditPeriodTemplateComponent } from './edit-period-template.component';

describe('EditPeriodTemplateComponent', () => {
  let component: EditPeriodTemplateComponent;
  let fixture: ComponentFixture<EditPeriodTemplateComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [EditPeriodTemplateComponent]
    });
    fixture = TestBed.createComponent(EditPeriodTemplateComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
