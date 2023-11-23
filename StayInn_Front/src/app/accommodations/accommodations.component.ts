import { Component, OnInit } from '@angular/core';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { Accommodation, AmenityEnum } from 'src/app/model/accommodation';
import { Router } from '@angular/router';
@Component({
  selector: 'app-accommodations',
  templateUrl: './accommodations.component.html',
  styleUrls: ['./accommodations.component.css']
})
export class AccommodationsComponent implements OnInit {
  accommodations: Accommodation[] = [];

  constructor(private accommodationService: AccommodationService,
              private router: Router) {}

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

  navigateToAvailablePeriods(accommodation: Accommodation): void{
    this.accommodationService.sendAccommodation(accommodation);
    this.router.navigate(['/availablePeriods']);
  }
  
  getAmenityName(amenity: number): string {
    return AmenityEnum[amenity];
  }
}
