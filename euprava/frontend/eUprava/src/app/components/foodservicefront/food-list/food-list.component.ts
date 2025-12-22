import { Component, OnInit } from '@angular/core';
import { FoodData } from 'src/app/models/food.model';
import { FoodService, ReviewSummary } from 'src/app/services/food.service';
import { Router } from '@angular/router';
import { environment } from 'src/app/environments/environment';

@Component({
  selector: 'app-food-list',
  templateUrl: './food-list.component.html',
  styleUrls: ['./food-list.component.css']
})
export class FoodListComponent implements OnInit {

  foods: FoodData[] = [];
  displayFoods: FoodData[] = []; // ✅ NOVO: lista za prikaz (sortirana)
  summaries: Record<string, ReviewSummary> = {};

  constructor(private foodService: FoodService, private router: Router) {}

  ngOnInit(): void {
    this.loadFoods();
  }

  loadFoods(): void {
    this.foodService.getAllFoods().subscribe(
      (data: FoodData[]) => {
        this.foods = data;
        this.displayFoods = [...this.foods]; // privremeno, dok ne stignu summaries
        this.loadSummariesBatch();
      },
      error => console.error('Greška prilikom preuzimanja hrane:', error)
    );
  }

  private loadSummariesBatch(): void {
    const ids = this.foods.map(f => f.id!).filter(Boolean);
    if (ids.length === 0) {
      this.summaries = {};
      this.applyCookSort();
      return;
    }

    this.foodService.batchFoodSummaries(ids).subscribe({
      next: (map) => {
        this.summaries = map || {};
        this.applyCookSort(); // ✅ NOVO: sortiraj tek kad stignu avg/count
      },
      error: () => {
        this.summaries = {};
        this.applyCookSort();
      }
    });
  }

  // ✅ NOVO: Kuvar vidi prvo bez recenzija, pa najmanji prosek, pa najveći
  private applyCookSort(): void {
    const list = [...this.foods];

    list.sort((a, b) => {
      const sa = a.id ? this.summaries[a.id] : undefined;
      const sb = b.id ? this.summaries[b.id] : undefined;

      const countA = sa?.ratingCount ?? 0;
      const countB = sb?.ratingCount ?? 0;

      // 1) bez recenzija ide prvo
      const noReviewsA = countA === 0;
      const noReviewsB = countB === 0;
      if (noReviewsA && !noReviewsB) return -1;
      if (!noReviewsA && noReviewsB) return 1;

      // 2) oba imaju recenzije => avg rastuće
      const avgA = sa?.avgRating ?? 0;
      const avgB = sb?.avgRating ?? 0;
      if (avgA !== avgB) return avgA - avgB;

      // 3) tie-break: manje ocena prvo, pa po imenu (stabilnije)
      if (countA !== countB) return countA - countB;

      return (a.foodName || '').localeCompare(b.foodName || '');
    });

    this.displayFoods = list;
  }

  deleteFood(foodId: string): void {
    this.foodService.deleteFood(foodId).subscribe(
      () => {
        this.foods = this.foods.filter(food => food.id !== foodId);
        delete this.summaries[foodId];
        this.applyCookSort(); // ✅ NOVO: osveži prikaz posle brisanja
      },
      error => console.error('Greška prilikom brisanja hrane:', error)
    );
  }

  goToUpdate(foodId: string): void {
    this.router.navigate(['/update-food', foodId]);
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

  // half-stars: returns 0 / 50 / 100 (or any %)
  getStarFillPercent(avg: number, starIndexZeroBased: number): number {
    const starNumber = starIndexZeroBased + 1;
    const diff = avg - (starNumber - 1);
    if (diff <= 0) return 0;
    if (diff >= 1) return 100;
    return Math.round(diff * 100);
  }
}
