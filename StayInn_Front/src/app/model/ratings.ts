export interface RatingAccommodation {
    ID: string;
    IDUSer: string;
    UsernameUser: string;
    IDAccommodation: string;
    Time: string;
    Rate: number;
}

export interface RatingHost {
    ID: string;
    IDUser: string;
    UsernameUser: string;
    HostUsername: string;
    Time: string;
    Rate: number;
}