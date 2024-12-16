import { Component, OnInit } from '@angular/core';
import { OrderData } from 'src/app/models/order.model';
import { FoodService } from 'src/app/services/food.service';

@Component({
  selector: 'app-my-orders',
  templateUrl: './my-orders.component.html',
  styleUrls: ['./my-orders.component.css']
})
export class MyOrdersComponent implements OnInit {

  myOrders: OrderData[] = [];
  isLoading: boolean = false;
  errorMessage: string = '';

  constructor(private foodService: FoodService) { }

  ngOnInit(): void {
    const userId = localStorage.getItem('user_id'); // Pretpostavljamo da ID korisnika postoji u LocalStorage
    if (userId) {
      this.loadMyOrders(userId);
    } else {
      this.errorMessage = 'Niste ulogovani.';
    }
  }

  loadMyOrders(userId: string): void {
    this.isLoading = true;
    this.foodService.getMyOrders(userId).subscribe(
      (data: OrderData[]) => {
        this.myOrders = data;
        this.isLoading = false;
      },
      error => {
        console.error('Greška prilikom preuzimanja porudžbina:', error);
        this.errorMessage = 'Jos uvek nemam porudžbine.';
        this.isLoading = false;
      }
    );
  }
  
  cancelOrder(orderId: string): void {
    this.foodService.cancelOrder(orderId).subscribe(
      () => {
        // Ažuriraj prikaz nakon uspešnog otkazivanja
        this.myOrders = this.myOrders.filter(order => order.id !== orderId);
      },
      (error) => {
        console.error('Greška prilikom otkazivanja porudžbine:', error);
      }
    );
  }
}
