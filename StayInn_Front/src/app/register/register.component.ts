import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { User } from '../model/user';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css']
})
export class RegisterComponent {
  form: FormGroup;

  constructor(
    private authService: AuthService,
    private fb: FormBuilder,
    private router: Router
  ) {
    this.form = this.fb.group({
      username: [null, Validators.required, Validators.min(3)],
      password: [null, Validators.required, Validators.min(6)],
      cPassword: [null, Validators.required],
      firstName: [null, Validators.required, Validators.max(35)],
      lastName: [null, Validators.required, Validators.max(35)],
      email: [null, Validators.email],
      address: [null, Validators.required, Validators.pattern("[A-Za-z0-9'\.\-\s\,]")]
    });
  }

  submit() {
    const user: User = {};

    if (this.form.value.password !== this.form.value.cPassword) {
      console.log(this.form.value.password + ' ' + this.form.value.cPassword);
      
      alert('Passwords do not match');
      return;
    }

    user.username = this.form.value.username;
    user.password = this.form.value.password;
    user.firstName = this.form.value.firstName;
    user.lastName = this.form.value.lastName;
    user.email = this.form.value.email;
    user.address = this.form.value.address;

    this.authService.register(user).subscribe(
      (result) => {
        alert('Successful registration');
        console.log(result);
        this.router.navigate(['login']);
      },
      (error) => {
        console.error('Error while registrating: ', error);
      }
    );
  }
}
