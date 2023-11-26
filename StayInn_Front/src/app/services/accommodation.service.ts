import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpResponse } from '@angular/common/http';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { Accommodation } from 'src/app/model/accommodation';
import { environment } from 'src/environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AccommodationService {
  private apiUrl = environment.baseUrl + '/api/accommodations';
  private currentAccommodation = new BehaviorSubject<Accommodation | null>(null);

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
}
