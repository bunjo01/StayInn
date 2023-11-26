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
      username: [null, Validators.pattern('^(?=.{3,20}$)(?![_.])(?!.*[_.]{2})[a-zA-Z0-9._]+(?<![_.])$')],
      password: [null, Validators.pattern('^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$')],
      cPassword: [null, Validators.pattern('^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{8,}$')],
      firstName: [null, Validators.pattern("^(?=.{1,35}$)[A-Za-z]+(?:[' -][A-Za-z]+)*$")],
      lastName: [null, Validators.pattern("^(?=.{1,35}$)[A-Za-z]+(?:[' -][A-Za-z]+)*$")],
      email: [null, Validators.email],
      address: [null, Validators.pattern("^[A-Za-z0-9](?!.*['\.\-\s\,]$)[A-Za-z0-9'\.\-\s\,]{0,68}[A-Za-z0-9]$")],
      role: ['', Validators.required]
    });
  }

  submit() {
    const user: User = {};

    if (this.form.value.password !== this.form.value.cPassword) {
      console.log(this.form.value.password + ' ' + this.form.value.cPassword);
      
      this.toastr.warning('Passwords do not match', 'Check passwords');
      return;
    }

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
        console.log(result);
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
