import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

import { FoodData } from 'src/app/models/food.model';
import { OrderData } from 'src/app/models/order.model';
import { AuthService } from 'src/app/services/auth.service';
import { FoodService, ReviewSummary } from 'src/app/services/food.service';
import { environment } from 'src/app/environments/environment';

@Component({
  selector: 'app-food-list-student',
  templateUrl: './food-list-student.component.html',
  styleUrls: ['./food-list-student.component.css']
})
export class FoodListStudentComponent implements OnInit {

  foods: FoodData[] = [];

  // avg/counts za sve
  summariesAvg: Record<string, ReviewSummary> = {};

  // canReview + myRating (po studentu) — dolazi iz GET summary uz auth
  summariesUser: Record<string, ReviewSummary> = {};

  // hover stanje po foodId
  hoverByFood: Record<string, number> = {};

  constructor(
    private foodService: FoodService,
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadFoods();
  }

  loadFoods(): void {
    this.foodService.getAllFoods().subscribe(
      (data: FoodData[]) => {
        this.foods = data;
        this.loadAvgBatch();
        this.loadUserSummaries(); // da dobijemo canReview + myRating
      },
      error => console.error('Greška prilikom preuzimanja hrane:', error)
    );
  }

  private loadAvgBatch(): void {
    const ids = this.foods.map(f => f.id!).filter(Boolean);
    if (ids.length === 0) return;

    this.foodService.batchFoodSummaries(ids).subscribe({
      next: (map) => this.summariesAvg = map || {},
      error: () => this.summariesAvg = {}
    });
  }

  private loadUserSummaries(): void {
    // ovde moramo per-food, jer batch endpoint ne vraća canReview/myRating
    const ids = this.foods.map(f => f.id!).filter(Boolean);
    if (ids.length === 0) return;

    const calls = ids.map(id =>
      this.foodService.getFoodReviewSummary(id).pipe(
        catchError(() => of(null as any))
      )
    );

    forkJoin(calls).subscribe((results) => {
      const map: Record<string, ReviewSummary> = {};
      results.forEach((s) => {
        if (!s || !s.foodId) return;
        // foodId sa backa može biti hex ili object-> ali mi mapiramo po url param id,
        // zato uzimamo key iz s.foodId ako je string, a ako nije — preskoči
        // (u praksi ti je string)
        map[s.foodId] = s;
      });

      // Ako backend vrati foodId kao ObjectId JSON (nije slučaj kod tebe),
      // fallback: pozovi još jednom mapiranje preko ids:
      ids.forEach((id) => {
        if (!map[id]) {
          // ako nije stiglo — set default
          map[id] = { foodId: id, avgRating: 0, ratingCount: 0, commentCount: 0, canReview: false, myRating: 0 };
        }
      });

      this.summariesUser = map;
    });
  }

  orderFood(foodId: string): void {
    const userId = this.authService.getUserId();
    if (!userId) {
      alert('Korisnik nije prijavljen.');
      return;
    }

    const orderData: OrderData = { food: { id: foodId } };

    this.foodService.createOrder(orderData, userId).subscribe(
      () => {
        alert('Porudžbina je uspešno kreirana!');
        this.router.navigate(['/my-orders']);
        // posle porudžbine, može da postane canReview=true:
        this.foodService.getFoodReviewSummary(foodId).subscribe({
          next: (s) => {
            if (s?.foodId) this.summariesUser[s.foodId] = s;
          }
        });
      },
      () => alert('Došlo je do greške prilikom kreiranja porudžbine.')
    );
  }

  getImageSrc(food: FoodData): string {
    if (food.imagePath) return `${environment.baseApiUrl}/food${food.imagePath}`;
    return 'assets/no-image.png';
  }

  openDetails(foodId: string): void {
    this.router.navigate(['/food', foodId]);
  }

  mapType1(type1?: string): string {
    switch (type1) {
      case 'PASTA': return 'Pasta';
      case 'PICA': return 'Pica';
      case 'SALATA': return 'Salata';
      default: return '-';
    }
  }

  mapType2(type2?: string): string {
    switch (type2) {
      case 'POSNO': return 'Posno';
      case 'MRSNO': return 'Mrsno';
      default: return '-';
    }
  }

  // ----- stars helpers -----
  getStarFillPercent(avg: number, starIndexZeroBased: number): number {
    const starNumber = starIndexZeroBased + 1;
    const diff = avg - (starNumber - 1);
    if (diff <= 0) return 0;
    if (diff >= 1) return 100;
    return Math.round(diff * 100);
  }

  setHover(foodId: string, value: number): void {
    this.hoverByFood[foodId] = value;
  }

  clearHover(foodId: string): void {
    delete this.hoverByFood[foodId];
  }

  private currentUserValue(foodId: string): number {
    const hover = this.hoverByFood[foodId];
    if (hover != null) return hover;

    const my = this.summariesUser[foodId]?.myRating || 0;
    return my;
  }

  getUserStarFillPercent(foodId: string, starIndexZeroBased: number): number {
    const value = this.currentUserValue(foodId);
    const starNumber = starIndexZeroBased + 1;
    return value >= starNumber ? 100 : 0;
  }

  setRating(foodId: string, rating: number): void {
    // optimistic UI
    const prev = this.summariesUser[foodId]?.myRating || 0;
    if (!this.summariesUser[foodId]) {
      this.summariesUser[foodId] = { foodId, avgRating: 0, ratingCount: 0, commentCount: 0, canReview: true, myRating: rating };
    } else {
      this.summariesUser[foodId].myRating = rating;
    }

    this.foodService.setFoodRating(foodId, rating).subscribe({
      next: (summaryFromBackend) => {
        // backend vraća osvežen summary (avg/ratingCount/commentCount + myRating)
        // ažuriraj avg map:
        this.summariesAvg[foodId] = {
          foodId,
          avgRating: summaryFromBackend.avgRating,
          ratingCount: summaryFromBackend.ratingCount,
          commentCount: summaryFromBackend.commentCount
        };

        // i user map:
        this.summariesUser[foodId] = {
          ...this.summariesUser[foodId],
          canReview: true,
          myRating: summaryFromBackend.myRating ?? rating
        };
      },
      error: () => {
        // rollback
        if (this.summariesUser[foodId]) this.summariesUser[foodId].myRating = prev;
        alert('Ne možeš da oceniš (ili nisi poručila ovu hranu).');
      }
    });
  }
}
