import { Component, EventEmitter, Input, Output } from '@angular/core';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { RatingService } from 'src/app/services/rating.service';

@Component({
  selector: 'app-rate-accommodation',
  templateUrl: './rate-accommodation.component.html',
  styleUrls: ['./rate-accommodation.component.css']
})
export class RateAccommodationComponent {

  accommodationID: string | null = null;
  constructor(private ratingService: RatingService, private toastr: ToastrService, private router: Router) {
    this.ratingService.getAccommodationID().subscribe(id => {
      this.accommodationID = id;
    });
  }

  ratingA: number = 0;

  setRating(value: number) {
    this.ratingA = value;
    console.log('Selected rating:', this.ratingA);
  }

  addRating() {
    if (this.accommodationID !== null) {
      const ratingData = {
        idAccommodation: this.accommodationID,
        rate: this.ratingA 
      };

      this.ratingService.addRatingAccommodation(ratingData).subscribe(
        (response) => {
          console.log('Rating added successfully:', response);
          this.toastr.success('Rating added successfully')
          this.router.navigate(['/history-reservation']);
        },
        (error) => {
          console.error('Error adding rating:', error);
          this.toastr.error('Error adding rating')
        }
      );
    } else {
      console.error('Accommodation ID is null');
    }
  }
}
