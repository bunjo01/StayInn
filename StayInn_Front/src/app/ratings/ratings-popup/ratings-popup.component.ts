import { Component, Inject, OnInit } from '@angular/core';
import { DatePipe } from '@angular/common';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { RatingHost, RatingAccommodation } from 'src/app/model/ratings';
import { RatingService } from 'src/app/services/rating.service';

@Component({
  selector: 'app-ratings-popup',
  templateUrl: './ratings-popup.component.html',
  styleUrls: ['./ratings-popup.component.css']
})
export class RatingsPopupComponent {
  ratingHost: RatingHost[] = [];
  ratingAccommodation: RatingAccommodation[] = [];
  ratingsType: 'host' | 'accommodation';

  constructor(
    public dialogRef: MatDialogRef<RatingsPopupComponent>,
    @Inject(MAT_DIALOG_DATA) public data: any,
    private ratingService: RatingService, 
    private datePipe: DatePipe
  ) { 
    this.ratingsType = data.type;
    }

  ngOnInit(): void {
    if (this.data.type === 'host') {
      this.ratingService.getUsersRatingForHost({ hostId: this.data.hostId }).subscribe(
        (ratings) => {
          this.ratingHost = ratings;
        },
        (error) => {
          console.error('Error fetching host ratings:', error);
        }
      );
    } else if (this.data.type === 'accommodation') {
      this.ratingService.getAllRatingsForAccommodation(this.data.accommodationId).subscribe(
        (ratings) => {
          this.ratingAccommodation = ratings;
          console.log('AccommodationS:', this.ratingAccommodation);
        },
        (error) => {
          console.error('Error fetching accommodation ratings:', error);
        }
      );
    }
  }

  close(): void {
    this.dialogRef.close();
  }

  formatDateTime(dateTime: string): string | null {
    return this.datePipe.transform(dateTime, 'dd.MM.yyyy HH:mm:ss');
  }

}
