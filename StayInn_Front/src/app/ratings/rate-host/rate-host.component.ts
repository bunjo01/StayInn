import { Component, Input } from '@angular/core';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { RatingService } from 'src/app/services/rating.service';

@Component({
  selector: 'app-rate-host',
  templateUrl: './rate-host.component.html',
  styleUrls: ['./rate-host.component.css']
})
export class RateHostComponent {
  @Input() hostUsername: string | null = null;
  ratingA: number = 0;

  constructor(private ratingService: RatingService, private toastr: ToastrService, private router: Router) {}

  setRating(value: number) {
    this.ratingA = value;
    console.log('Selected rating:', this.ratingA);
  }

  addRating() {
    if (this.hostUsername !== null) {
      const ratingData = {
        hostUsername: this.hostUsername,
        rate: this.ratingA
      };

      this.ratingService.addRatingHost(ratingData).subscribe(
        (response) => {
          console.log('Host rating added successfully:', response);
          this.toastr.success('Host rating added successfully');
          this.router.navigate(['']);
        },
        (error) => {
          console.error('Error adding host rating:', error);
          this.toastr.error('Error adding host rating');
        }
      );
    } else {
      console.error('Host username is null');
    }
  }

}
