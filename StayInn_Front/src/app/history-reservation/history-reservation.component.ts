import { HttpClient } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { ReservationService } from '../services/reservation.service';
import { ProfileService } from '../services/profile.service';
import { Router } from '@angular/router';
import * as decode from 'jwt-decode';
import { JwtPayload } from 'src/app/model/user';

@Component({
  selector: 'app-history-reservation',
  templateUrl: './history-reservation.component.html',
  styleUrls: ['./history-reservation.component.css']
})
export class HistoryReservationComponent implements OnInit{
  expiredReservations: any[] = [];
  loggedinUserUsername : any;
  loggedinUserId: any;

  constructor(private http: HttpClient, private reservationService: ReservationService, private profileService: ProfileService, private router: Router) { }

  ngOnInit(): void {
      this.loggedinUserUsername = this.getUsernameFromToken();
      this.getUserId();
      this.reservationService.getReservationByUserExp()
      .subscribe((reservations: any[]) => {
        this.expiredReservations = reservations;
      }, (error) => {
        console.error('Greška prilikom dohvatanja isteklih rezervacija:', error);
      });
  }

  getUserId(){
    this.profileService.getUser(this.loggedinUserUsername).subscribe((result) => {
      this.loggedinUserId = result.id
    })
  }

  getUsernameFromToken(){
    const token = localStorage.getItem('token');
    if (token === null) {
      this.router.navigate(['login']);
      return;
    }

    const tokenPayload = decode.jwtDecode(token) as JwtPayload;

    return tokenPayload.username
  }
}
