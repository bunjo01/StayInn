import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ReservationService } from 'src/app/services/reservation.service';
import { AvailablePeriodByAccommodation, ReservationFormData } from 'src/app/model/reservation';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { HttpErrorResponse } from '@angular/common/http';

@Component({
  selector: 'app-add-available-period-template',
  templateUrl: './add-available-period-template.component.html',
  styleUrls: ['./add-available-period-template.component.css']
})
export class AddAvailablePeriodTemplateComponent {
  accommodation: any;
  formData: AvailablePeriodByAccommodation = { ID: '', IDAccommodation: '', IDUser: '', StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };

  constructor(private reservationService: ReservationService,
              private accommodationService: AccommodationService,
              private router: Router,
              private toastr: ToastrService,
    ) {}

    ngOnInit(): void {
      this.getAccommodation();
    }

  submitForm() {
    let id = '5d353bef-f1e4-4d4e-ad21-f6e084cd96e2'
    let idUser = '656e4f2f1cef0b331e349d33'

    this.formData.IDAccommodation = this.accommodation.id;
    this.formData.ID = id;
    this.formData.IDUser = idUser;

    this.reservationService.createReservation(this.formData)
      .subscribe(response => {
        console.log('Period created successfully:', response);
        this.toastr.success('Available period created successfully');
        this.formData = { ID: '', IDAccommodation: '', IDUser:'',StartDate: '', EndDate: '', Price: 0, PricePerGuest: false };
        this.router.navigate(['']);
      }, error => {
        console.error('Error creating available period:', error);
        if (error instanceof HttpErrorResponse) {
          const errorMessage = `${error.error}`;
          this.toastr.error(errorMessage, 'Add Period Error');
        } else {
          this.toastr.error('An unexpected error occurred', 'Add Period Error');
        }
      });

  }

  getAccommodation(): void {
    this.accommodationService.getAccommodation().subscribe((data) => {
      this.accommodation = data;
    });
  }

}
