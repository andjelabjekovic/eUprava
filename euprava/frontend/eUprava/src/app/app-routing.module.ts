import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { CreateAppointmentComponent } from './components/create-appointment/create-appointment.component';
import { AppointmentManagementComponent } from './components/appointment-management/appointment-management.component';
import { SystematicCheckupComponent } from './components/systematic-checkup/systematic-checkup.component';
import { AppointmentListComponent } from './components/appointment-list-component/appointment-list-component.component';
import { AppointmentUpdateComponent } from './components/appointment-update/appointment-update.component';
import { AppointmentListUpdateComponent } from './components/appointment-list-update/appointment-list-update.component';
import { CreateTherapyComponent } from './components/create-therapy/create-therapy.component';
import { StudentAppointmentManagementComponent } from './components/student-appointment-management/student-appointment-management.component';
import { StudentAppointmentListComponent } from './components/student-appointment-list/student-appointment-list.component';
import { StudentCancelAppointmentListComponent } from './components/student-cancel-appointment-list/student-cancel-appointment-list.component';
import {LoginComponent} from "./components/login/login.component";
import {RegisterComponent} from "./components/register/register.component";

import { HomeRadnikComponent } from './components/foodservicefront/home-radnik/home-radnik.component'; // Apsolutna putanja do komponente
import { CreateFoodComponent } from './components/foodservicefront/therapy-list/create-food.component';
import {HomepageComponent} from "./homepage/homepage.component";
import { FoodListComponent } from './components/foodservicefront/food-list/food-list.component'; // Ispravite putanju prema vašem projektu
import { StudentHomepageComponent } from './components/foodservicefront/student-homepage/student-homepage.component';  // Importuj StudentHomepageComponent
import { FoodListStudentComponent } from './components/foodservicefront/food-list-student/food-list-student.component';  // Importuj FoodListStudentComponent
import { CreateFoodRealComponent } from './components/foodservicefront/create-food-real/create-food-real.component'; // Dodaj CreateFoodRealComponent
import { UpdateFoodComponent } from './components/foodservicefront/update-food/update-food.component'; // Dodaj UpdateFoodComponent
import { OrderListRadnikComponent } from './components/foodservicefront/order-list-radnik/order-list-radnik.component';
import { AcceptedOrdersComponent } from './components/foodservicefront/accepted-orders/accepted-orders.component';

const routes: Routes = [
  {path:'accepted-orders',component:AcceptedOrdersComponent},
  { path: 'order-list-radnik',component: OrderListRadnikComponent },
  { path: 'student-homepage', component: StudentHomepageComponent },  
  { path: 'food-list-student', component: FoodListStudentComponent },
  { path: 'food-list', component: FoodListComponent },
  { path: 'therapy-list', component: CreateFoodComponent },
  { path: 'home-radnik', component: HomeRadnikComponent },
  { path: 'create-food-real', component: CreateFoodRealComponent },  // Ruta za CreateFoodRealComponent
  { path: 'update-food/:id', component: UpdateFoodComponent },  // Ruta za UpdateFoodComponent sa parametrom
  
  { path: 'create-appointment', component: CreateAppointmentComponent },
  { path: 'student-appointment-management', component: StudentAppointmentManagementComponent },
  { path: 'student-appointment-list', component: StudentAppointmentListComponent },
  { path: 'create-systematicCheck', component: SystematicCheckupComponent },
  { path: 'appointment-list', component: AppointmentListComponent },
  { path: 'update-appointment/:id', component: AppointmentUpdateComponent },
  { path: 'update-appointment-list', component: AppointmentListUpdateComponent },
  { path: 'create-therapy', component: CreateTherapyComponent },
  { path: 'cancel-appointment', component: StudentCancelAppointmentListComponent },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'appointment-management', component: AppointmentManagementComponent },
  { path: '', redirectTo: '/login', pathMatch: 'full' },
  { path: 'homepage', component: HomepageComponent},
];


@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
