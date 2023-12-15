import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-profile-menu',
  templateUrl: './profile-menu.component.html',
  styleUrls: ['./profile-menu.component.css']
})
export class ProfileMenuComponent implements OnInit {
  isMenuOpen: boolean = false;
  userRole:any;

  constructor(private router: Router, private authService: AuthService, private toastr: ToastrService) {}
  ngOnInit(): void {
    this.setUserRole();
  }

  toggleMenu() {
    this.isMenuOpen = !this.isMenuOpen;
  }

  navigateTo(route: string) {
    this.router.navigate([route]);
    this.isMenuOpen = false;
    console.log(route);
    if (route === '/logout') {
      this.toastr.info("Logged out");
      this.authService.logout();
    }
  }

  getUsername(): string | undefined {
    return this.authService.getUsernameFromToken();
  }

  setUserRole(){
    this.userRole = this.authService.getRoleFromToken()
  }
}
