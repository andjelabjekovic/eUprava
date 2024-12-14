import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { FoodService } from 'src/app/services/food.service';  // Importuj FoodService ako je potrebno
import { FoodData } from 'src/app/models/food.model';
import { AuthService } from 'src/app/services/auth.service';

@Component({
  selector: 'app-create-food-real',
  templateUrl: './create-food-real.component.html',
  styleUrls: ['./create-food-real.component.css']
})
export class CreateFoodRealComponent implements OnInit {
  foodForm: FormGroup;

  constructor(
    private fb: FormBuilder, 
    private router: Router, 
    private foodService: FoodService,
    private authService: AuthService
  ) {
    this.foodForm = this.fb.group({
      foodName: ['', [Validators.required, Validators.maxLength(100)]]
    });
  }

  ngOnInit(): void {
  }

  onSubmit(): void {
    if (this.foodForm.valid) {
      const loggedUserId = this.authService.getUserId() || '';
      
      const foodData: FoodData = {
        
        user_id: loggedUserId,
        foodName: this.foodForm.value.foodName,
      
      };

      console.log('Pokušaj kreiranja hrane:', foodData);

      this.foodService.createFood(foodData, loggedUserId).subscribe(
        response => {
          console.log('Hrana uspešno kreirana:', response);
          this.router.navigate(['/food-list']);
        },
        error => {
          console.error('Greška prilikom kreiranja hrane:', error);
        }
      );
    } else {
      // Ako forma nije validna, označi sva polja kao dodirnuta radi prikaza grešaka
      Object.keys(this.foodForm.controls).forEach(field => {
        const control = this.foodForm.get(field);
        if (control) {
          control.markAsTouched({ onlySelf: true });
        }
      });
    }
  }

  hasError(controlName: string, errorName: string) {
    return this.foodForm.controls[controlName]?.hasError(errorName);
  }
}