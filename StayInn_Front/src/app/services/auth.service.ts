import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { User } from '../model/user';
import { JwtHelperService } from '@auth0/angular-jwt';
import { Router } from '@angular/router';
import * as decode from 'jwt-decode';
import { JwtPayload } from '../model/user';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = environment.baseUrl + '/api/auths';

  constructor(
    private http: HttpClient,
    private jwtHelper: JwtHelperService,
    private router: Router
  ) { }

  register(body: User): Observable<User> {
    return this.http.post<User>(this.apiUrl + '/register', body);
  }

  login(credentials: { username: string, password: string }): Observable<any> {
    return this.http.post<any>(this.apiUrl + '/login', credentials);
  }

  changePassword(requestBody: any): Observable<any> {
    return this.http.post<any>(this.apiUrl + '/change-password', requestBody);
  }

  sendRecoveryMail(requestBody: any): Observable<any> {
    return this.http.post<any>(this.apiUrl + '/recover-password', requestBody)
  }

  resetPassword(requestBody: any): Observable<any> {
    return this.http.post<any>(this.apiUrl + '/recovery-password', requestBody)
  }

  logout(){
    localStorage.removeItem('token')
    this.router.navigate(['login'])
  }

  getToken() {
    return localStorage.getItem('token')
  }

  getRoleFromToken(){
    const token = localStorage.getItem('token');
    if (token === null) {
      this.router.navigate(['login']);
      return;
    }

    const tokenPayload = decode.jwtDecode(token) as JwtPayload;

    return tokenPayload.role
  }

  getRoleFromTokenNoRedirect(){
    const token = localStorage.getItem('token');
    if (token === null) {
      return;
    }

    const tokenPayload = decode.jwtDecode(token) as JwtPayload;

    return tokenPayload.role
  }

  getUsernameFromToken(){
    const token = localStorage.getItem('token');
    if (token === null) {
      this.router.navigate(['login']);
      return;
    }

    const tokenPayload = decode.jwtDecode(token) as JwtPayload;

    return tokenPayload.username
  }

  public isAuthenticated(): boolean {
    const token = localStorage.getItem('token');
    return !this.jwtHelper.isTokenExpired(token);
  }
}
