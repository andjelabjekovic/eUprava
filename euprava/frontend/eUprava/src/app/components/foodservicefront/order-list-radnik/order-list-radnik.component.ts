// src/app/components/order-list-radnik/order-list-radnik.component.ts

import { Component, OnInit } from '@angular/core';
import { OrderData } from 'src/app/models/order.model';
import { FoodService } from 'src/app/services/food.service';

@Component({
  selector: 'app-order-list-radnik',
  templateUrl: './order-list-radnik.component.html',
  styleUrls: ['./order-list-radnik.component.css']
})
export class OrderListRadnikComponent implements OnInit {

  orders: OrderData[] = []; // Lista porudžbina
  isLoading: boolean = false; // Indikator učitavanja
  errorMessage: string = ''; // Poruka o grešci

  constructor(private foodService: FoodService) { }

  ngOnInit(): void {
    this.loadOrders(); // Učitaj porudžbine prilikom inicijalizacije komponente
  }

  loadOrders(): void {
    this.isLoading = true;
    this.foodService.getAllOrders().subscribe(
      (data: OrderData[]) => {
        this.orders = data;
        this.isLoading = false;
      },
      error => {
        console.error('Greška prilikom preuzimanja porudžbina:', error);
        this.errorMessage = 'Došlo je do greške prilikom preuzimanja porudžbina.';
        this.isLoading = false;
      }
    );
  }
}
