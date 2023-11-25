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
      username: ['', Validators.pattern('^(?=.{3,20}$)(?![_.])(?!.*[_.]{2})[a-zA-Z0-9._]+(?<![_.])$')],
      password: ['', Validators.pattern('^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$')],
    });
  }
  
  send(): void {
    // TODO check why password validator always fails
    // if (this.form.invalid) {
    //   this.form.markAllAsTouched();
    //   return;
    // }

    if (!this.recaptchaResolved) {
      this.toastr.warning('Please complete the reCAPTCHA verification.', 'reCAPTCHA');
      return;
    }

    this.authService.login(this.form.value).subscribe(
      (result) => {
        this.toastr.success('Login successful', 'Login');
        console.log(result);
        // result = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOiIxNzAwOTQzMzMzIiwidXNlcm5hbWUiOiJidW5qbyJ9.2xVsKyUBl4Fr0ziu2OLTOxo07jvYSrc0_ibcNwA4pHE"
        localStorage.setItem('token', result.token);
        this.router.navigate(['']);
      },
      (error) => {
        console.error('Error during login: ', error);
        this.toastr.error('Login failed. Please check your credentials.', 'Failed login');
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
