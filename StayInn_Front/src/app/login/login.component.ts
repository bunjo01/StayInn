// login.component.ts
import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { ReCaptchaV3Service } from 'ng-recaptcha';
import { environment } from 'src/environments/environment';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {
  form: FormGroup;
  recaptchaSiteKey: string = environment.recaptcha.siteKey;
  recaptchaResolved: boolean = false;

  constructor(
    private authService: AuthService,
    private fb: FormBuilder,
    private router: Router,
    private recaptchaV3Service: ReCaptchaV3Service,
    private toastr: ToastrService
  ) {
    this.form = this.fb.group({
      username: ['', Validators.required],
      password: ['', Validators.required],
    });
  }
  
  send(): void {
    const usernameRegex: RegExp = /^(?=.{3,20}$)(?![_.])(?!.*[_.]{2})[a-zA-Z0-9._]+(?<![_.])$/;
    const passwordRegex: RegExp = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$/;

    if (!this.recaptchaResolved) {
      this.toastr.warning('Please complete the reCAPTCHA verification.', 'reCAPTCHA');
      return;
    }

    if (!usernameRegex.test(this.form.value.username)) {
      this.toastr.warning('Input for username is not valid', 'Invalid username');
      return;
    }

    if (!passwordRegex.test(this.form.value.password)) {
      this.toastr.warning('Input for password is not valid', 'Invalid password');
      return;
    }

    this.authService.login(this.form.value).subscribe(
      (result) => {
        this.toastr.success('Login successful', 'Login');
        localStorage.setItem('token', result.token);
        this.router.navigate(['']);
      },
      (error) => {
        console.error('Error during login: ', error);
        this.toastr.error('Login failed. Please check your credentials or activate your account.', 'Failed login');
      }
    );
  }

  solveCaptcha(): void {
    this.recaptchaV3Service.execute('importantAction').subscribe((token: string) => {
      console.debug(`reCAPTCHA resolved with token: [${token}]`);
      this.recaptchaResolved = true;
      this.send();
    });
  }
}
