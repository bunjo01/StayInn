import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { User } from '../model/user';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = environment.baseUrl + '/api/auths';

  constructor(
    private http: HttpClient
  ) { }

  register(body: User): Observable<User> {
    return this.http.post<User>(this.apiUrl + '/register', body);
  }
}
