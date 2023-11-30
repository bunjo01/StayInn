import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { AvailablePeriodByAccommodation } from 'src/app/model/reservation';
import { ReservationService } from 'src/app/services/reservation.service';

@Component({
  selector: 'app-edit-period-template',
  templateUrl: './edit-period-template.component.html',
  styleUrls: ['./edit-period-template.component.css']
})
export class EditPeriodTemplateComponent {
  currentAvailablePeriod: any;
  formData: AvailablePeriodByAccommodation = { ID: '', IDAccommodation: '', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };

  constructor(private reservationService: ReservationService,
              private router:Router){
  }

  ngOnInit():void {
    this.getAvailablePeriod()
  }

  submitForm() {
    this.formData.IDAccommodation = this.currentAvailablePeriod.IDAccommodation;
    this.formData.ID = this.currentAvailablePeriod.ID;

    this.reservationService.updateAvailablePeriod(this.formData)
      .subscribe(response => {
        console.log('Period updated successfully:', response);
        this.formData = { ID: '', IDAccommodation: '', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };
      }, error => {
        console.error('Error updating period:', error);
      });
      this.router.navigate(['/availablePeriods']);

  }

  getAvailablePeriod(): void {
    this.reservationService.getAvailablePeriod().subscribe((data) => {
      this.currentAvailablePeriod = data;
    })
  }

}
