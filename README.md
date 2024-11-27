# Accommodation Offer and Reservation Platform

This project aims to implement a platform for offering and reserving accommodations, with a microservice architecture. The application supports different user roles, such as unauthenticated users, hosts, and guests, along with a variety of functionalities for managing accommodations, reservations, and users.

---

## **System Roles**

- **Unauthenticated User (UA)**  
  - Creates a new account (host or guest)
  - Logs into an existing account
  - Searches for accommodations (cannot make reservations)

- **Host (H)**  
  - Creates and manages accommodations
  - Defines availability, prices, and amenities
  - Can search accommodations but cannot reserve them

- **Guest (G)**  
  - Makes reservations
  - Cancels a reservation before the start date
  - Rates accommodations and hosts

---

## **System Components**

1. **Client Application**  
   - Graphical interface for users

2. **Server Application**  
   - **Auth** - User credentials and registration/login
   - **Profile** - Basic user information
   - **Accommodations** - Accommodation information
   - **Reservations** - Availability, prices, and reservations
   - **Notifications** - Notifications for users

---

## **Main Functionalities**

1. **Registration and Login**  
   - User authentication (username, password, basic data)

2. **Account Management**  
   - Ability to modify personal data and delete the account under certain conditions

3. **Creating and Managing Accommodations**  
   - Host can create accommodations, upload pictures, and define availability and prices

4. **Reservations**  
   - Guests can make reservations, and reservations can be canceled before the start date

5. **Ratings**  
   - Guests can rate accommodations and hosts based on previous bookings

6. **Filtering and Searching for Accommodations**  
   - Accommodations are searched by location, number of guests, date, price, and amenities

7. **Notifications**  
   - Notifications for hosts

---

## **Technical Characteristics of the System**

- **Microservice Architecture** with clearly separated services
- **API Gateway** for communication between the client and server applications (REST API)
- **Containerization** of all services and databases using **Docker Compose**
- **Fault Tolerance** for partial system failures
- **Image Caching** of accommodations in Redis

---

## **Security Requirements**

- **HTTPS Communication** between services and the client application
- **Authentication and Access Control** for users (RBAC model)
- **Data Protection** - encryption and hashing of sensitive data
- **Logging

---
