import { Component } from '@angular/core';
import { AccommodationService } from '../services/accommodation.service';
import { Accommodation } from '../model/accommodation';

@Component({
  selector: 'app-entry',
  templateUrl: './entry.component.html',
  styleUrls: ['./entry.component.css']
})
export class EntryComponent {
  constructor(private accommodationService: AccommodationService){}

  searchAccommodation(location: string, numberOfGuests: number): void {
    // Provera da li su uneti podaci ili ne, i pretraga svih smeštaja ako nisu
    if (!location && !numberOfGuests) {
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
      this.accommodationService.searchAccommodations(location, numberOfGuests).subscribe(
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

    this.searchAccommodation(location, numberOfGuests);
  }

}
