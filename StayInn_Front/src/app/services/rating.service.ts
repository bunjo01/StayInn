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

  getUsersRatingForAccommodation(accommodationId: string): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<any>(this.baseUrl + `/rating/accommodation/${accommodationId}/byGuest`, {headers: headers});
  }

  getAverageRatingForAccommodation(accommodationId: string): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.get<any>(this.baseUrl + `/ratings/average/${accommodationId}`, {headers: headers});
  }

  getUsersRatingForHost(body: any): Observable<any[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return this.http.post<any[]>(this.baseUrl + `/rating/host/byGuest`,JSON.stringify(body), {headers: headers});
  }

  deleteRatingsAccommodationByUser(idRating: string) {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    console.log("DELETE URL:", this.baseUrl + `/rating/accommodation/${idRating}` )
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
}
