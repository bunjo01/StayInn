import { Component, OnInit } from '@angular/core';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { Accommodation, AmenityEnum } from 'src/app/model/accommodation';
@Component({
  selector: 'app-accommodations',
  templateUrl: './accommodations.component.html',
  styleUrls: ['./accommodations.component.css']
})
export class AccommodationsComponent implements OnInit {
  accommodations: Accommodation[] = [];

  constructor(private accommodationService: AccommodationService) {}

  ngOnInit(): void {
    this.loadAccommodations();
  }

  loadAccommodations() {
    this.accommodationService.getAccommodations().subscribe(
      (result) => {
        this.accommodations = result;
      },
      (error) => {
        console.error('Error fetching accommodations:', error);
      }
    );
  }
  
  getAmenityName(amenity: number): string {
    return AmenityEnum[amenity];
  }
}
