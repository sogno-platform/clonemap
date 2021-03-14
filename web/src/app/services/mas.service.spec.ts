import { TestBed } from '@angular/core/testing';

import { MasService } from './mas.service';

describe('MasService', () => {
  let service: MasService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(MasService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
