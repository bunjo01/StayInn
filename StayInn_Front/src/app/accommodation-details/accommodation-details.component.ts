import { Component, OnInit } from '@angular/core';
import { Accommodation, AmenityEnum } from '../model/accommodation';
import { AccommodationService } from '../services/accommodation.service';
import { Router } from '@angular/router';
import { Image } from '../model/image';
import { DomSanitizer, SafeResourceUrl } from '@angular/platform-browser';
import { MatDialog } from '@angular/material/dialog';
import { RatingsPopupComponent } from '../ratings/ratings-popup/ratings-popup.component';

@Component({
  selector: 'app-accommodation-details',
  templateUrl: './accommodation-details.component.html',
  styleUrls: ['./accommodation-details.component.css']
})
export class AccommodationDetailsComponent implements OnInit {
  accommodation: Accommodation | null = null;
  images: Image[] = [];

  constructor(
    private accommodationService: AccommodationService, 
    private router: Router,
    private sanitizer: DomSanitizer,
    private dialog: MatDialog
    ) { }

  ngOnInit(): void {
    this.accommodationService.getAccommodation().subscribe(
      data => {
        this.accommodation = data;
        const accId: string = this.accommodation?.id || "";

        this.accommodationService.getAccommodationImages(accId).subscribe(
          (images: Image[]) => {
            console.log(images);
            this.images = images;
          },
          (error: Error) => {
            console.log(error);
          }
        );
      },
      error => {
        console.error('Error fetching accommodation details:', error);
      }
    );
  }

  getAmenityName(amenity: number): string {
    return AmenityEnum[amenity];
  }

  navigateToUpdateAccommodation(id: string): void{
    this.accommodationService.sendAccommodationId(id);
    if (this.accommodation) {
      this.accommodationService.sendAccommodation(this.accommodation);
    } else {
      console.error('Accommodation not defined or null');
    }
    this.router.navigateByUrl('/update-accommodation');
  }

  deleteAccommodation(id: string): void {
    this.accommodationService.deleteAccommodation(id).subscribe(
      () => {
        console.log('Smeštaj uspešno obrisan.');
        this.router.navigate([''])
      },
      error => {
        console.error('Greška prilikom brisanja smeštaja:', error);
      }
    );
  }

  amenityIcons: { [key: number]: string } = {
    0: "../../assets/images/essentials.png",
    1: "../../assets/images/wifi.png",
    2: "../../assets/images/parking.png",
    3: "../../assets/images/air-condition.png",
    4: "../../assets/images/kitchen.png",
    5: "../../assets/images/tv.png",
    6: "../../assets/images/pool.png",
    7: "../../assets/images/petFriendly.png",
    8: "../../assets/images/hairDryer.png",
    9: "../../assets/images/iron.png",
    10: "../../assets/images/indoorFireplace.png",
    11: "../../assets/images/heating.png",
    12: "../../assets/images/washer.png",
    13: "../../assets/images/hanger.png",
    14: "../../assets/images/hotWater.png",
    15: "../../assets/images/privateBathroom.png",
    16: "../../assets/images/gym.png",
    17: "../../assets/images/smokingAllowed.png"
  };

  getSafeImage(base64Image: string): SafeResourceUrl {
    // Determine the image type based on the content
    const isPng = base64Image.startsWith('/9j/') || base64Image.startsWith('iVBOR');
    const isJpeg = base64Image.startsWith('/8A') || base64Image.startsWith('/9A') || base64Image.startsWith('R0lGOD');
  
    // Default to PNG if neither PNG nor JPEG is detected
    const imageType = isPng ? 'image/png' : (isJpeg ? 'image/jpeg' : 'image/png');
  
    // Construct the data URL with the detected image type
    const imageUrl = `data:${imageType};base64,${base64Image}`;
  
    // Return the sanitized URL
    return this.sanitizer.bypassSecurityTrustResourceUrl(imageUrl);
  }

  showHostRatings(hostId: string): void {
    const dialogRef = this.dialog.open(RatingsPopupComponent, {
      width: '600px',
      height: 'fit-content',
      data: { type: 'host', hostId: hostId }
    });
  
    dialogRef.afterClosed().subscribe(result => {
      console.log('The dialog was closed');
    });
  }
  
  showAccommodationRatings(accommodationId: string): void {
    const dialogRef = this.dialog.open(RatingsPopupComponent, {
      width: '600px',
      height: 'fit-content',
      data: { type: 'accommodation', accommodationId: accommodationId }
    });
  
    dialogRef.afterClosed().subscribe(result => {
      console.log('The dialog was closed');
    });
  }
}
