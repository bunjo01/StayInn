import { Component } from '@angular/core';
import { FormGroup, FormBuilder, Validators } from '@angular/forms';
import { Route, Router } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { AuthService } from '../services/auth.service';
import { HttpErrorResponse } from '@angular/common/http';

@Component({
  selector: 'app-forget-password',
  templateUrl: './forget-password.component.html',
  styleUrls: ['./forget-password.component.css']
})
export class ForgetPasswordComponent {
  resetPasswordForm!: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private authService: AuthService,
    private toastr: ToastrService,
    private router: Router
  ) { }

  ngOnInit() {
    this.resetPasswordForm = this.formBuilder.group({
      email: [null, [Validators.required, Validators.email]]
    });
  }

  onSubmit() {
    if (this.resetPasswordForm.invalid) {
      return;
    }

    const email = this.resetPasswordForm.value.email;

    const requestBody = {
      email: email
    }

    // TODO: Povezati servis za slanje mejla
    this.authService.sendRecoveryMail(requestBody).subscribe(
      (result) => {
        this.toastr.success('Successful sent recovery mail', 'Send mail');
        this.router.navigate(['']);
      },
      (error) => {
        console.error('Error while sending mail: ', error);
        if (error instanceof HttpErrorResponse) {
          // Prikazivanje statusnog koda i poruke greške kroz ToastrService
          const errorMessage = `${error.error}`;
          this.toastr.error(errorMessage, 'Send mail error');
        } else {
          // Ukoliko greška nije HTTP greška, prikazuje se generička poruka
          this.toastr.error('An unexpected error occurred', 'Change Password Error');
        }
      }
    );
  }
}
