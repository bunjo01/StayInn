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

const routes: Routes = [
  { path: '', component: EntryComponent },
  { path: 'login', component: LoginComponent},
  { path: 'register', component: RegisterComponent},
  { path: 'addAvailablePeriod', component: AddAvailablePeriodTemplateComponent, canActivate : [RoleGuardService], data: { 
    expectedRole: 'bunjo'
  } },
  { path: 'availablePeriods', component: AvailablePeriodsComponent},
  { path: 'addReservation', component: AddReservationComponent},
  { path: 'reservations', component: ReservationsComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
