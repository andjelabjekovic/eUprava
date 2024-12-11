import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { FoodData } from '../models/food.model'; // Assuming you've moved FoodData to a separate model file
import { Observable } from 'rxjs';
import { environment } from 'src/app/environments/environment';

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

  // Optionally other CRUD methods can be added here
}
