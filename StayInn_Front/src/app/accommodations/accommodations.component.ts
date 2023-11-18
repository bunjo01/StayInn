import { Component, OnInit } from '@angular/core';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { Accommodation } from 'src/app/model/accommodation';
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
      (data) => {
        this.accommodations = data;
      },
      (error) => {
        console.error('Error fetching accommodations:', error);
      }
    );
  }

}
