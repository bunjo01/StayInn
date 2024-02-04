import { Component, Input, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { Observable } from 'rxjs';
import { Accommodation } from 'src/app/model/accommodation';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { AuthService } from 'src/app/services/auth.service';
import { RatingService } from 'src/app/services/rating.service';

@Component({
  selector: 'app-rate-accommodation',
  templateUrl: './rate-accommodation.component.html',
  styleUrls: ['./rate-accommodation.component.css']
})
export class RateAccommodationComponent implements OnInit {
  @Input() accommodationID: string | null = null;
  @Input() hostId: string | null = null;
  accommodation$: Observable<Accommodation> | undefined;
  ratingA: number = 0;
  guestsRatingOfAccommodation: any;
  averageRating:any;
  userRole:any;

  constructor(
    private ratingService: RatingService,
    private toastr: ToastrService,
    private router: Router,
    private accommodationService: AccommodationService,
    private authService: AuthService
  ) { }

  ngOnInit(): void {
    if (this.accommodationID) {
      this.accommodation$ = this.accommodationService.getAccommodationById(this.accommodationID);
      this.setGuestsRatingOfAccommodation();
      this.setAverageRatingForAccommodation();
      this.setUserRole();
    }
  }

  setRating(value: number) {
    this.ratingA = value;
  }

  setGuestsRatingOfAccommodation() {
    if (this.accommodationID != null) {
      this.ratingService.getUsersRatingForAccommodation(this.accommodationID).subscribe((result) => {
        this.guestsRatingOfAccommodation = result
      });
    }
  }

  setAverageRatingForAccommodation() {
    if (this.accommodationID != null) {
      this.ratingService.getAverageRatingForAccommodation(this.accommodationID).subscribe((result) => {
        this.averageRating = result
      })
    }
  }

  deleteAccommodationRating() {
    if (this.accommodationID != null) {
      this.ratingService.deleteRatingsAccommodationByUser(this.guestsRatingOfAccommodation.id).subscribe((result) => {})      
    }
    this.router.navigate(['']);
  }

  setUserRole() {
    this.userRole = this.authService.getRoleFromTokenNoRedirect();
  }

  addRating() {
    if (this.accommodationID !== null && this.hostId !== null) {
      const ratingData = {
        idAccommodation: this.accommodationID,
        idHost: this.hostId,
        rate: this.ratingA
      };

      this.ratingService.addRatingAccommodation(ratingData).subscribe(
        (response) => {
          console.log('Rating added successfully:', response);
          this.toastr.success('Rating added successfully');
          this.router.navigate(['']);
        },
        (error) => {
          console.error('Error adding rating:', error);
          this.toastr.error('Error adding rating');
        }
      );
    } else {
      console.error('Accommodation ID is null');
    }
  }
}
