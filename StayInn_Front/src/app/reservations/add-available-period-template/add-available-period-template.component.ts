import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ReservationService } from 'src/app/services/reservation.service';
import { ReservationFormData } from 'src/app/model/reservation';

@Component({
  selector: 'app-add-available-period-template',
  templateUrl: './add-available-period-template.component.html',
  styleUrls: ['./add-available-period-template.component.css']
})
export class AddAvailablePeriodTemplateComponent {

  formData: ReservationFormData = { IDAccommodation: '', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };

  constructor(private reservationService: ReservationService) {}

  submitForm() {
    const accommodationId = '3a21db92-0381-4ec8-b983-8f227c004f22';

    this.formData.IDAccommodation = accommodationId;

    this.reservationService.createReservation(this.formData)
      .subscribe(response => {
        console.log('Reservation created successfully:', response);
        this.formData = { IDAccommodation: '', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };
      }, error => {
        console.error('Error creating reservation:', error);
      });
  }

}
