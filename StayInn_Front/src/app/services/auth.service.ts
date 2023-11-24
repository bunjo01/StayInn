import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { User } from '../model/user';
import { JwtHelperService } from '@auth0/angular-jwt';

interface JwtPayload {
  role: string; 
}


@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = environment.baseUrl + '/api/auths';

  constructor(
    private http: HttpClient,
    private jwtHelper: JwtHelperService
  ) { }

  register(body: User): Observable<User> {
    return this.http.post<User>(this.apiUrl + '/register', body);
  }

  login(credentials: { username: string, password: string }): Observable<any> {
    return this.http.post<any>(this.apiUrl + '/login', credentials);
  }

  public isAuthenticated(): boolean {
    // const token = localStorage.getItem('token');
    const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOiIxNzAwOTQzMzMzIiwidXNlcm5hbWUiOiJidW5qbyJ9.2xVsKyUBl4Fr0ziu2OLTOxo07jvYSrc0_ibcNwA4pHE'

    console.log(token);
    return !this.jwtHelper.isTokenExpired(token);
  }
}
