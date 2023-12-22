import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { AuthService } from '../services/auth.service';
import { HttpErrorResponse } from '@angular/common/http';

@Component({
  selector: 'app-reset-password',
  templateUrl: './reset-password.component.html',
  styleUrls: ['./reset-password.component.css']
})
export class ResetPasswordComponent {
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
      newPassword: [null, Validators.required],
      newPassword1: [null, Validators.required],
      recoverUUID: [null, Validators.required],
    });
  }

  onSubmit() {
    const newPassword = this.changePasswordForm.value.newPassword;
    const newPassword1 = this.changePasswordForm.value.newPassword1;
    const recoveryUUID = this.changePasswordForm.value.recoverUUID;
    
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
      newPassword: newPassword,
      recoveryUUID: recoveryUUID,
    }

    this.authService.resetPassword(requestBody).subscribe(
      (result) => {
        this.toastr.success('Successful change password', 'Change password');
        this.router.navigate(['login']);
      },
      (error) => {
        console.error('Error while sending mail: ', error);
        if (error instanceof HttpErrorResponse) {
          const errorMessage = `${error.error}`;
          this.toastr.error(errorMessage, 'Recover Password Error');
        } else {
          this.toastr.error('An unexpected error occurred', 'Recover Password Error');
        }
      }
    );
  }
}
