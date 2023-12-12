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
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    reservationData.StartDate = this.formatDate(reservationData.StartDate);
    reservationData.EndDate = this.formatDate(reservationData.EndDate);

    return this.http.post(this.baseUrl + '/period', JSON.stringify(reservationData), { headers });
  }

  getAvailablePeriods(id: string): Observable<AvailablePeriodByAccommodation[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.get<AvailablePeriodByAccommodation[]>(`${this.baseUrl}/${id}/periods`, { headers });
  }

  getReservationByAvailablePeriod(id: string): Observable<ReservationByAvailablePeriod[]> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.get<ReservationByAvailablePeriod[]>(`${this.baseUrl}/${id}/reservations`, { headers });
  }

  getReservationByUserExp(): Observable<ReservationByAvailablePeriod[]> {
    const token = localStorage.getItem('token')
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.get<ReservationByAvailablePeriod[]>(`${this.baseUrl}/expired`, { headers });
  }

  createReservationByAccommodation(reservationData: ReservationByAvailablePeriod): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    reservationData.StartDate = this.formatDate(reservationData.StartDate);
    reservationData.EndDate = this.formatDate(reservationData.EndDate);

    return this.http.post(this.baseUrl + '/reservation', JSON.stringify(reservationData), { headers });
  }

  updateAvailablePeriod(periodData: AvailablePeriodByAccommodation): Observable<any> {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    periodData.StartDate = this.formatDate(periodData.StartDate);
    periodData.EndDate = this.formatDate(periodData.EndDate);

    return this.http.patch(this.baseUrl + '/period', JSON.stringify(periodData), { headers });
  }

  deleteReservation(idPeriod: string , idReservations : string) {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    return this.http.delete(this.baseUrl + `/${idPeriod}/${idReservations}`, { headers });
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
