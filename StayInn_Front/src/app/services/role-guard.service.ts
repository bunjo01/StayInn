// src/app/auth/role-guard.service.ts
import { Injectable } from '@angular/core';
import { Router, CanActivate, ActivatedRouteSnapshot } from '@angular/router';
import { AuthService } from './auth.service';
import * as decode from 'jwt-decode';
import { JwtPayload } from '../model/user';

@Injectable()
export class RoleGuardService implements CanActivate {
  constructor(public auth: AuthService, public router: Router) {}

  canActivate(route: ActivatedRouteSnapshot): boolean {
    const expectedRole = route.data.expectedRole;
    const token = localStorage.getItem('token');
    
    if (token === null) {
      this.router.navigate(['login']);
      return false;
    }

    const tokenPayload = decode.jwtDecode(token) as JwtPayload;

    if (tokenPayload.role !== expectedRole) {
      this.router.navigate(['notFound']);
      return false;
    }

    return true;
  }
}
