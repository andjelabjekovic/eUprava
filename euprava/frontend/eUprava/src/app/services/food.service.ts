import { HttpClient,HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { FoodData } from '../models/food.model'; // Assuming you've moved FoodData to a separate model file
import { Observable } from 'rxjs';
import { environment } from 'src/app/environments/environment';
import { OrderData } from '../models/order.model'; 


@Injectable({
  providedIn: 'root'
})
export class FoodService {

  private url = "food";
  constructor(private http: HttpClient) { }

  uploadFoodImage(foodId: string, file: File): Observable<any> {
  const formData = new FormData();
  formData.append('image', file);

  return this.http.post<any>(
    `${environment.baseApiUrl}/${this.url}/food/${foodId}/image`,
    formData
  );
}

  // Fetch all therapies
  getAllFoods(): Observable<FoodData[]> {
    return this.http.get<FoodData[]>(`${environment.baseApiUrl}/${this.url}/food`);
  }
  // Brisanje hrane
  deleteFood(foodId: string): Observable<void> {
    return this.http.delete<void>(`${environment.baseApiUrl}/${this.url}/food/${foodId}`);
  }

  // Nova metoda: Kreiranje porudžbine
  createOrder(orderData: OrderData, userId: string): Observable<any> {
    return this.http.post<any>(`${environment.baseApiUrl}/${this.url}/order?userId=${userId}`, orderData);
  }
   
  createFood(foodData: FoodData, userId: string): Observable<any> {
    return this.http.post<any>(`${environment.baseApiUrl}/${this.url}/food?cookId=${userId}`, foodData);
  }
  updateFood(id: string, updatedFood: FoodData): Observable<any> {
    return this.http.put<any>(`${environment.baseApiUrl}/${this.url}/food/${id}`, updatedFood);
  }
    getAllOrders(): Observable<OrderData[]> {
      return this.http.get<OrderData[]>(`${environment.baseApiUrl}/${this.url}/order`);
    }

   // Prihvatanje porudžbine (ažuriranje statusa na 'Prihvacena')
   acceptOrder(orderId: string): Observable<any> {
    return this.http.put<any>(`${environment.baseApiUrl}/${this.url}/order/${orderId}`, {});
  }  

  getMyOrders(userId: string): Observable<OrderData[]> {
    let params = new HttpParams().set('user_id', userId);
  
    return this.http.get<OrderData[]>(`${environment.baseApiUrl}/${this.url}/my-orders`, { params });
  }
  
  cancelOrder(orderId: string): Observable<void> {
    return this.http.put<void>(`${environment.baseApiUrl}/${this.url}/order/${orderId}/cancel`, {});
  }
  
  // Dohvatanje prihvaćenih porudžbina
  getAcceptedOrders(): Observable<OrderData[]> {
    return this.http.get<OrderData[]>(`${environment.baseApiUrl}/${this.url}/accepted-orders`);
  }
 getFoodById(id: string): Observable<FoodData> {
  return this.http.get<FoodData>(
    `${environment.baseApiUrl}/${this.url}/food/${id}`
  );
}

  /*getAllMyOrders(): Observable<OrderData[]> {
    return this.http.get<OrderData[]>(`${environment.baseApiUrl}/${this.url}/my-orders`);
  }*/
}
