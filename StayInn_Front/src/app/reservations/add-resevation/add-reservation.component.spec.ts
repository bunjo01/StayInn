import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AddResevationComponent } from './add-resevation.component';

describe('AddResevationComponent', () => {
  let component: AddResevationComponent;
  let fixture: ComponentFixture<AddResevationComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [AddResevationComponent]
    });
    fixture = TestBed.createComponent(AddResevationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
