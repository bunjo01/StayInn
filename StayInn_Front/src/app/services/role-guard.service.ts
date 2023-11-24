// src/app/auth/role-guard.service.ts
import { Injectable } from '@angular/core';
import { Router, CanActivate, ActivatedRouteSnapshot } from '@angular/router';
import { AuthService } from './auth.service';
import * as decode from 'jwt-decode';

interface JwtPayload {
  role: string; 
  username: string;
}

@Injectable()
export class RoleGuardService implements CanActivate {
  constructor(public auth: AuthService, public router: Router) {}

  canActivate(route: ActivatedRouteSnapshot): boolean {
    const expectedRole = route.data.expectedRole;
    // const token = localStorage.getItem('token');

    const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOiIxNzAwOTQzMzMzIiwidXNlcm5hbWUiOiJidW5qbyJ9.2xVsKyUBl4Fr0ziu2OLTOxo07jvYSrc0_ibcNwA4pHE'

    console.log(token)
    
    if (token === null) {
      this.router.navigate(['login']);
      return false;
    }

    console.log("proslo")

    // Use the JwtPayload interface
    const tokenPayload = decode.jwtDecode(token) as JwtPayload;


    tokenPayload.username == 'bunjo'
    if (tokenPayload.username !== expectedRole) {
      this.router.navigate(['login']);
      return false;
    }

    return true;
  }
}
