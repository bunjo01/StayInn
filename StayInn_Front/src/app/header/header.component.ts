import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
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
  @ViewChild('dropdownMenu')
  dropdownMenu!: ElementRef;

  constructor(
    private router: Router, 
    private profileService: ProfileService, 
    private authService: AuthService,
    private toastr: ToastrService
    ) { }

  ngOnInit() { }

  isUserLoggedIn() {
    return this.authService.isAuthenticated()
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
    
    this.dropdownMenu.nativeElement.value = '';
  }
}
