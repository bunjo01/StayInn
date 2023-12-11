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
    } else if(selectedOption === 'history-reservation'){
      this.router.navigate(['/history-reservation'])
    }
    
    this.dropdownMenu.nativeElement.value = '';
  }
}
