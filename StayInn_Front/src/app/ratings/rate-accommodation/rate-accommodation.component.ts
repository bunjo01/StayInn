import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { Observable } from 'rxjs';
import { Accommodation } from 'src/app/model/accommodation';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { RatingService } from 'src/app/services/rating.service';

@Component({
  selector: 'app-rate-accommodation',
  templateUrl: './rate-accommodation.component.html',
  styleUrls: ['./rate-accommodation.component.css']
})
export class RateAccommodationComponent implements OnInit{

  accommodationID: string | null = null;
  accommodation$: Observable<Accommodation> | undefined;
  constructor(private ratingService: RatingService, private toastr: ToastrService, private router: Router, private accommodationService: AccommodationService) {
  }

  ngOnInit(): void {
    this.ratingService.getAccommodationID().subscribe(id => {
      this.accommodationID = id;
      if (this.accommodationID) {
        this.accommodation$ = this.accommodationService.getAccommodationById(this.accommodationID);
      }
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
          this.toastr.success('Rating added successfully');
          this.router.navigate(['/history-reservation']);
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
