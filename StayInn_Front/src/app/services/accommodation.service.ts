import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpResponse } from '@angular/common/http';
import { BehaviorSubject, Observable, Subject, of } from 'rxjs';
import { Accommodation } from 'src/app/model/accommodation';
import { environment } from 'src/environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AccommodationService {
  private apiUrl = environment.baseUrl + '/api/accommodations';
  private currentAccommodation = new BehaviorSubject<Accommodation | null>(null);
  private searchedAccommodationsSubject = new Subject<Accommodation[]>();

  jwtToken = localStorage.getItem('token');
  headers = new HttpHeaders({
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${this.jwtToken}`
  });

  constructor(
    private http: HttpClient
  ) { }

  getAccommodations(): Observable<Accommodation[]> {
    return this.http.get<Accommodation[]>(this.apiUrl + '/accommodation');
  }

  sendAccommodation(data: Accommodation) {
    this.currentAccommodation.next(data);
  }

  getAccommodation() {
    return this.currentAccommodation.asObservable();
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
    return this.http.post<Accommodation>(this.apiUrl + '/accommodation', accommodation, { headers: this.headers });
  }
}
