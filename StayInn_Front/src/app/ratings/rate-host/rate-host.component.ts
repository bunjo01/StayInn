import { Component, Input, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { Accommodation } from 'src/app/model/accommodation';
import { RatingHost } from 'src/app/model/ratings';
import { RatingService } from 'src/app/services/rating.service';

@Component({
  selector: 'app-rate-host',
  templateUrl: './rate-host.component.html',
  styleUrls: ['./rate-host.component.css']
})
export class RateHostComponent implements OnInit{
  @Input() hostId: string | null = null;
  ratingH: number = 0;
  currentRating: RatingHost | null = null;
  guestsRate: any;
  averageHostRate: any;

  constructor(private ratingService: RatingService, private toastr: ToastrService, private router: Router) {}

  ngOnInit(): void {
    if (this.hostId){
      this.getHostsRateByGuest()
      this.getHostsAverageRate()
    }
  }

  setRating(value: number) {
    this.ratingH = value;
    console.log('Selected rating:', this.ratingH);
  }

  addRating() {
    if (this.hostId !== null) {
      const ratingData = {
        idHost: this.hostId,
        rate: this.ratingH
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

  getHostsRateByGuest(){
    let id = this.hostId
    const body = {"id":id}
    this.ratingService.getUsersRatingForHost(body).subscribe((result) => {
      this.guestsRate = result
    })
  }

  getHostsAverageRate(){
    let id = this.hostId
    const body = {"id":id}
    this.ratingService.getAverageRatingForUser(body).subscribe((result) => {
      this.averageHostRate = result
    })
  }

  deleteRatingsHostByUser(){
    if(this.hostId != null){
      this.ratingService.deleteRatingsHostByUser(this.guestsRate.id).subscribe((result) => {})      
    }
  }

  getRating() {
    if (this.hostId !== null) {
      this.ratingService.getRatingHostByUser(this.hostId).subscribe(
        (rating) => {
          this.currentRating = rating;
          console.log('Current host rating:', this.currentRating);
        },
        (error) => {
          console.error('Error fetching host rating:', error);
        }
      );
    }
  }

}
