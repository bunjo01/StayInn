import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
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
    AppRoutingModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    RecaptchaV3Module,
  ],
  providers: [
    {
      provide: RECAPTCHA_V3_SITE_KEY,
      useValue: '6LeTihYpAAAAAAv9D98iix0zlwb9OQt7TmgOswwT',
    },
    DatePipe
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
