import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
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
              private router: Router,
              private toastr: ToastrService) {}

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
    this.reservationService.deleteReservation(reservation.IDAvailablePeriod, reservation.ID).subscribe(
      (result) => {
        // Uspesno brisanje
        this.toastr.success('Reservation deleted successfully!', 'Success');
        this.refreshPage();
      },
      (error) => {
        // Greška prilikom brisanja
        this.toastr.error('Failed to delete reservation!', 'Error');
        console.error('Error deleting reservation: ', error);
      }
    );
  }

  refreshPage(){
    location.reload();
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
