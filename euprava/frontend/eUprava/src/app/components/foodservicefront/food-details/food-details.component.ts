import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { FoodService } from 'src/app/services/food.service';
import { FoodData } from 'src/app/models/food.model';
import { environment } from 'src/app/environments/environment';

@Component({
  selector: 'app-food-details',
  templateUrl: './food-details.component.html',
  styleUrls: ['./food-details.component.css']
})
export class FoodDetailsComponent implements OnInit {

  food?: FoodData;
  loading = true;

  constructor(
    private route: ActivatedRoute,
    private foodService: FoodService
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.foodService.getFoodById(id).subscribe({
        next: (data) => {
          this.food = data;
          this.loading = false;
        },
        error: () => {
          this.loading = false;
        }
      });
    }
  }

  getImageSrc(): string {
    if (this.food?.imagePath) {
      return `${environment.baseApiUrl}/food${this.food.imagePath}`;
    }
    return 'assets/no-image.png';
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
}
