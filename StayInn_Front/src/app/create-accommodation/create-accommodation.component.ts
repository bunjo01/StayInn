import { Component } from '@angular/core';
import { Accommodation, AmenityEnum } from '../model/accommodation';
import { AccommodationService } from '../services/accommodation.service';
import { ToastrService } from 'ngx-toastr';
import { Router } from '@angular/router';

@Component({
  selector: 'app-create-accommodation',
  templateUrl: './create-accommodation.component.html',
  styleUrls: ['./create-accommodation.component.css']
})
export class CreateAccommodationComponent {

  newAccommodation: Accommodation = {
    name: '',
    location: '',
    amenities: [],
    minGuests: 0,
    maxGuests: 0 
  };

  amenityValues = Object.values(AmenityEnum).filter(value => typeof value === 'number');

  constructor(private accommodationService: AccommodationService, private toastr: ToastrService, private router: Router) {}

  createAccommodation(): void {
    // Mapiranje odabranih stavki checkbox-ova u listu brojeva na osnovu AmenityEnum
    if(this.newAccommodation){
      
      this.newAccommodation.amenities = this.amenityValues
        .filter((amenity, index) => this.newAccommodation.amenities[index])
        .map((_, index) => index);
    }

    console.log('Data to be sent:', this.newAccommodation);

    this.accommodationService.createAccommodation(this.newAccommodation).subscribe(
      (createdAccommodation) => {
        this.newAccommodation = { name: '', location: '', amenities: [], minGuests: 0, maxGuests: 0 };
        this.toastr.success('Accommodation created successfully', 'Accommodation');
        this.router.navigate(['']);
      },
      (error) => {
        console.error('Error creating accommodation:', error);
        this.toastr.error('Error creating accommodation', 'Accommodation');
      }
    );
  }

  getAmenityName(amenity: number): string {
    return AmenityEnum[amenity];
  }
}