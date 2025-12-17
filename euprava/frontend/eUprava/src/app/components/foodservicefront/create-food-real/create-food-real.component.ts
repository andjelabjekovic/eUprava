import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { FoodService } from 'src/app/services/food.service';
import { FoodData } from 'src/app/models/food.model';
import { AuthService } from 'src/app/services/auth.service';

@Component({
  selector: 'app-create-food-real',
  templateUrl: './create-food-real.component.html',
  styleUrls: ['./create-food-real.component.css']
})
export class CreateFoodRealComponent implements OnInit {
  foodForm: FormGroup;

  selectedFile: File | null = null;
  imagePreviewUrl: string | null = null;

  isSubmitting = false;

  constructor(
    private fb: FormBuilder,
    private router: Router,
    private foodService: FoodService,
    private authService: AuthService
  ) {
    this.foodForm = this.fb.group({
      foodName: ['', [Validators.required, Validators.maxLength(100)]],
      type1: ['', [Validators.required]],
      type2: ['', [Validators.required]],
    });
  }

  ngOnInit(): void {}

  onFileSelected(event: any): void {
    const file: File | null = event?.target?.files?.[0] || null;
    if (!file) return;

    this.selectedFile = file;

    // preview
    const reader = new FileReader();
    reader.onload = () => {
      this.imagePreviewUrl = reader.result as string;
    };
    reader.readAsDataURL(file);
  }

  removeSelectedImage(): void {
    this.selectedFile = null;
    this.imagePreviewUrl = null;
  }

  onSubmit(): void {
    if (this.foodForm.invalid || this.isSubmitting) {
      Object.keys(this.foodForm.controls).forEach(field => {
        const control = this.foodForm.get(field);
        control?.markAsTouched({ onlySelf: true });
      });
      return;
    }

    this.isSubmitting = true;

    const loggedUserId = this.authService.getUserId() || '';

    const foodData: FoodData = {
      user_id: loggedUserId,
      foodName: this.foodForm.value.foodName,
      type1: this.foodForm.value.type1,
      type2: this.foodForm.value.type2,
    };

    this.foodService.createFood(foodData, loggedUserId).subscribe(
      (createdFood: any) => {
        const createdFoodId: string | undefined = createdFood?.id;

        // ako nema slike ili nema id -> samo idi na listu
        if (!this.selectedFile || !createdFoodId) {
          this.isSubmitting = false;
          this.router.navigate(['/food-list']);
          return;
        }

        // upload slike posle kreiranja
        this.foodService.uploadFoodImage(createdFoodId, this.selectedFile).subscribe(
          () => {
            this.isSubmitting = false;
            this.router.navigate(['/food-list']);
          },
          (err) => {
            console.error('Greška upload slike:', err);
            this.isSubmitting = false;
            // Hrana je kreirana, slika nije — i dalje idemo na listu
            this.router.navigate(['/food-list']);
          }
        );
      },
      error => {
        console.error('Greška prilikom kreiranja hrane:', error);
        this.isSubmitting = false;
      }
    );
  }

  hasError(controlName: string, errorName: string) {
    return this.foodForm.controls[controlName]?.hasError(errorName);
  }
}
