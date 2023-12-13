import { DatePipe } from '@angular/common';
import { HttpClient, HttpHeaders, HttpResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { RatingAccommodation, RatingHost } from '../model/ratings';

@Injectable({
  providedIn: 'root'
})
export class RatingService {
  private baseUrl = environment.baseUrl + '/api/notifications';
  private dataSubject = new BehaviorSubject<string | null>(null);

  constructor(private http: HttpClient) {}

  getNotifications(username: string): Observable<Notification[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.get<Notification[]>(this.baseUrl + '/' + username, { headers });
  }

  addRatingAccommodation(ratingAccommodation: any): Observable<HttpResponse<any>> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.post(this.baseUrl + '/rating/accommodation', JSON.stringify(ratingAccommodation), {
      headers,
      responseType: 'text',
      observe: 'response'
    });
  }

  getRatingsAccommodationByUser(): Observable<RatingAccommodation[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.get<RatingAccommodation[]>(this.baseUrl + '/ratings/accommodationByUser', { headers });
  }

  getRatingAccommodationByUser(accommodationId: string): Observable<RatingAccommodation> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<RatingAccommodation>(this.baseUrl + '/rating/accommodation/getByAccommodationId' + accommodationId, { headers });
  }

  deleteRatingsAccommodationByUser(idRating: string) {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.delete(this.baseUrl + `/rating/accommodation/${idRating}`, { headers });
  }

  sendAccommodationID(data: string) {
    this.dataSubject.next(data);
  }

  getAccommodationID() {
    return this.dataSubject.asObservable();
  }

  addRatingHost(ratingHost: any): Observable<HttpResponse<any>> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
  
    return this.http.post(this.baseUrl + '/rating/host', JSON.stringify(ratingHost), {
      headers,
      responseType: 'text',
      observe: 'response'
    });
  }

  getRatingHostByUser(host: any): Observable<RatingHost> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<RatingHost>(this.baseUrl + '/rating/host/getByHost', { headers });
  }

  deleteRatingHostByUser(idRating: string) {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.delete(this.baseUrl + `/rating/host/${idRating}`, { headers });
  }
}
