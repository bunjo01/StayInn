import { TestBed } from '@angular/core/testing';

import { TimestampInterceptor } from './timestamp.interceptor';

describe('TimestampInterceptor', () => {
  beforeEach(() => TestBed.configureTestingModule({
    providers: [
      TimestampInterceptor
      ]
  }));

  it('should be created', () => {
    const interceptor: TimestampInterceptor = TestBed.inject(TimestampInterceptor);
    expect(interceptor).toBeTruthy();
  });
});
