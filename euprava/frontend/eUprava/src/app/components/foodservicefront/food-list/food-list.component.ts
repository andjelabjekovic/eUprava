import { Component,OnInit } from '@angular/core';
import { FoodData } from 'src/app/models/food.model';
import { FoodService } from 'src/app/services/food.service';

@Component({
  selector: 'app-food-list',
  templateUrl: './food-list.component.html',
  styleUrls: ['./food-list.component.css']
})
export class FoodListComponent implements OnInit {

  foods: FoodData[] = []; // Lista hrane

  constructor(private foodService: FoodService) {}

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
  
}
