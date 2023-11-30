import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { ReservationByAvailablePeriod } from 'src/app/model/reservation';
import { ReservationService } from 'src/app/services/reservation.service';

@Component({
  selector: 'app-reservations',
  templateUrl: './reservations.component.html',
  styleUrls: ['./reservations.component.css']
})
export class ReservationsComponent {
  reservations: ReservationByAvailablePeriod[] = [];
  availablePeriod: any;

  constructor(private reservationService: ReservationService,
              private router: Router) {}

  ngOnInit(): void {
    this.getAvailablePeriod();
    this.reservationService.getReservationByAvailablePeriod(this.availablePeriod.ID).subscribe(
      (data) => {
      this.reservations = data
    },
    error => {
      console.error('Error fething reservations: ', error)
    })
  }

  deleteReservation(reservation: ReservationByAvailablePeriod) {
    this.reservationService.deleteReservation(reservation.IDAvailablePeriod, reservation.ID).subscribe((result) => {})
    this.router.navigate(['/reservations'])
  }

  navigateToAddReservation(): void {
    this.router.navigate(['/addReservation']);
  }

  getAvailablePeriod(): void {
    this.reservationService.getAvailablePeriod().subscribe((data) => {
      this.availablePeriod = data;
    });
  }

}
