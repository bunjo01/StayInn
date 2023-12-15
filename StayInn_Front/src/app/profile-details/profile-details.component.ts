import { Component, OnInit } from '@angular/core';
import { User } from '../model/user';
import { ProfileService } from '../services/profile.service';
import { AuthService } from '../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ReservationService } from '../services/reservation.service';
import { ReservationByAvailablePeriod } from '../model/reservation';
import { AccommodationService } from '../services/accommodation.service';
import { Accommodation } from '../model/accommodation';
import { RatingService } from '../services/rating.service';

@Component({
  selector: 'app-profile-details',
  templateUrl: './profile-details.component.html',
  styleUrls: ['./profile-details.component.css']
})
export class ProfileDetailsComponent implements OnInit {
  form!: FormGroup;
  userProfile?: User;
  role: string = "";
  reservations?: ReservationByAvailablePeriod[] = [];
  accommodations?: Accommodation[] = [];
  notifications?: Notification[] = [];
  username: string = this.authService.getUsernameFromToken() || '';
  updateSuccess: boolean = false;

  constructor(
    private profileService: ProfileService,
    private authService: AuthService,
    private reservationService: ReservationService,
    private accommodationService: AccommodationService,
    private ratingService: RatingService,
    private toastr: ToastrService,
    private fb: FormBuilder
    ) { }

  ngOnInit(): void {
    this.loadUserProfile();

    this.form = this.fb.group({
      username: [null, Validators.required],
      firstName: [null, Validators.required],
      lastName: [null, Validators.required],
      email: [null, Validators.email],
      address: [null, Validators.required],
    });

    this.role = this.authService.getRoleFromToken() || "";
  }

  loadUserProfile() {
    this.profileService.getUser(this.username).subscribe(
      (data) => {
        this.userProfile = data;

        if (this.role === "GUEST" && this.userProfile.id) {
          this.reservationService.getReservationByUser(this.username).subscribe( data => {
            this.reservations = data;
          }, error => {
            console.log("Error getting reservations for guest: ", error);
          });
        } 
        else if (this.role === "HOST" && this.userProfile.id) {
          this.accommodationService.getAccommodationsByUser(this.username).subscribe( data => {
            this.accommodations = data;
          }, error => { 
            console.log("Error getting accommodations for host", error) 
          });

          this.ratingService.getNotifications(this.username).subscribe( data => {
            this.notifications = data;
            this.notifications.reverse();
          }, error => {
            console.log("Error getting notifications for host", error);
          })
        }

      },
      (error) => {
        console.error('Error fetching user profile: ', error);
      }
    );
  }

  updateProfile() {
    if (this.form.invalid) {
      this.toastr.warning('Please fill in the required fields correctly', 'Form Validation');
      return;
    }

    const usernameRegex: RegExp = /^(?=.{3,20}$)(?![_.])(?!.*[_.]{2})[a-zA-Z0-9._]+(?<![_.])$/;
    const nameRegex: RegExp = /^(?=.{1,35}$)[A-Za-z]+(?:[' -][A-Za-z]+)*$/;
    const addressRegex: RegExp = /^[A-Za-z0-9](?!.*['\.\-\s\,]$)[A-Za-z0-9'\.\-\s\,]{0,68}[A-Za-z0-9]$/;

    if (!usernameRegex.test(this.form.value.username)) {
      this.toastr.warning(' Alphanumeric characters, underscore and dot are allowed. Min. length 3, max. length 20.' +
      ' Special characters can`t be next to each other',
        'Invalid username');
        return;
    }

    if (!nameRegex.test(this.form.value.firstName) || !nameRegex.test(this.form.value.lastName)) {
      this.toastr.warning('First Name and/or Last Name are not valid inputs', 'Invalid personal detail');
      return;
    }

    if (!addressRegex.test(this.form.value.address)) {
      this.toastr.warning('Address format is not valid', 'Invalid address');
      return;
    }

    if (this.form.value.email == "" || this.form.get('email')?.hasError('email')) {
      this.toastr.warning('Email is not valid', 'Invalid email');
      return;
    }

    if (this.userProfile) {
      // Update userProfile with form values
      this.userProfile.username = this.form.value.username;
      this.userProfile.firstName = this.form.value.firstName;
      this.userProfile.lastName = this.form.value.lastName;
      this.userProfile.email = this.form.value.email;
      this.userProfile.address = this.form.value.address;

      this.profileService.updateUser(this.username, this.userProfile).subscribe(
        () => {
          if (this.username === this.userProfile?.username) {
            this.toastr.success('Profile updated successfully', 'Profile update');
            this.loadUserProfile();
          } else {
            this.toastr.success('Profile updated successfully. Relogin with new username',
             'Profile update');
            this.authService.logout();
          }
        },
        (error) => {
          if (error.status === 503) {
            this.toastr.error('Unable to contact auth service. Please try again later',
             'Auth service offline');
          } else {
            this.toastr.error('There was an error while updating your profile', 'Update failed');
          }
          console.error('Error updating user profile: ', error);
        }
      );
    } else {
      console.error('User profile is undefined. Cannot update.');
    }
  }

  deleteProfile() {
    const username: string = this.authService.getUsernameFromToken() || "";

    const hasActiveReservations = this.reservations?.some((reservation: ReservationByAvailablePeriod) => {
      return new Date(reservation.EndDate).getTime() > new Date().getTime();
    });

    if (hasActiveReservations) {
      this.toastr.warning('Cannot delete user profile while you have active reservations', 'Unable to delete');
      return;
    }

    this.profileService.deleteUser(username).subscribe(
      (response) => {
        this.toastr.success('Profile deleted successfully', 'Profile delete');
        this.authService.logout();
      },
      (error) => {
        if (error.error.StatusCode === 400) {
          this.toastr.error('There are active reservations', 'Unable to delete profile');
          return;
        }
        if (error.status === 503) {
          this.toastr.error('Unable to contact auth service. Please try again later',
           'Auth service offline');
        } else {
          this.toastr.error('There was an error while deleting your profile', 'Delete failed');
        }
        console.error('Error deleting user profile: ', error);
      });
  }
}
