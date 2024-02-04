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

  getUsersRatingForAccommodation(accommodationId: string): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<any>(this.baseUrl + `/rating/accommodation/${accommodationId}/byGuest`, { headers });
  }

  getAverageRatingForAccommodation(accommodationId: string): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<any>(this.baseUrl + `/ratings/average/${accommodationId}`, { headers });
  }

  deleteRatingsAccommodationByUser(idRating: string) {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.delete(this.baseUrl + `/rating/accommodation/${idRating}`, { headers });
  }

  getRatingsAccommodationByUser(): Observable<RatingAccommodation[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.get<RatingAccommodation[]>(this.baseUrl + '/ratings/accommodationByUser', { headers });
  }

  getAverageRatingForUser(body: any): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.post<any>(this.baseUrl + `/ratings/average/host`,JSON.stringify(body) ,{ headers });
  }

  getUsersRatingForHost(body: any): Observable<any[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.post<any[]>(this.baseUrl + `/rating/host/byGuest`,JSON.stringify(body), { headers });
  }

  deleteRatingsHostByUser(idRating: string) {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.delete(this.baseUrl + `/rating/host/${idRating}`, { headers });
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

  getRatingHostByUser(idAccommodation: any): Observable<RatingHost> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<RatingHost>(this.baseUrl + `/rating/accommodation/${idAccommodation}`, { headers });
  }

  deleteRatingHostByUser(idRating: string) {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.delete(this.baseUrl + `/rating/host/${idRating}`, { headers });
  }

  sendAccommodationID(data: string) {
    this.dataSubject.next(data);
  }

  getAccommodationID() {
    return this.dataSubject.asObservable();
  }

  getAllHostRatingsByUser(): Observable<RatingHost[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<RatingHost[]>(this.baseUrl + '/ratings/hostByUser', { headers });
  }

  getAllAccommodationRatingsByUser(): Observable<RatingAccommodation[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.get<RatingAccommodation[]>(this.baseUrl + '/ratings/accommodation/byHost', { headers });
  }

  getAllRatingsForAccommodation(accommodationId: string): Observable<RatingAccommodation[]> {
    return this.http.get<RatingAccommodation[]>(this.baseUrl + '/ratings/accommodation/' + accommodationId);
  }

  getAllRatingsForHost(body: any): Observable<RatingHost[]> {
    return this.http.post<RatingHost[]>(this.baseUrl + '/ratings/host/host-ratings', body);
  }

  getAllHostRatings(hostUsername: string): Observable<RatingHost[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<RatingHost[]>(this.baseUrl + '/ratings/host/' + hostUsername, { headers });
  }

}
