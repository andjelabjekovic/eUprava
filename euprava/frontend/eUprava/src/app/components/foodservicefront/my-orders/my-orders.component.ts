import { Component, OnInit } from '@angular/core';
import { OrderData } from 'src/app/models/order.model';
import { FoodService } from 'src/app/services/food.service';
import { AuthService } from 'src/app/services/auth.service'; // ← dodaj ako nisi već

@Component({
  selector: 'app-my-orders',
  templateUrl: './my-orders.component.html',
  styleUrls: ['./my-orders.component.css']
})
export class MyOrdersComponent implements OnInit {

  myOrders: OrderData[] = [];
  isLoading: boolean = false;
  errorMessage: string = '';
   userRole: string = ''; 

  constructor(private foodService: FoodService,private authService: AuthService) { }

 ngOnInit(): void {
    this.userRole = this.authService.getUserType() || '';

    const userId = this.authService.getUserId();
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
