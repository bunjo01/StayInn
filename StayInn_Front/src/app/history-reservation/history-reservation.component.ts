import { HttpClient } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { ReservationService } from '../services/reservation.service';
import { ProfileService } from '../services/profile.service';
import { Router } from '@angular/router';
import * as decode from 'jwt-decode';
import { JwtPayload } from 'src/app/model/user';
import { AccommodationService } from '../services/accommodation.service';
import { Accommodation, DisplayedAccommodation } from '../model/accommodation';
import { RatingService } from '../services/rating.service';
import { RatingAccommodation, RatingHost } from '../model/ratings';
import { AuthService } from '../services/auth.service';
import { finalize, forkJoin } from 'rxjs';
import { DatePipe } from '@angular/common';

@Component({
  selector: 'app-history-reservation',
  templateUrl: './history-reservation.component.html',
  styleUrls: ['./history-reservation.component.css']
})
export class HistoryReservationComponent implements OnInit{
  expiredReservations: any[] = [];
  loggedinUserUsername : any;
  loggedinUserId: any;
  displayedAccommodations: DisplayedAccommodation[] = [];
  userRatings: RatingAccommodation[] = [];
  hostRating: RatingHost[] = [];
  accommodations: Accommodation[] = [];
  accommodationNames: Map<string, string> = new Map();
  loadingAccommodations = false;

  constructor(private http: HttpClient, private reservationService: ReservationService, private profileService: ProfileService,
     private accommodationService: AccommodationService , private router: Router, private ratingService: RatingService, private authService: AuthService, private datePipe: DatePipe) { }

  ngOnInit(): void {
      this.loggedinUserUsername = this.authService.getUsernameFromToken();
      console.log(this.loggedinUserUsername);
      this.getUserId();
      this.reservationService.getReservationByUserExp()
      .subscribe((reservations: any[]) => {
        this.expiredReservations = reservations;
        this.loadAccommodations();
      }, (error) => {
        console.error('Greška prilikom dohvatanja isteklih rezervacija:', error);
      });

      this.ratingService.getRatingsAccommodationByUser().subscribe(
        (ratings: RatingAccommodation[]) => {
          this.userRatings = ratings;
          this.loadAccommodationNames();
        },
        (error) => {
          console.error('Greška prilikom dohvatanja ocjena za smještaje:', error);
        }
      );

      this.ratingService.getAllHostRatingsByUser().subscribe(
        (ratings: RatingHost[]) => {
          this.hostRating = ratings;
        },
        (error) => {
          console.error('Greška prilikom dohvatanja ocjena za smještaje:', error);
        }
      );

  }

  getAccommodationNameById(id: string): string {
    return this.accommodationNames.get(id) || 'Unknown Accommodation';
  }
  

  getUserId(){
    this.profileService.getUser(this.loggedinUserUsername).subscribe((result) => {
      this.loggedinUserId = result.id
    })
  }

  loadAccommodations() {
    this.expiredReservations.forEach((reservation) => {
      this.accommodationService.getAccommodationById(reservation.IDAccommodation).subscribe(
        (accommodation: Accommodation) => {
          const displayedAccommodation: DisplayedAccommodation = {
            reservationInfo: reservation,
            accommodationInfo: accommodation
          };
          this.displayedAccommodations.push(displayedAccommodation);
        },
        (error) => {
          console.error('Greška prilikom dohvatanja informacija o smeštaju:', error);
        }
      );
    });
  }

  loadAccommodationNames() {
    this.loadingAccommodations = true;

    const observables = this.userRatings.map(rating => 
      this.accommodationService.getAccommodationById(rating.idAccommodation)
    );

    forkJoin(observables)
      .pipe(
        finalize(() => this.loadingAccommodations = false)
      )
      .subscribe(
        (accommodations: any[]) => {
          accommodations.forEach((accommodation, index) => {
            this.accommodationNames.set(this.userRatings[index].idAccommodation, accommodation?.name || 'Unknown Accommodation');
          });
        },
        (error) => {
          console.error('Error fetching accommodation names:', error);
        }
      );
  }

  navigateToRateAccommodation(idAccommodation: string) {
    this.ratingService.sendAccommodationID(idAccommodation);
    this.router.navigate(['/rate-accommodation']);
  }
  
  formatDateTime(dateTime: string): string | null {
    return this.datePipe.transform(dateTime, 'dd.MM.yyyy. HH:mm:ss');
  }
}
