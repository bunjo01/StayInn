import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { AuthService } from '../services/auth.service';

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
      newPassword: [null, Validators.pattern('^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$')],
      newPassword1: [null, Validators.pattern('^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$')],
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


    // TODO: Pozvati service za reset lozinke
    this.authService.resetPassword(requestBody).subscribe(
      (result) => {
        this.toastr.success('Successful change password', 'Change password');
        this.router.navigate(['login']);
      },
      (error) => {
        console.error('Error while sending mail: ', error);
      }
    );
  }
}
