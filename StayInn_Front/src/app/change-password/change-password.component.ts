import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { ToastrService } from 'ngx-toastr';

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
      currentPassword: ['', Validators.required],
      newPassword: ['', Validators.required],
      newPassword1: ['', Validators.required]
    });
  }

  onSubmit() {
    const currentPassword = this.changePasswordForm.value.currentPassword;
    const newPassword = this.changePasswordForm.value.newPassword;
    const newPassword1 = this.changePasswordForm.value.newPassword1;
    const usernameUser = this.authService.getUsernameFromToken();
    
    if (newPassword !== newPassword1) {
      this.toastr.error("New passwords do not match.", 'Error');
      return;
    }

    const requestBody = {
      usernameUser: usernameUser,
      currentPassword: currentPassword,
      newPassword: newPassword,
      newPassword1: newPassword1
    }

    this.authService.changePassword(requestBody).subscribe(
      (result) => {
        this.toastr.success('Successful change password', 'Change password');
        console.log(result);
        this.router.navigate(['login']);
      },
      (error) => {
        console.error('Error while changing password: ', error);
      }
    );
  }
}