import { Injectable } from '@angular/core';
import {
  HttpRequest,
  HttpHandler,
  HttpEvent,
  HttpInterceptor
} from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable()
export class TimestampInterceptor implements HttpInterceptor {

  constructor() {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const timestamp = Date.now().toString();
    const modifiedRequest = request.clone({
      setHeaders: {'X-Timestamp': timestamp},
    });

    return next.handle(modifiedRequest);
  }
}
