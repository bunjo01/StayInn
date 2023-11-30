import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AvailablePeriodByAccommodation } from 'src/app/model/reservation';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { AuthService } from 'src/app/services/auth.service';
import { ReservationService } from 'src/app/services/reservation.service';

@Component({
  selector: 'app-available-periods',
  templateUrl: './available-periods.component.html',
  styleUrls: ['./available-periods.component.css']
})
export class AvailablePeriodsComponent implements OnInit {
  availablePeriods: AvailablePeriodByAccommodation[] = [];
  accommodation: any;

  constructor(private reservationService: ReservationService,
              private accommodationService: AccommodationService,
              private router: Router,
              private authService: AuthService) {}

  ngOnInit(): void {
    this.getAccommodation();
    this.reservationService.getAvailablePeriods(this.accommodation.id).subscribe(
      data => {
        this.availablePeriods = data;
      },
      error => {
        console.error('Error fetching available periods:', error);
      }
    );
  }

  navigateToAddReservation(period: AvailablePeriodByAccommodation): void {
    period.IDAccommodation = this.accommodation.id
    this.reservationService.sendAvailablePeriod(period);
    this.router.navigate(['/addReservation']);
  }

  navigateToReservations(period: AvailablePeriodByAccommodation): void {
    this.reservationService.sendAvailablePeriod(period)
    this.router.navigate(['/reservations']);
  }

  navigateToAddReservationPeriod(): void {
    this.router.navigate(['/addAvailablePeriod']);
  }

  navigateToEditAvailablePeriod(period: AvailablePeriodByAccommodation): void {
    this.reservationService.sendAvailablePeriod(period)
    this.router.navigate(['/editAvailablePeriod']);
  }

  getAccommodation(): void {
    this.accommodationService.getAccommodation().subscribe((data) => {
      this.accommodation = data;
    });
  }

  isHost(){
    return this.authService.getRoleFromToken() === 'HOST'
  }

  isGuest(){
    return this.authService.getRoleFromToken() === 'GUEST'
  }

}
