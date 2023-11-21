import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ReservationFormData } from '../model/reservation';
import { DatePipe } from '@angular/common';
import { environment } from 'src/environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ReservationService {
  private baseUrl = environment.baseUrl + '/api/reservations';

  constructor(private http: HttpClient,
    private datePipe: DatePipe) {}

  createReservation(reservationData: ReservationFormData): Observable<any> {

    reservationData.StartDate = this.formatDate(reservationData.StartDate);
    reservationData.EndDate = this.formatDate(reservationData.EndDate);

    return this.http.post(this.baseUrl + '/period', JSON.stringify(reservationData));
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
