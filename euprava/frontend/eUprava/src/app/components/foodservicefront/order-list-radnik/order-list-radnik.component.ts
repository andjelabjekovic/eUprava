// src/app/components/order-list-radnik/order-list-radnik.component.ts

import { Component, OnInit } from '@angular/core';
import { OrderData } from 'src/app/models/order.model';
import { FoodService } from 'src/app/services/food.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-order-list-radnik',
  templateUrl: './order-list-radnik.component.html',
  styleUrls: ['./order-list-radnik.component.css']
})
export class OrderListRadnikComponent implements OnInit {

  orders: OrderData[] = []; // Lista porudžbina
  isLoading: boolean = false; // Indikator učitavanja
  errorMessage: string = ''; // Poruka o grešci

  constructor(private foodService: FoodService,
      private router: Router) { }

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
        this.errorMessage = 'Nema porudžbina.';
        this.isLoading = false;
      }
    );
  }
  acceptOrder(orderId: string): void {
    if (confirm("Da li ste sigurni da želite da prihvatite ovu porudžbinu?")) {
      this.foodService.acceptOrder(orderId).subscribe(
        response => {
          console.log('Porudžbina prihvaćena:', response);
          // Ponovo učitaj porudžbine da vidiš ažurirani status
          this.loadOrders();
          this.router.navigate(['/accepted-orders']);
        },
        error => {
          console.error('Greška prilikom prihvatanja porudžbine:', error);
          this.errorMessage = 'Došlo je do greške prilikom prihvatanja porudžbine.';
        }
      );
    }
  }
}
