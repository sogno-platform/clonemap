import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LogseriesComponent } from './logseries.component';

describe('LogseriesComponent', () => {
  let component: LogseriesComponent;
  let fixture: ComponentFixture<LogseriesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ LogseriesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(LogseriesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
