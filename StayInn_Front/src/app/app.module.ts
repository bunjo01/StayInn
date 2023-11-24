import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ToastrModule } from 'ngx-toastr';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { EntryComponent } from './entry/entry.component';
import { HttpClientModule } from '@angular/common/http';
import { LoginComponent } from './login/login.component';
import { RegisterComponent } from './register/register.component';
import { HeaderComponent } from './header/header.component';
import { AccommodationsComponent } from './accommodations/accommodations.component';
import { FooterComponent } from './footer/footer.component';
import { AddAvailablePeriodTemplateComponent } from './reservations/add-available-period-template/add-available-period-template.component';
import { DatePipe } from '@angular/common';
import { AvailablePeriodsComponent } from './reservations/available-periods/available-periods.component';
import { AddReservationComponent } from './reservations/add-resevation/add-reservation.component';
import { ReservationsComponent } from './reservations/reservations/reservations.component';
import { RECAPTCHA_V3_SITE_KEY, RecaptchaV3Module } from 'ng-recaptcha';
import { environment } from 'src/environments/environment';
import { JwtHelperService, JWT_OPTIONS } from '@auth0/angular-jwt';
import { AuthGuardService } from './services/auth-guard.service';
import { RoleGuardService } from './services/role-guard.service';


@NgModule({
  declarations: [
    AppComponent,
    EntryComponent,
    LoginComponent,
    RegisterComponent,
    HeaderComponent,
    AccommodationsComponent,
    FooterComponent,
    AddAvailablePeriodTemplateComponent,
    AvailablePeriodsComponent,
    AddReservationComponent,
    ReservationsComponent,

  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    AppRoutingModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    RecaptchaV3Module,
    ToastrModule.forRoot({
      closeButton: true,
      timeOut: 4000,
      extendedTimeOut: 500,
      preventDuplicates: true,
      countDuplicates: true,
      resetTimeoutOnDuplicate: true
    })
  ],
  providers: [
    {
      provide: RECAPTCHA_V3_SITE_KEY,
      useValue: environment.recaptcha.siteKey,
    },
    DatePipe,
    JwtHelperService,
    { provide: JWT_OPTIONS, useValue: JWT_OPTIONS },
    AuthGuardService,
    RoleGuardService,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
