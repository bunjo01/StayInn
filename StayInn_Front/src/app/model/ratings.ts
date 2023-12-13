export interface RatingAccommodation {
    ID: string;
    IDUSer: string;
    IdHost: string;
    UsernameUser: string;
    HostUsername: string;
    IDAccommodation: string;
    Time: string;
    Rate: number;
}

export interface RatingHost {
    ID: string;
    IDUser: string;
    UsernameUser: string;
    HostID: string;
    HostUsername: string;
    Time: string;
    Rate: number;
}