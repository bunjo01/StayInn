import { Component, OnDestroy, OnInit } from '@angular/core';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { Accommodation, AmenityEnum } from 'src/app/model/accommodation';
import { Router } from '@angular/router';
import { Subscription } from 'rxjs';
import { AuthService } from '../services/auth.service';
@Component({
  selector: 'app-accommodations',
  templateUrl: './accommodations.component.html',
  styleUrls: ['./accommodations.component.css']
})
export class AccommodationsComponent implements OnInit, OnDestroy {
  role: string = "";
  accommodations: Accommodation[] = [];
  showCreateAccommodationForm: boolean = false;
  private accommodationSubscription!: Subscription;

  constructor(private accommodationService: AccommodationService,
              private router: Router,
              private authService: AuthService) {}

  ngOnDestroy(): void {
    this.accommodationSubscription.unsubscribe();
  }

  ngOnInit(): void {
    this.loadAccommodations();
    this.listenForSearchedAccommodations();
    this.role = this.authService.getRoleFromTokenNoRedirect() || "";
  }

  toggleCreateAccommodationForm(): void {
    this.showCreateAccommodationForm = true;
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

  listenForSearchedAccommodations(): void {
    this.accommodationSubscription = this.accommodationService.getSearchedAccommodations().subscribe(
      (searchedAccommodations) => {
        this.accommodations = searchedAccommodations;
      },
      (error) => {
        console.error('Error fetching searched accommodations:', error);
      }
    );
    
  }

  navigateToAvailablePeriods(accommodation: Accommodation): void{
    this.accommodationService.sendAccommodation(accommodation);
    this.router.navigate(['/availablePeriods']);
  }

  navigateToAccommodationDetails(accommodation: Accommodation): void{
    this.accommodationService.sendAccommodation(accommodation);
    this.router.navigate(['/accommodation-details']);
  }
  
  getAmenityName(amenity: number): string {
    return AmenityEnum[amenity];
  }
}
