import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from 'src/app/environments/environment';
import { FoodData } from '../models/food.model';
import { OrderData } from '../models/order.model';

export interface ReviewSummary {
  foodId: string;
  avgRating: number;
  ratingCount: number;
  commentCount: number;
  canReview?: boolean; // prisutno kad backend enrich-uje (ako je prosao middleware)
  myRating?: number;   // prisutno kad backend enrich-uje
}

export interface ReviewComment {
  id?: string;
  foodId?: string;
  userId?: string;
  author: string;
  text: string;
  createdAt?: string;
}

@Injectable({
  providedIn: 'root'
})
export class FoodService {

  private url = "food";

  constructor(private http: HttpClient) {}

  // ---------- helpers ----------
  private authHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');
    let headers = new HttpHeaders();
    if (token) headers = headers.set('Authorization', `Bearer ${token}`);
    return headers;
  }

  // ---------- existing ----------
  uploadFoodImage(foodId: string, file: File): Observable<any> {
    const formData = new FormData();
    formData.append('image', file);

    return this.http.post<any>(
      `${environment.baseApiUrl}/${this.url}/food/${foodId}/image`,
      formData
    );
  }

  getAllFoods(): Observable<FoodData[]> {
    return this.http.get<FoodData[]>(`${environment.baseApiUrl}/${this.url}/food`);
  }

  deleteFood(foodId: string): Observable<void> {
    return this.http.delete<void>(`${environment.baseApiUrl}/${this.url}/food/${foodId}`);
  }

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

  acceptOrder(orderId: string): Observable<any> {
    return this.http.put<any>(`${environment.baseApiUrl}/${this.url}/order/${orderId}`, {});
  }

  getMyOrders(userId: string): Observable<OrderData[]> {
    let params = new HttpParams().set('user_id', userId);
    return this.http.get<OrderData[]>(`${environment.baseApiUrl}/${this.url}/my-orders`, { params });
  }

  getRecommendations(userId: string): Observable<FoodData[]> {
  let params = new HttpParams().set('user_id', userId);
  return this.http.get<FoodData[]>(
    `${environment.baseApiUrl}/${this.url}/recommendations`,
    { params }
  );
}


  cancelOrder(orderId: string): Observable<void> {
    return this.http.put<void>(`${environment.baseApiUrl}/${this.url}/order/${orderId}/cancel`, {});
  }

  getAcceptedOrders(): Observable<OrderData[]> {
    return this.http.get<OrderData[]>(`${environment.baseApiUrl}/${this.url}/accepted-orders`);
  }

  getFoodById(id: string): Observable<FoodData> {
    return this.http.get<FoodData>(`${environment.baseApiUrl}/${this.url}/food/${id}`);
  }

  // ---------- REVIEWS ----------
  // GET /food/{id}/reviews/summary
  getFoodReviewSummary(foodId: string): Observable<ReviewSummary> {
    return this.http.get<ReviewSummary>(
      `${environment.baseApiUrl}/${this.url}/food/${foodId}/reviews/summary`,
      { headers: this.authHeaders() }
    );
  }

  // POST /foods/reviews/summaries  body: { foodIds: [...] }
  batchFoodSummaries(foodIds: string[]): Observable<Record<string, ReviewSummary>> {
    return this.http.post<Record<string, ReviewSummary>>(
      `${environment.baseApiUrl}/${this.url}/foods/reviews/summaries`,
      { foodIds }
    );
  }

  // POST /food/{id}/reviews/rating body: { rating: 1..5 }
  setFoodRating(foodId: string, rating: number): Observable<ReviewSummary> {
    return this.http.post<ReviewSummary>(
      `${environment.baseApiUrl}/${this.url}/food/${foodId}/reviews/rating`,
      { rating },
      { headers: this.authHeaders() }
    );
  }

  // GET /food/{id}/reviews/comments
  listFoodComments(foodId: string): Observable<ReviewComment[]> {
    return this.http.get<ReviewComment[]>(
      `${environment.baseApiUrl}/${this.url}/food/${foodId}/reviews/comments`
    );
  }

  // POST /food/{id}/reviews/comments body: { text: "..." }
  addFoodComment(foodId: string, text: string): Observable<any> {
    return this.http.post<any>(
      `${environment.baseApiUrl}/${this.url}/food/${foodId}/reviews/comments`,
      { text },
      { headers: this.authHeaders() }
    );
  }
}
