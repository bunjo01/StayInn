import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpResponse } from '@angular/common/http';
import { BehaviorSubject, Observable, Subject, of } from 'rxjs';
import { Accommodation } from 'src/app/model/accommodation';
import { environment } from 'src/environments/environment';
import { Image } from '../model/image';

@Injectable({
  providedIn: 'root'
})
export class AccommodationService {
  private apiUrl = environment.baseUrl + '/api/accommodations';
  private currentAccommodation = new BehaviorSubject<Accommodation | null>(null);
  private idAccommodation = new BehaviorSubject<string | null>(null);
  private searchedAccommodationsSubject = new Subject<Accommodation[]>();

  constructor(
    private http: HttpClient
  ) { }

  getAccommodations(): Observable<Accommodation[]> {
    return this.http.get<Accommodation[]>(this.apiUrl + '/accommodation');
  }

  getAccommodationById(id: string): Observable<Accommodation> {
    return this.http.get<Accommodation>(this.apiUrl + `/accommodation/${id}`);
  }

  getAccommodationsByUser(username: string): Observable<Accommodation[]> {
    return this.http.get<Accommodation[]>(this.apiUrl + `/user/${username}/accommodations`);
  }

  updateAccommodation(accommodation: Accommodation, id: string): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.put<any>(this.apiUrl + `/accommodation/${id}`, accommodation, { headers: headers });
  }

  sendAccommodation(data: Accommodation) {
    this.currentAccommodation.next(data);
  }

  sendAccommodationId(accommodationId: string) {
    this.idAccommodation.next(accommodationId);
  }

  getAccommodation() {
    return this.currentAccommodation.asObservable();
  }

  getAccommodationID() {
    return this.idAccommodation.asObservable();
  }

  searchAccommodations(location: string, numberOfGuests: number, startDate: string, endDate: string): Observable<Accommodation[]> {
    let apiUrl = this.apiUrl + `/search?location=${location}&numberOfGuests=${numberOfGuests}`;
  
    // Provera da li postoje vrednosti za startDate i endDate
    if (startDate && endDate) {
      apiUrl += `&startDate=${startDate}&endDate=${endDate}`;
    }
  
    return this.http.get<Accommodation[]>(apiUrl);
  } 

  sendSearchedAccommodations(accommodations: Accommodation[]): void {
    this.searchedAccommodationsSubject.next(accommodations);
  }

  getSearchedAccommodations(): Observable<Accommodation[]> {
    return this.searchedAccommodationsSubject.asObservable();
  }  

  createAccommodation(accommodation: Accommodation): Observable<Accommodation> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.post<Accommodation>(this.apiUrl + '/accommodation', accommodation, { headers });
  }

  createAccommodationImages(images: Image[]): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.post<any>(this.apiUrl + '/accommodation/images', images, { headers });
  }

  getAccommodationImages(accID: string): Observable<Image[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.get<Image[]>(this.apiUrl + '/accommodation/' + accID + '/images');
  }

  deleteAccommodation(id: string): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.delete<any>(this.apiUrl + `/accommodation/${id}`, { headers });
  }
}
