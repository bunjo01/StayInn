import { Component, OnInit } from '@angular/core';
import { User } from '../model/user';
import { ProfileService } from '../services/profile.service';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-profile-details',
  templateUrl: './profile-details.component.html',
  styleUrls: ['./profile-details.component.css']
})
export class ProfileDetailsComponent implements OnInit {
  userProfile?: User;
  username: string = this.authService.getUsernameFromToken() || '';

  constructor(private profileService: ProfileService, private authService: AuthService) {}

  ngOnInit(): void {
    this.profileService.getUser(this.username).subscribe(
      (data) => {
        this.userProfile = data;
      },
      (error) => {
        console.error('Error fetching user profile: ', error);
      }
    );
  }
}
