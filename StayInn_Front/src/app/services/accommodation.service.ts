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

  // private accommodations: Accommodation[] = [
  //   { id: '1', location: 'Location 1', image: 'assets/images/stay-inn.jpg' },
  //   { id: '2', location: 'Location 2', image: 'assets/images/stay-inn.jpg' },
  //   { id: '3', location: 'Location 3', image: 'assets/images/stay-inn.jpg' },
  //   { id: '4', location: 'Location 4', image: 'assets/images/stay-inn.jpg' },
  //   { id: '5', location: 'Location 1', image: 'assets/images/stay-inn.jpg' },
  //   { id: '6', location: 'Location 2', image: 'assets/images/stay-inn.jpg' },
  //   { id: '7', location: 'Location 3', image: 'assets/images/stay-inn.jpg' },
  //   { id: '8', location: 'Location 4', image: 'assets/images/stay-inn.jpg' },
  //   { id: '9', location: 'Location 1', image: 'assets/images/stay-inn.jpg' },
  //   { id: '10', location: 'Location 2', image: 'assets/images/stay-inn.jpg' },
  //   { id: '11', location: 'Location 3', image: 'assets/images/stay-inn.jpg' },
  //   { id: '12', location: 'Location 4', image: 'assets/images/stay-inn.jpg' },
  // ];

  // getAccommodations(): Observable<Accommodation[]> {
  //   return of(this.accommodations);
  // }
}
