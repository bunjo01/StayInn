
export interface ReservationFormData {
  IDAccommodation: string;
  IDUser: string;
  StartDate: string;
  EndDate: string;
  Price: number;
  PricePerGuest: boolean; 
}
  
export interface AvailablePeriodByAccommodation {
  ID: string;
  IDAccommodation: string;
  IDUser: string;
  StartDate: string;
  EndDate: string;
  Price: number;
  PricePerGuest: boolean;
}
  
export interface ReservationByAvailablePeriod {
  ID: string;
  IDAccommodation: string;
  IDAvailablePeriod: string;
  IDUser: string;
  StartDate: string;
  EndDate: string;
  GuestNumber: number;
  Price: number;
}
