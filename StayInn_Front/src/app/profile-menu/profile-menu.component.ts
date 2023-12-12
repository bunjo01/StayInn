import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-profile-menu',
  templateUrl: './profile-menu.component.html',
  styleUrls: ['./profile-menu.component.css']
})
export class ProfileMenuComponent {
  isMenuOpen: boolean = false;

  constructor(private router: Router, private authService: AuthService, private toastr: ToastrService) {}

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
}
