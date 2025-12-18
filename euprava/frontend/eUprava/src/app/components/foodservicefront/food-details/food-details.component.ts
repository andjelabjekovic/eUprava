import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { FoodService, ReviewComment, ReviewSummary } from 'src/app/services/food.service';
import { FoodData } from 'src/app/models/food.model';
import { environment } from 'src/app/environments/environment';
import { AuthService } from 'src/app/services/auth.service';

@Component({
  selector: 'app-food-details',
  templateUrl: './food-details.component.html',
  styleUrls: ['./food-details.component.css']
})
export class FoodDetailsComponent implements OnInit {

  food?: FoodData;
  loading = true;

  summary?: ReviewSummary;
  comments: ReviewComment[] = [];

  commentText = '';
  sendingComment = false;

  userType: string | null = null;

  constructor(
    private route: ActivatedRoute,
    private foodService: FoodService,
    private authService: AuthService
  ) {}

  ngOnInit(): void {
    this.userType = this.authService.getUserType();

    const id = this.route.snapshot.paramMap.get('id');
    if (!id) {
      this.loading = false;
      return;
    }

    this.foodService.getFoodById(id).subscribe({
      next: (data) => {
        this.food = data;
        this.loading = false;

        // Reviews summary + comments
        this.loadSummary(id);
        this.loadComments(id);
      },
      error: () => {
        this.loading = false;
      }
    });
  }

  private loadSummary(foodId: string): void {
    this.foodService.getFoodReviewSummary(foodId).subscribe({
      next: (s) => this.summary = s,
      error: () => {
        // ako auth enrichment ne radi, bar uzmi avg/counts preko non-auth batch fallback:
        this.foodService.batchFoodSummaries([foodId]).subscribe({
          next: (m) => this.summary = m?.[foodId],
          error: () => {}
        });
      }
    });
  }

  private loadComments(foodId: string): void {
    this.foodService.listFoodComments(foodId).subscribe({
      next: (list) => this.comments = list || [],
      error: () => this.comments = []
    });
  }

  get canStudentComment(): boolean {
    // student + canReview (tj poručila)
    return this.userType === 'student' && !!this.summary?.canReview;
  }

  addComment(): void {
    if (!this.food?.id) return;

    const text = (this.commentText || '').trim();
    if (!text) return;

    this.sendingComment = true;
    this.foodService.addFoodComment(this.food.id, text).subscribe({
      next: () => {
        this.commentText = '';
        this.sendingComment = false;
        // reload comments + summary counts
        this.loadComments(this.food!.id!);
        this.loadSummary(this.food!.id!);
      },
      error: () => {
        this.sendingComment = false;
        alert('Ne možeš da ostaviš komentar (moraš biti student i moraš poručiti ovu hranu).');
      }
    });
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
