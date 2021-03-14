import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DFComponent } from './df.component';

describe('DFComponent', () => {
  let component: DFComponent;
  let fixture: ComponentFixture<DFComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DFComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DFComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
