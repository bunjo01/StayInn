import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { EntryComponent } from './entry/entry.component';
import { LoginComponent } from './login/login.component';
import { RegisterComponent } from './register/register.component';
import { AddAvailablePeriodTemplateComponent } from './reservations/add-available-period-template/add-available-period-template.component';
import { AvailablePeriodsComponent } from './reservations/available-periods/available-periods.component';
import { AddReservationComponent } from './reservations/add-resevation/add-reservation.component';
import { ReservationsComponent } from './reservations/reservations/reservations.component';
import { AuthGuardService } from './services/auth-guard.service';
import { RoleGuardService } from './services/role-guard.service';
import { UnauthorizedComponent } from './unauthorized/unauthorized.component';
import { ChangePasswordComponent } from './change-password/change-password.component';
import { ProfileDetailsComponent } from './profile-details/profile-details.component';
import { ForgetPasswordComponent } from './forget-password/forget-password.component';
import { ResetPasswordComponent } from './reset-password/reset-password.component';
import { CreateAccommodationComponent } from './create-accommodation/create-accommodation.component';
import { EditPeriodTemplateComponent } from './reservations/edit-period-template/edit-period-template.component';
import { AccommodationDetailsComponent } from './accommodation-details/accommodation-details.component';
import { EditAccommodationComponent } from './edit-accommodation/edit-accommodation.component';

const routes: Routes = [
  { path: '', component: EntryComponent },
  { path: 'login', component: LoginComponent},
  { path: 'register', component: RegisterComponent},
  { path: 'change-password', component: ChangePasswordComponent},
  { path: 'forget-password', component: ForgetPasswordComponent},
  { path: 'recover-account', component: ResetPasswordComponent},
  { path: 'addAvailablePeriod', component: AddAvailablePeriodTemplateComponent, canActivate : [RoleGuardService], data: { 
    expectedRole: 'HOST'
  } },
  { path: 'availablePeriods', component: AvailablePeriodsComponent, canActivate: [AuthGuardService] },
  { path: 'addReservation', component: AddReservationComponent, canActivate: [AuthGuardService]},
  { path: 'reservations', component: ReservationsComponent, canActivate: [AuthGuardService]},
  { path: 'notFound', component: UnauthorizedComponent},
  { path: 'profile', component: ProfileDetailsComponent},
  { path: 'create-accommodation', component: CreateAccommodationComponent, canActivate : [RoleGuardService], data: { 
    expectedRole: 'HOST' }},
  { path: 'update-accommodation', component: EditAccommodationComponent, canActivate : [RoleGuardService], data: { 
    expectedRole: 'HOST' }},
  { path: 'editAvailablePeriod', component: EditPeriodTemplateComponent, canActivate : [RoleGuardService], data: { 
    expectedRole: 'HOST' }},
  {path: 'accommodation-details', component: AccommodationDetailsComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
