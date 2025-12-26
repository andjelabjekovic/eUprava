import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FoodListStudentComponent } from './food-list-student.component';

describe('FoodListStudentComponent', () => {
  let component: FoodListStudentComponent;
  let fixture: ComponentFixture<FoodListStudentComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [FoodListStudentComponent]
    });
    fixture = TestBed.createComponent(FoodListStudentComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
 