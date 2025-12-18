export interface ReviewSummary {
  foodId: string;
  avgRating: number;
  ratingCount: number;
  commentCount: number;
  canReview?: boolean;
  myRating?: number;
}

export interface ReviewComment {
  id?: string;
  foodId?: string;
  userId?: string;
  author: string;
  text: string;
  createdAt?: string;
}
