import { Component, OnInit } from '@angular/core';
import { Accommodation, AmenityEnum } from '../model/accommodation';
import { ActivatedRoute, Router } from '@angular/router';
import { AccommodationService } from '../services/accommodation.service';
import { ToastrService } from 'ngx-toastr';
import { HttpErrorResponse } from '@angular/common/http';

@Component({
  selector: 'app-edit-accommodation',
  templateUrl: './edit-accommodation.component.html',
  styleUrls: ['./edit-accommodation.component.css']
})
export class EditAccommodationComponent implements OnInit{
  currentAccommodation: any;
  accommodation: Accommodation = {
    name: '',
    location: '',
    amenities: [],
    minGuests: 1,
    maxGuests: 1 
  };
  
  amenityValues = Object.values(AmenityEnum).filter(value => typeof value === 'number');
  
  getAmenityName(amenity: number): string {
    return AmenityEnum[amenity];
  }
  constructor(
    private accommodationService: AccommodationService,
    private route: ActivatedRoute,
    private toastr: ToastrService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.getAccommodation();
  }

  getAccommodation(): void {
    this.accommodationService.getAccommodation().subscribe((data) => {
      this.currentAccommodation = data;
      this.accommodation = { ...this.currentAccommodation };
      this.accommodation.amenities = [];
    })
    
  }

  getAmenityValue(amenity: AmenityEnum): AmenityEnum {
    return amenity;
  } 
  
  updateAccommodation(): void {
    this.accommodation.amenities = this.amenityValues
        .filter((_, index) => this.accommodation.amenities[index])
        .map((amenity, ) => amenity as AmenityEnum);
    this.accommodationService.getAccommodationID().subscribe(accommodationId => {
      if (accommodationId) {
        this.accommodationService.updateAccommodation(this.accommodation, accommodationId).subscribe(
          updatedAccommodation => {
            console.log('Accommodation updated:', updatedAccommodation);
            this.toastr.success("Update accommodation successfully", "Accommodation Update")
          },
          error => {
            console.error('Error updating accommodation:', error);
            this.toastr.error("Update accommodation error", "Accommodation Error Update")
          }
        );
      } else {
        console.error('Accommodation ID is not available.');
      }
      this.router.navigate(['']).then(() => {
        location.reload();
      });
    });
  }
}
