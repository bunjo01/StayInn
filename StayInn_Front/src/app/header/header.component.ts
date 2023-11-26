import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { ProfileService } from '../services/profile.service';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})
export class HeaderComponent implements OnInit {
  userProfile: any;
  constructor(private router: Router, private profileService: ProfileService) {}

  ngOnInit() {
    // Učitavanje podataka o korisniku prilikom inicijalizacije komponente
    // this.loadUserProfile();
  }

  // loadUserProfile() {
  //   // Pozivamo servis za dobijanje podataka o korisniku
  //   this.profileService.getUser().subscribe(
  //     (data) => {
  //       this.userProfile = data;
  //     },
  //     (error) => {
  //       console.error('Error fetching user profile: ', error);
  //     }
  //   );
  // }

  isUserLoggedIn() {
    const token = localStorage.getItem('token');
    return !!token;
  }

  handleDropdownChange(event: any) {
    const selectedOption = event.target.value;

    if (selectedOption === 'profile') {
      // Otvorite link ka profilu
      this.router.navigate(['/profile']);
    } else if (selectedOption === 'logout') {
      // Izlogujte korisnika
      this.logout();
    }
  }

  logout() {
    localStorage.removeItem('token');
    this.router.navigate(['/']);
  }
}
