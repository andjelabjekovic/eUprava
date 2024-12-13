import { Component,OnInit  } from '@angular/core';
import { FoodData } from 'src/app/models/food.model';
import { FoodService } from 'src/app/services/food.service';
@Component({
  selector: 'app-food-list-student',
  templateUrl: './food-list-student.component.html',
  styleUrls: ['./food-list-student.component.css']
})
export class FoodListStudentComponent implements OnInit {

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
  
  Order(foodId: string): void {
    this.foodService.orderFood(foodId).subscribe(
      response => {
        console.log('Uspešno naručeno:', response);
        alert('Hrana je uspešno naručena!');
      },
      error => {
        console.error('Greška prilikom naručivanja hrane:', error);
        alert('Došlo je do greške prilikom naručivanja.');
      }
    );
  }
  
}

