import { Component, OnInit } from '@angular/core';
import { OrderData } from 'src/app/models/order.model';
import { FoodService } from 'src/app/services/food.service';

@Component({
  selector: 'app-accepted-orders',
  templateUrl: './accepted-orders.component.html',
  styleUrls: ['./accepted-orders.component.css']
})
export class AcceptedOrdersComponent implements OnInit {

  acceptedOrders: OrderData[] = []; // Lista prihvaćenih porudžbina
  isLoading: boolean = false; // Indikator učitavanja
  errorMessage: string = ''; // Poruka o grešci

  constructor(private foodService: FoodService) { }

  ngOnInit(): void {
    this.loadAcceptedOrders(); // Učitaj prihvaćene porudžbine prilikom inicijalizacije komponente
  }

  // Metoda za učitavanje prihvaćenih porudžbina
  loadAcceptedOrders(): void {
    this.isLoading = true;
    this.foodService.getAcceptedOrders().subscribe(
      (data: OrderData[]) => {
        this.acceptedOrders = data;
        this.isLoading = false;
      },
      error => {
        console.error('Greška prilikom preuzimanja prihvaćenih porudžbina:', error);
        this.errorMessage = 'Došlo je do greške prilikom preuzimanja prihvaćenih porudžbina.';
        this.isLoading = false;
      }
    );
  }

}
