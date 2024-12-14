// src/app/components/food-list-student/food-list-student.component.ts

import { Component, OnInit } from '@angular/core';
import { FoodData } from 'src/app/models/food.model';
import { FoodService } from 'src/app/services/food.service';
import { AuthService } from 'src/app/services/auth.service';
import { OrderData } from 'src/app/models/order.model'; // Importujte OrderData

@Component({
  selector: 'app-food-list-student',
  templateUrl: './food-list-student.component.html',
  styleUrls: ['./food-list-student.component.css']
})
export class FoodListStudentComponent implements OnInit {

  foods: FoodData[] = []; // Lista hrane

  constructor(
    private foodService: FoodService, 
    private authService: AuthService
  ) {}

  ngOnInit(): void {
    this.loadFoods(); // Učitava hranu kada se komponenta inicijalizuje
  }

  loadFoods(): void {
    this.foodService.getAllFoods().subscribe(
      (data: FoodData[]) => {
        this.foods = data; // Popunjava listu hrane
      },
      error => {
        console.error('Greška prilikom preuzimanja hrane:', error);
      }
    );
  }

  orderFood(foodId: string): void {
    const userId = this.authService.getUserId();
    if (!userId) {
      // Obrada slučaja kada korisnik nije prijavljen
      console.error('Korisnik nije prijavljen.');
      return;
    }

    const orderData: OrderData = {
      food: { id: foodId }
    };

    this.foodService.createOrder(orderData, userId).subscribe(
      response => {
        console.log('Porudžbina uspešno kreirana:', response);
        // Opcionalno: Prikazati poruku o uspehu ili osvežiti listu
        alert('Porudžbina je uspešno kreirana!');
      },
      error => {
        console.error('Greška prilikom kreiranja porudžbine:', error);
        // Opcionalno: Prikazati poruku o grešci
        alert('Došlo je do greške prilikom kreiranja porudžbine.');
      }
    );
  }
}
