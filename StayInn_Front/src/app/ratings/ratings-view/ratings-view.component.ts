import { Component, OnInit } from '@angular/core';
import { DatePipe } from '@angular/common';
import { RatingHost, RatingAccommodation } from 'src/app/model/ratings';
import { RatingService } from 'src/app/services/rating.service';
import { Accommodation } from 'src/app/model/accommodation';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { AuthService } from 'src/app/services/auth.service';

@Component({
  selector: 'app-ratings-view',
  templateUrl: './ratings-view.component.html',
  styleUrls: ['./ratings-view.component.css'],
  providers: [DatePipe]
})
export class RatingsViewComponent implements OnInit {
  ratingHost: RatingHost[] = [];
  ratingAccommodation: RatingAccommodation[] = [];
  accommodations: Accommodation[] = [];
  username: string = this.authService.getUsernameFromToken() || '';

  constructor(private ratingService: RatingService,private authService: AuthService, private datePipe: DatePipe, private accommodationService: AccommodationService) {}

  ngOnInit() {
    this.getAllHostRatings();
    this.getAllAccommodationRatings();
    this.getAccommodationsByUser();
  }

  getAllHostRatings() {
    this.ratingService.getAllHostRatings(this.username).subscribe(
      (ratings) => {
        this.ratingHost = ratings;
        console.log('All host ratings:', this.ratingHost);
      },
      (error) => {
        console.error('Error fetching all host ratings:', error);
      }
    );
  }

  getAllAccommodationRatings() {
    this.ratingService.getAllAccommodationRatingsByUser().subscribe(
      (ratings) => {
        this.ratingAccommodation = ratings;
        console.log('All accommodation ratings:', this.ratingAccommodation);
      },
      (error) => {
        console.error('Error fetching all accommodation ratings:', error);
      }
    );
  }

  getAccommodationsByUser() {
    this.accommodationService.getAccommodationsByUser(this.username).subscribe(
      (result) => {
        this.accommodations = result;
        console.log('Accommodations:', this.accommodations);
      },
      (error) => {
        console.error('Error fetching accommodations:', error);
      }
    );
  }

  formatDateTime(dateTime: string): string | null {
    return this.datePipe.transform(dateTime, 'dd.MM.yyyy HH:mm:ss');
  }
}
