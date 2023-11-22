import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, Subject } from 'rxjs';
import { AvailablePeriodByAccommodation, ReservationByAvailablePeriod, ReservationFormData } from '../model/reservation';
import { DatePipe } from '@angular/common';
import { environment } from 'src/environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ReservationService {
  private baseUrl = environment.baseUrl + '/api/reservations';
  private dataSubject = new BehaviorSubject<AvailablePeriodByAccommodation | null>(null);
  
  constructor(private http: HttpClient,
    private datePipe: DatePipe) {}

  createReservation(reservationData: AvailablePeriodByAccommodation): Observable<any> {

    reservationData.StartDate = this.formatDate(reservationData.StartDate);
    reservationData.EndDate = this.formatDate(reservationData.EndDate);

    return this.http.post(this.baseUrl + '/period', JSON.stringify(reservationData));
  }

  getAvailablePeriods(id: string): Observable<AvailablePeriodByAccommodation[]> {
    return this.http.get<AvailablePeriodByAccommodation[]>(`${this.baseUrl}/${id}/periods`);
  }

  getReservationByAvailablePeriod(id: string): Observable<ReservationByAvailablePeriod[]> {
    return this.http.get<ReservationByAvailablePeriod[]>(`${this.baseUrl}/${id}/reservations`);
  }

  createReservationByAccommodation(reservationData: ReservationByAvailablePeriod): Observable<any> {

    reservationData.StartDate = this.formatDate(reservationData.StartDate);
    reservationData.EndDate = this.formatDate(reservationData.EndDate);

    return this.http.post(this.baseUrl + '/reservation', JSON.stringify(reservationData));
  }

  sendAvailablePeriod(data: AvailablePeriodByAccommodation) {
    this.dataSubject.next(data);
  }

  getAvailablePeriod() {
    return this.dataSubject.asObservable();
  }

  private formatDate(date: string | null): string {
    if (!date) {
      return ''; 
    }

    let formattedDate = this.datePipe.transform(new Date(date), 'yyyy-MM-ddTHH:mm:ssZ');

    if (!formattedDate) {
      return ''; 
    }

    formattedDate = formattedDate?.slice(0, -5)
    formattedDate = formattedDate + "Z"

    return formattedDate;
  }
  
}
