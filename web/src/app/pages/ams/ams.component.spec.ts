import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AMSComponent } from './ams.component';

describe('AMSComponent', () => {
  let component: AMSComponent;
  let fixture: ComponentFixture<AMSComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ AMSComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(AMSComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
