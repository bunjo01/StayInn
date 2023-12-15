import { ReservationByAvailablePeriod } from "./reservation";

export interface Accommodation {
    id?: string;
    hostID?: string;
    name?: string;
    location?: string;
    amenities: AmenityEnum[];
    minGuests?: number;
    maxGuests?: number;
    image?: string;
}
  
export enum AmenityEnum {
  Essentials = 0,
  WiFi = 1,
  Parking = 2,
  AirConditioning = 3,
  Kitchen = 4,
  TV = 5,
  Pool = 6,
  PetFriendly = 7,
  HairDryer = 8,
  Iron = 9,
  IndoorFireplace = 10,
  Heating = 11,
  Washer = 12,
  Hangers = 13,
  HotWater = 14,
  PrivateBathroom = 15,
  Gym = 16,
  SmokingAllowed = 17,
}

export interface DisplayedAccommodation {
  reservationInfo: ReservationByAvailablePeriod;
  accommodationInfo: Accommodation;
}
