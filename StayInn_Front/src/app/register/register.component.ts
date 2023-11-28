import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { User } from '../model/user';
import { AuthService } from '../services/auth.service';
import { environment } from 'src/environments/environment';
import { ReCaptchaV3Service } from 'ng-recaptcha';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css']
})
export class RegisterComponent {
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
      username: [null, Validators.required],
      password: [null, Validators.required],
      cPassword: [null, Validators.required],
      firstName: [null, Validators.required],
      lastName: [null, Validators.required],
      email: [null, Validators.email],
      address: [null, Validators.required],
      role: ['', Validators.required]
    });
  }

  submit() {
    const usernameRegex: RegExp = /^(?=.{3,20}$)(?![_.])(?!.*[_.]{2})[a-zA-Z0-9._]+(?<![_.])$/;
    const passwordRegex: RegExp = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$/;
    const nameRegex: RegExp = /^(?=.{1,35}$)[A-Za-z]+(?:[' -][A-Za-z]+)*$/;
    const addressRegex: RegExp = /^[A-Za-z0-9](?!.*['\.\-\s\,]$)[A-Za-z0-9'\.\-\s\,]{0,68}[A-Za-z0-9]$/;

    if (!usernameRegex.test(this.form.value.username)) {
      this.toastr.warning(' Alphanumeric characters, underscore and dot are allowed. Min. length 3, max. length 20.' +
      ' Special characters can`t be next to each other',
       'Invalid username');
       return;
    }

    if (!passwordRegex.test(this.form.value.password)) {
      this.toastr.warning('Minimum eight characters, at least one uppercase letter, one lowercase letter and one number',
      'Invalid password');
      return;
    }

    if (this.form.value.password !== this.form.value.cPassword) {
      this.toastr.warning('Passwords do not match', 'Check passwords');
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

    const user: User = {};

    user.username = this.form.value.username;
    user.password = this.form.value.password;
    user.firstName = this.form.value.firstName;
    user.lastName = this.form.value.lastName;
    user.email = this.form.value.email;
    user.address = this.form.value.address;
    user.role = this.form.value.role;

    this.authService.register(user).subscribe(
      (result) => {
        this.toastr.success('Successful registration', 'Registration');
        this.toastr.info('Activation email sent, check your inbox', 'Account activation');
        this.router.navigate(['login']);
      },
      (error) => {
        console.error('Error while registrating: ', error);
      }
    );
  }

  solveCaptcha(): void {
    this.recaptchaV3Service.execute('importantAction').subscribe((token: string) => {
      console.debug(`reCAPTCHA resolved with token: [${token}]`);
      this.recaptchaResolved = true;
      this.submit();
    });
  }
}
