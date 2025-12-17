import { Component,OnInit } from '@angular/core';
import { FoodData } from 'src/app/models/food.model';
import { FoodService } from 'src/app/services/food.service';
import { Router } from '@angular/router';
import { environment } from 'src/app/environments/environment';

@Component({
  selector: 'app-food-list',
  templateUrl: './food-list.component.html',
  styleUrls: ['./food-list.component.css']
})
export class FoodListComponent implements OnInit {

  foods: FoodData[] = []; // Lista hrane

  constructor(private foodService: FoodService,private router: Router) {}

  ngOnInit(): void {
    this.loadFoods(); // Učitava hranu kada se komponenta inicijalizuje
  }

  loadFoods(): void {
    this.foodService.getAllFoods().subscribe(
      (data: FoodData[]) => {
        this.foods = data; // Popunjava listu hrane
      },
      error => {
        console.error('Greška prilikom preuzimanja hrane:', error);
      }
    );
  }
  deleteFood(foodId: string): void {
    this.foodService.deleteFood(foodId).subscribe(
      () => {
        // Uklanja obrisanu hranu iz liste
        this.foods = this.foods.filter(food => food.id !== foodId);
        console.log('Food deleted successfully');
      },
      error => {
        console.error('Greška prilikom brisanja hrane:', error);
      }
    );
  }
  goToUpdate(foodId: string): void {
    // Navigiraj na '/update-food/<foodId>'
    this.router.navigate(['/update-food', foodId]);
  }
 
getImageSrc(food: FoodData): string {
  // ako backend vrati imagePath npr "/uploads/<id>.jpg"
  if (food.imagePath) {
    // baseApiUrl ti je tipa http://localhost:8000, pa dodaj /food (gateway route) + imagePath
    return `${environment.baseApiUrl}/food${food.imagePath}`;
  }
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

}
