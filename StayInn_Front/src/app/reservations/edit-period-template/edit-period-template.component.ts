import { HttpErrorResponse } from '@angular/common/http';
import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { AvailablePeriodByAccommodation } from 'src/app/model/reservation';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { ReservationService } from 'src/app/services/reservation.service';

@Component({
  selector: 'app-edit-period-template',
  templateUrl: './edit-period-template.component.html',
  styleUrls: ['./edit-period-template.component.css']
})
export class EditPeriodTemplateComponent {
  currentAvailablePeriod: any;
  formData: AvailablePeriodByAccommodation = { ID: '', IDUser:'',IDAccommodation: '', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };

  constructor(private reservationService: ReservationService,
              private router:Router,
              private toastr: ToastrService,
              private accommodationService: AccommodationService){
  }

  ngOnInit():void {
    this.getAvailablePeriod()
  }

  submitForm() {
    this.formData.IDAccommodation = this.currentAvailablePeriod.IDAccommodation;
    this.formData.ID = this.currentAvailablePeriod.ID;
    this.formData.IDUser = this.currentAvailablePeriod.IDUser;
    this.reservationService.updateAvailablePeriod(this.formData)
      .subscribe(response => {
        console.log('Period updated successfully:', response);
        this.formData = { ID: '', IDAccommodation: '', IDUser:'', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };
        this.router.navigate(['']);
      }, error => {
        console.error('Error updating period:', error);
        if (error instanceof HttpErrorResponse) {
          const errorMessage = `${error.error}`;
          this.toastr.error(errorMessage, 'Update Period Error');
        } else {
          this.toastr.error('An unexpected error occurred', 'Update Period Error');
        }
      });
  }

  getAvailablePeriod(): void {
    this.reservationService.getAvailablePeriod().subscribe((data) => {
      this.currentAvailablePeriod = data;
      this.currentAvailablePeriod.StartDate = new Date(this.currentAvailablePeriod.StartDate).toISOString().split('T')[0];
      this.currentAvailablePeriod.EndDate = new Date(this.currentAvailablePeriod.EndDate).toISOString().split('T')[0];
      this.formData = { ...this.currentAvailablePeriod };
    })
  }

}
