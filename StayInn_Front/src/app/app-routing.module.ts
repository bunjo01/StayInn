import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { EntryComponent } from './entry/entry.component';
import { LoginComponent } from './login/login.component';
import { RegisterComponent } from './register/register.component';
import { AddAvailablePeriodTemplateComponent } from './reservations/add-available-period-template/add-available-period-template.component';

const routes: Routes = [
  { path: '', component: EntryComponent },
  { path: 'login', component: LoginComponent},
  { path: 'register', component: RegisterComponent},
  { path: 'addAvailablePeriod', component: AddAvailablePeriodTemplateComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
