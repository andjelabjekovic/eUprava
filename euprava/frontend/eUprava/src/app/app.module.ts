import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { HttpClientModule} from '@angular/common/http';
import { CreateAppointmentComponent } from './components/create-appointment/create-appointment.component';
import { AppointmentManagementComponent } from './components/appointment-management/appointment-management.component';
import { SystematicCheckupComponent } from './components/systematic-checkup/systematic-checkup.component';
import { AppointmentListComponent } from './components/appointment-list-component/appointment-list-component.component';
import { CommonModule } from '@angular/common';
import { AppointmentUpdateComponent } from './components/appointment-update/appointment-update.component';
import { AppointmentListUpdateComponent } from './components/appointment-list-update/appointment-list-update.component';
import { CreateTherapyComponent } from './components/create-therapy/create-therapy.component';
import { StudentAppointmentManagementComponent } from './components/student-appointment-management/student-appointment-management.component';
import { StudentAppointmentListComponent } from './components/student-appointment-list/student-appointment-list.component';
import { StudentCancelAppointmentListComponent } from './components/student-cancel-appointment-list/student-cancel-appointment-list.component';
import { NavbarComponent } from './components/navbar/navbar.component';

import { CreateFoodComponent } from './components/foodservicefront/therapy-list/create-food.component';

import {LoginComponent} from "./components/login/login.component";
import {RegisterComponent} from "./components/register/register.component";

import { UpdateFoodComponent } from './components/foodservicefront/update-food/update-food.component';
import { HomeRadnikComponent } from './components/foodservicefront/home-radnik/home-radnik.component';


import { HomepageComponent } from './homepage/homepage.component';
import { FoodListComponent } from './components/foodservicefront/food-list/food-list.component';
import { FoodListStudentComponent } from './components/foodservicefront/food-list-student/food-list-student.component';
import { StudentHomepageComponent } from './components/foodservicefront/student-homepage/student-homepage.component';
import { CreateFoodRealComponent } from './components/foodservicefront/create-food-real/create-food-real.component';
import { OrderListRadnikComponent } from './components/foodservicefront/order-list-radnik/order-list-radnik.component';
import { AcceptedOrdersComponent } from './components/foodservicefront/accepted-orders/accepted-orders.component';
import { MyOrdersComponent } from './components/foodservicefront/my-orders/my-orders.component';
import { FoodDetailsComponent } from './components/foodservicefront/food-details/food-details.component';


@NgModule({
  declarations: [
    AppComponent,
    CreateAppointmentComponent,
    AppointmentManagementComponent,
    SystematicCheckupComponent,
    AppointmentListComponent,
    AppointmentUpdateComponent,
    AppointmentListUpdateComponent,
    CreateTherapyComponent,
    StudentAppointmentManagementComponent,
    StudentAppointmentListComponent,
    StudentCancelAppointmentListComponent,

    CreateFoodComponent,
    UpdateFoodComponent,
    HomeRadnikComponent,
   


    HomepageComponent,
    FoodListComponent,
    FoodListStudentComponent,
    StudentHomepageComponent,
    CreateFoodRealComponent,
    OrderListRadnikComponent,
    AcceptedOrdersComponent,
    MyOrdersComponent,
    FoodDetailsComponent,

  ],
  imports: [
    BrowserModule,
    HttpClientModule,
    AppRoutingModule,
    FormsModule,
    ReactiveFormsModule,
    CommonModule,
    NavbarComponent,
    LoginComponent,
    RegisterComponent
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
