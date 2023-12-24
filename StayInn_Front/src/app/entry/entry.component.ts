import { Component } from '@angular/core';
import { AccommodationService } from '../services/accommodation.service';
import { Accommodation } from '../model/accommodation';
import { DatePipe, formatDate } from '@angular/common';

@Component({
  selector: 'app-entry',
  templateUrl: './entry.component.html',
  styleUrls: ['./entry.component.css']
})
export class EntryComponent {
  role: string = "";
  constructor (
    private accommodationService: AccommodationService,
    private datePipe: DatePipe
    ) { }

  searchAccommodation(location: string, numberOfGuests: number, startDate: string, endDate: string): void {
    // Get all accommodations if nothing is set as search param
    if (!location && !numberOfGuests && !startDate && !endDate) {
      this.accommodationService.getAccommodations().subscribe(
        (result: Accommodation[]) => {
          this.accommodationService.sendSearchedAccommodations(result);
        },
        (error) => {
          console.log('Error searching accommodations:', error);
        }
      );
    } else {
      // Searching accommodations based on params
      this.accommodationService.searchAccommodations(location, numberOfGuests, startDate, endDate).subscribe(
        (result: Accommodation[]) => {
          this.accommodationService.sendSearchedAccommodations(result);
        },
        (error) => {
          console.log('Error searching accommodations:', error);
        }
      );
    }
  }

  onSearchSubmit(): void {
    const location = (document.getElementById('location') as HTMLInputElement).value;
    const numberOfGuests = parseInt((document.getElementById('numberOfGuests') as HTMLInputElement).value, 10);

    const startDateValue = (document.getElementById('check-in') as HTMLInputElement).value;
    const endDateValue = (document.getElementById('check-out') as HTMLInputElement).value;

    let startDate: string = '';
    let endDate: string = '';

    // Converting date to format accepted by the backend
    if (startDateValue && endDateValue) {
      startDate = this.formatDate(startDateValue);
      endDate = this.formatDate(endDateValue);
    }

    this.searchAccommodation(location, numberOfGuests, startDate, endDate);
  }

  private formatDate(date: string | null): string {
    if (!date) {
      return ''; 
    }

    let formattedDate = this.datePipe.transform(new Date(date), 'yyyy-MM-ddTHH:mm:ssZ');

    if (!formattedDate) {
      return ''; 
    }

    formattedDate = formattedDate?.slice(0, -5)
    formattedDate = formattedDate + "Z"

    return formattedDate;
  }
}
