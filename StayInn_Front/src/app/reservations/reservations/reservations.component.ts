import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { ReservationByAvailablePeriod } from 'src/app/model/reservation';
import { ProfileService } from 'src/app/services/profile.service';
import { ReservationService } from 'src/app/services/reservation.service';
import { AuthService } from 'src/app/services/auth.service';

@Component({
  selector: 'app-reservations',
  templateUrl: './reservations.component.html',
  styleUrls: ['./reservations.component.css']
})
export class ReservationsComponent {
  reservations: ReservationByAvailablePeriod[] = [];
  availablePeriod: any;
  loggedinUserUsername : any;
  loggedinUserId: any;

  constructor(private reservationService: ReservationService,
              private router: Router,
              private toastr: ToastrService,
              private profileService: ProfileService,
              private authService: AuthService) {}

  ngOnInit(): void {
    this.getAvailablePeriod();
    this.loggedinUserUsername = this.authService.getRoleFromToken();
    this.getUserId();
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
        this.toastr.success('Reservation deleted successfully!', 'Success');
        this.router.navigate(['']);
      },
      (error) => {
        this.toastr.error('Failed to delete reservation!', 'Error');
        console.error('Error deleting reservation: ', error);
      }
    );
  }

  navigateToAddReservation(): void {
    this.router.navigate(['/addReservation']);
  }

  getAvailablePeriod(): void {
    this.reservationService.getAvailablePeriod().subscribe((data) => {
      this.availablePeriod = data;
    });
  }

  isOwnerOfReservation(userId: string){
    return this.loggedinUserId == userId
  }

  getUserId(){
    this.profileService.getUser(this.loggedinUserUsername).subscribe((result) => {
      this.loggedinUserId = result.id
    })
  }
}
