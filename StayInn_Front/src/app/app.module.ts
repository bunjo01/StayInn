import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ToastrModule } from 'ngx-toastr';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { EntryComponent } from './entry/entry.component';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
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
import { TimestampInterceptor } from './interceptors/timestamp.interceptor';
import { JwtHelperService, JWT_OPTIONS } from '@auth0/angular-jwt';
import { AuthGuardService } from './services/auth-guard.service';
import { RoleGuardService } from './services/role-guard.service';
import { UnauthorizedComponent } from './unauthorized/unauthorized.component';
import { ProfileDetailsComponent } from './profile-details/profile-details.component';
import { ChangePasswordComponent } from './change-password/change-password.component';
import { ResetPasswordComponent } from './reset-password/reset-password.component';
import { ForgetPasswordComponent } from './forget-password/forget-password.component';
import { CreateAccommodationComponent } from './create-accommodation/create-accommodation.component';
import { EditPeriodTemplateComponent } from './reservations/edit-period-template/edit-period-template.component';
import { AccommodationDetailsComponent } from './accommodation-details/accommodation-details.component';
import { EditAccommodationComponent } from './edit-accommodation/edit-accommodation.component';
import { HistoryReservationComponent } from './history-reservation/history-reservation.component';
import { RateAccommodationComponent } from './ratings/rate-accommodation/rate-accommodation.component';
import { RateHostComponent } from './ratings/rate-host/rate-host.component';
import { ProfileMenuComponent } from './profile-menu/profile-menu.component';
import { RatingsViewComponent } from './ratings/ratings-view/ratings-view.component';
import { RatingsPopupComponent } from './ratings/ratings-popup/ratings-popup.component';
import { MatDialogModule } from '@angular/material/dialog';

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
    UnauthorizedComponent,
    ChangePasswordComponent,
    ProfileDetailsComponent,
    ResetPasswordComponent,
    ForgetPasswordComponent,
    CreateAccommodationComponent,
    EditPeriodTemplateComponent,
    AccommodationDetailsComponent,
    EditAccommodationComponent,
    HistoryReservationComponent,
    RateAccommodationComponent,
    RateHostComponent,
    ProfileMenuComponent,
    RatingsViewComponent,
    RatingsPopupComponent,
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
    }),
    MatDialogModule
  ],
  providers: [
    {
      provide: RECAPTCHA_V3_SITE_KEY,
      useValue: environment.recaptcha.siteKey,
    },
    DatePipe,
    {
      provide: HTTP_INTERCEPTORS,
      useClass: TimestampInterceptor,
      multi: true,
    },
    JwtHelperService,
    { provide: JWT_OPTIONS, useValue: JWT_OPTIONS },
    AuthGuardService,
    RoleGuardService,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
