import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { HttpErrorResponse } from '@angular/common/http';

@Component({
  selector: 'app-change-password',
  templateUrl: './change-password.component.html',
  styleUrls: ['./change-password.component.css']
})
export class ChangePasswordComponent implements OnInit{
  changePasswordForm!: FormGroup;
  errorMessage!: string;
  

  constructor(
    private formBuilder: FormBuilder,
    private router: Router,
    private authService: AuthService,
    private toastr: ToastrService
  ) { }

  ngOnInit() {
    this.changePasswordForm = this.formBuilder.group({
      currentPassword: [null, Validators.required],
      newPassword: [null, Validators.pattern(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$/)],
      newPassword1: [null, Validators.pattern(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$/)]
    });
  }

  onSubmit() {
    const currentPassword = this.changePasswordForm.value.currentPassword;
    const newPassword = this.changePasswordForm.value.newPassword;
    const newPassword1 = this.changePasswordForm.value.newPassword1;
    const username = this.authService.getUsernameFromToken();
    
    if (newPassword !== newPassword1) {
      this.changePasswordForm.setErrors({ 'passwordMismatch': true });
      this.toastr.error("New passwords do not match.", 'Error');
      return;
    }

    const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$/;
    if (!passwordRegex.test(newPassword)) {
      this.toastr.error("New password must contain at least 8 characters, including one uppercase letter, one lowercase letter, and one number.", 'Error');
      return;
    }

    const requestBody = {
      username: username,
      currentPassword: currentPassword,
      newPassword: newPassword,
    }


    this.authService.changePassword(requestBody).subscribe(
      (result) => {
        this.toastr.success('Successful change password', 'Change password');
        console.log(result);
        this.router.navigate(['login']);
      },
      (error) => {
        console.error('Error while changing password: ', error);
        if (error instanceof HttpErrorResponse) {
          const errorMessage = `${error.error}`;
          this.toastr.error(errorMessage, 'Change Password Error');
        } else {
          this.toastr.error('An unexpected error occurred', 'Change Password Error');
        }
      }
    );
  }
}