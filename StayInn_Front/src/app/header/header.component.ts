import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { ProfileService } from '../services/profile.service';
import { AuthService } from '../services/auth.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})
export class HeaderComponent implements OnInit {
  userProfile: any;
  constructor(
    private router: Router, 
    private profileService: ProfileService, 
    private authService: AuthService,
    private toastr: ToastrService
    ) { }

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
      this.router.navigate(['/profile']);
    } else if (selectedOption === 'changePassword') {
      this.router.navigate(['/change-password']);
    } else if (selectedOption === 'logout') {
      this.toastr.info("Logged out");
      this.authService.logout();
    }
  }
}
