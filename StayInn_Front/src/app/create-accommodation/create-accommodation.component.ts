import { Component, ElementRef, ViewChild } from '@angular/core';
import { Accommodation, AmenityEnum } from '../model/accommodation';
import { AccommodationService } from '../services/accommodation.service';
import { ToastrService } from 'ngx-toastr';
import { Router } from '@angular/router';
import { Image } from '../model/image';

@Component({
  selector: 'app-create-accommodation',
  templateUrl: './create-accommodation.component.html',
  styleUrls: ['./create-accommodation.component.css']
})
export class CreateAccommodationComponent {

  @ViewChild('imageInput')
  imageInput!: ElementRef;

  newAccommodation: Accommodation = {
    name: '',
    location: '',
    amenities: [],
    minGuests: 1,
    maxGuests: 1 
  };

  amenityValues = Object.values(AmenityEnum).filter(value => typeof value === 'number');

  imageCounter: number = 0;
  images: Image[] = [];

  constructor(private accommodationService: AccommodationService, private toastr: ToastrService, private router: Router) {}

  createAccommodation(): void {
    if (this.newAccommodation) {
      this.newAccommodation.amenities = this.amenityValues
        .filter((_, index) => this.newAccommodation.amenities[index])
        .map((amenity, _) => amenity as AmenityEnum);
    }

    console.log('Data to be sent:', this.newAccommodation);

    this.accommodationService.createAccommodation(this.newAccommodation).subscribe(
      (createdAccommodation) => {
        this.newAccommodation = { name: '', location: '', amenities: [], minGuests: 0, maxGuests: 0 };

        this.images.forEach(image => {
          image.acc_id = createdAccommodation.id || "";
          console.log(image);
        });

        this.accommodationService.createAccommodationImages(this.images).subscribe(
          () => {
            this.toastr.success('Accommodation created successfully');
            this.router.navigate(['']);
          },
          (error: Error) => {
            console.log(error);
            this.toastr.error("Error while creating images", "Image error");
          }
        );
      },
      (error) => {
        console.error('Error creating accommodation:', error);
        this.toastr.error('Error creating accommodation', 'Accommodation');
      }
    );
  }

  handleImageUpload(): void {
    this.images = [];

    const files = this.imageInput.nativeElement.files;

    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      const reader = new FileReader();

      reader.onload = (e: any) => {
        const imageData = e.target.result.split(',')[1];
        const image: Image = { id: String(this.imageCounter++), acc_id: '', data: imageData };
        this.images.push(image);

        if (this.imageCounter === files.length) {
          const result = window.confirm('Images processed. Do you want to submit data for creation?');
    
          if (result) {
            this.createAccommodation();
          } else {
            return;
          }
        }
      };

      reader.readAsDataURL(file);
    }
  }

  getAmenityName(amenity: number): string {
    return AmenityEnum[amenity];
  }
}