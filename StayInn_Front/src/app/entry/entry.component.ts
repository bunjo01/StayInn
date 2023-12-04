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
  constructor(private accommodationService: AccommodationService,
    private datePipe: DatePipe){}

  searchAccommodation(location: string, numberOfGuests: number, startDate: string, endDate: string): void {
    // Provera da li su uneti podaci ili ne, i pretraga svih smeštaja ako nisu
    if (!location && !numberOfGuests && !startDate && !endDate) {
      // Poziv servisa za dohvat svih smeštaja
      this.accommodationService.getAccommodations().subscribe(
        (result: Accommodation[]) => {
          this.accommodationService.sendSearchedAccommodations(result);
        },
        (error) => {
          console.log('Error searching accommodations:', error);
        }
      );
    } else {
      // Pretraga smeštaja na osnovu unetih vrednosti
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

    let startDate: string = ''; // Inicijalizacija sa praznim stringom
    let endDate: string = ''; // Inicijalizacija sa praznim stringom

    // Konverzija datuma u format koji želite slati na server
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
