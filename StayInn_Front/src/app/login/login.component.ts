// login.component.ts
import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { ReCaptchaV3Service } from 'ng-recaptcha';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {
  form: FormGroup;
  recaptchaSiteKey: string = '6LeTihYpAAAAAAv9D98iix0zlwb9OQt7TmgOswwT';
  recaptchaResolved: boolean = false;

  constructor(
    private authService: AuthService,
    private fb: FormBuilder,
    private router: Router,
    private recaptchaV3Service: ReCaptchaV3Service
  ) {
    this.form = this.fb.group({
      username: ['', Validators.required],
      password: ['', Validators.required],
    });
  }
  
  send(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    if (!this.recaptchaResolved) {
      alert('Please complete the reCAPTCHA verification.');
      return;
    }

    this.authService.login(this.form.value).subscribe(
      (result) => {
        alert('Login successful');
        console.log(result);
      },
      (error) => {
        console.error('Error during login: ', error);
        alert('Login failed. Please check your credentials.');
      }
    );
  }

  onCaptchaResolved(token: string): void {
    console.debug(`reCAPTCHA resolved with token: [${token}]`);
    this.recaptchaResolved = true;
  }
}
