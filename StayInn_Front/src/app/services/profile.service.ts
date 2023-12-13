import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';
import { User } from '../model/user';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ProfileService {
  private apiUrl = environment.baseUrl + '/api/profiles';

  constructor(private http: HttpClient) { }

  getUser(username: String): Observable<User> {
    // Uzimanje JWT tokena iz lokalnog skladišta
    const token = localStorage.getItem('token');

    // Postavljanje zaglavlja sa JWT tokenom
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    // Slanje HTTP zahteva sa zaglavljem koje uključuje JWT token
    return this.http.get<User>(this.apiUrl + '/users' + "/" + username, { headers });
  }

  checkUsernameAvailability(username: string): Observable<{ available: boolean }> {
    const url = `${this.apiUrl}/users/check-username/${username}`;
    return this.http.get<{ available: boolean }>(url);
}

  updateUser(username: string, updatedUser: User): Observable<User> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
  
    return this.http.put<User>(`${this.apiUrl}/users/${username}`, updatedUser, { headers });
  }

  deleteUser(username: string) {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.delete(`${this.apiUrl}/users/${username}`, { headers });
  }
}
