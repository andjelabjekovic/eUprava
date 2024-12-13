import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { FoodService } from 'src/app/services/food.service';
import { FoodData } from 'src/app/models/food.model'; 


@Component({
  selector: 'app-update-food',
  templateUrl: './update-food.component.html',
  styleUrls: ['./update-food.component.css']
})
export class UpdateFoodComponent implements OnInit {
  updateForm: FormGroup;
  foodId!: string; // ID hrane koju ažuriramo

  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private foodService: FoodService
  ) {
    this.updateForm = this.fb.group({
      foodName: ['', [Validators.required, Validators.maxLength(100)]]
    });
  }

  ngOnInit(): void {
    // Uzmi ID iz rute
    this.foodId = this.route.snapshot.paramMap.get('id') || '';
    if (this.foodId) {
      // Uzmi podatke o hrani po ID
      this.foodService.getFoodById(this.foodId).subscribe(
        (food: FoodData) => {
          // Postavi inicijalnu vrednost forme
          this.updateForm.patchValue({
            foodName: food.foodName
          });
        },
        (error: any) => {
          console.error('Error fetching food:', error);
        }
      );
    }
  }

  onSubmit(): void {
    if (this.updateForm.valid) {
      const updatedFood: FoodData = {
        foodName: this.updateForm.value.foodName,
        stanje2: '' // Nije obavezno menjati, ostavi prazno ili uzmi staru vrednost ako treba
      };

      this.foodService.updateFood(this.foodId, updatedFood).subscribe(
        response => {
          console.log('Food updated successfully:', response);
          this.router.navigate(['/food-list']);
        },
        error => {
          console.error('Error updating food:', error);
        }
      );
    } else {
      Object.keys(this.updateForm.controls).forEach(field => {
        const control = this.updateForm.get(field);
        if (control) {
          control.markAsTouched({ onlySelf: true });
        }
      });
    }
  }

  hasError(controlName: string, errorName: string): boolean {
    return this.updateForm.controls[controlName].hasError(errorName);
  }
}
