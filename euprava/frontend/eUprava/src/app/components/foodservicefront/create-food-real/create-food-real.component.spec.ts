import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateFoodRealComponent } from './create-food-real.component';

describe('CreateFoodRealComponent', () => {
  let component: CreateFoodRealComponent;
  let fixture: ComponentFixture<CreateFoodRealComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [CreateFoodRealComponent]
    });
    fixture = TestBed.createComponent(CreateFoodRealComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
