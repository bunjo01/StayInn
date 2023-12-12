import { DatePipe } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { RatingAccommodation } from '../model/ratings';

@Injectable({
  providedIn: 'root'
})
export class RatingService {
  private baseUrl = environment.baseUrl + '/api/notifications';
  private dataSubject = new BehaviorSubject<string | null>(null);

  constructor(private http: HttpClient) {}

  addRatingAccommodation(ratingAccommodation: any): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.post<any>(this.baseUrl + '/rating/accommodation', JSON.stringify(ratingAccommodation), { headers });
  }


  getRatingsAccommodationByUser(): Observable<RatingAccommodation[]>{
  return this.http.get<RatingAccommodation[]>(this.baseUrl +   '/ratings/accommodationByUser');
  }

  deleteRatingsAccommodationByUser(idRating: string){
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.delete(this.baseUrl + `/rating/${idRating}`, { headers });
  }

  sendAccommodationID(data: string) {
    this.dataSubject.next(data);
  }

  getAccommodationID() {
    return this.dataSubject.asObservable();
  }

}
