import { HttpClient } from '@angular/common/http';
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
    // Dohvatanje svih porudžbina
    getAllOrders(): Observable<OrderData[]> {
      return this.http.get<OrderData[]>(`${environment.baseApiUrl}/${this.url}/order`);
    }
  
   // Nova metoda za ažuriranje Food-a (PUT)
   /*updateFood(id: string, updatedFood: FoodData): Observable<any> {
    return this.http.put<any>(`${environment.baseApiUrl}/${this.url}/${id}`, updatedFood);
  }*/
  // Nova metoda: Dohvatanje hrane po ID-u (potrebno za Update stranicu)
  getFoodById(id: string): Observable<FoodData> {
    return this.http.get<FoodData>(`${environment.baseApiUrl}/${this.url}/${id}`);
  }
}
