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

  constructor(private http: HttpClient, private reservationService: ReservationService, private profileService: ProfileService,
     private accommodationService: AccommodationService , private router: Router, private ratingService: RatingService, private authService: AuthService) { }

  ngOnInit(): void {
      this.loggedinUserUsername = this.getUsernameFromToken();
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

  getUserId(){
    this.profileService.getUser(this.loggedinUserUsername).subscribe((result) => {
      this.loggedinUserId = result.id
    })
  }

  getUsernameFromToken() {
    const token = localStorage.getItem('token');
    if (!token) {
      this.router.navigate(['login']);
      return null;  // Dodajte povratnu vrednost ukoliko token nije prisutan
    }
  
    try {
      const tokenPayload = decode.jwtDecode(token) as JwtPayload;
      return tokenPayload.username;
    } catch (error) {
      console.error('Greška prilikom dekodiranja tokena:', error);
      this.router.navigate(['login']);
      return null;  // Dodajte povratnu vrednost u slučaju greške prilikom dekodiranja tokena
    }
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

  navigateToRateAccommodation(idAccommodation: string) {
    this.ratingService.sendAccommodationID(idAccommodation);
    this.router.navigate(['/rate-accommodation']);
  }  
}