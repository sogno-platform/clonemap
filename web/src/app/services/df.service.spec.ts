import { TestBed } from '@angular/core/testing';

import { DfService } from './df.service';

describe('DfService', () => {
  let service: DfService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(DfService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
