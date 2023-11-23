import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AvailablePeriodByAccommodation, ReservationByAvailablePeriod } from 'src/app/model/reservation';
import { ReservationService } from 'src/app/services/reservation.service';


@Component({
  selector: 'app-add-reservation',
  templateUrl: './add-reservation.component.html',
  styleUrls: ['./add-reservation.component.css']
})
export class AddReservationComponent implements OnInit {
  availablePeriod: any;
  
  formData: ReservationByAvailablePeriod = {
    StartDate: '',
    EndDate: '',
    GuestNumber: 0,
    ID: '',
    IDAccommodation: '',
    IDAvailablePeriod: '',
    IDUser: '',
    Price: 0
  };

  constructor(
    private reservationService: ReservationService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.getAvailablePeriod();
  }

  submitForm() {
    const userId = '655e33ae4b3f315471824211';
    const id = '123e4567-e89b-12d3-a456-426614174022';
    
    this.formData.IDUser = userId;
    this.formData.IDAccommodation = this.availablePeriod.IDAccommodation;
    this.formData.IDAvailablePeriod = this.availablePeriod.ID;
    this.formData.Price = this.availablePeriod.Price;
    this.formData.ID = id;

    console.log(this.formData);

    this.reservationService.createReservationByAccommodation(this.formData)
      .subscribe(response => {
        console.log('Reservation created successfully:', response);
        this.formData = {
          StartDate: '',
          EndDate: '',
          GuestNumber: 0,
          ID: '',
          IDAccommodation: '',
          IDAvailablePeriod: '',
          IDUser: '',
          Price: 0
        };
        this.router.navigate(['/availablePeriods']);
      }, error => {
        console.error('Error creating reservation:', error);
      });
    }

  getAvailablePeriod(): void {
    this.reservationService.getAvailablePeriod().subscribe((data) => {
      this.availablePeriod = data;
    });
  }
}