import { Component, OnInit } from '@angular/core';
import { Accommodation, AmenityEnum } from '../model/accommodation';
import { AccommodationService } from '../services/accommodation.service';
import { Router } from '@angular/router';
import { Image } from '../model/image';

@Component({
  selector: 'app-accommodation-details',
  templateUrl: './accommodation-details.component.html',
  styleUrls: ['./accommodation-details.component.css']
})
export class AccommodationDetailsComponent implements OnInit {
  accommodation: Accommodation | null = null;
  images: Image[] = [];

  constructor(private accommodationService: AccommodationService, private router: Router) {}

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
      console.error('Smeštaj nije definisan ili je null.');
    }
    this.router.navigateByUrl('/update-accommodation');
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

}
