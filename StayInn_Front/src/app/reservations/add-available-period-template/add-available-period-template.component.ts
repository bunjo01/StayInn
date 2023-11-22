import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ReservationService } from 'src/app/services/reservation.service';
import { AvailablePeriodByAccommodation, ReservationFormData } from 'src/app/model/reservation';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-add-available-period-template',
  templateUrl: './add-available-period-template.component.html',
  styleUrls: ['./add-available-period-template.component.css']
})
export class AddAvailablePeriodTemplateComponent {
  accommodation: any;
  formData: AvailablePeriodByAccommodation = { ID: '', IDAccommodation: '', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };

  constructor(private reservationService: ReservationService,
              private accommodationService: AccommodationService,
              private router: Router
    ) {}

    ngOnInit(): void {
      this.getAccommodation();
    }

  submitForm() {
    let id = '5d353bef-f1e4-4d4e-ad21-f6e084cd96e2'

    this.formData.IDAccommodation = this.accommodation.id;
    this.formData.ID = id;

    this.reservationService.createReservation(this.formData)
      .subscribe(response => {
        console.log('Reservation created successfully:', response);
        this.formData = { ID: '', IDAccommodation: '', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };
      }, error => {
        console.error('Error creating reservation:', error);
      });
      this.router.navigate(['/availablePeriods']);

  }

  getAccommodation(): void {
    this.accommodationService.getAccommodation().subscribe((data) => {
      this.accommodation = data;
    });
  }

}
