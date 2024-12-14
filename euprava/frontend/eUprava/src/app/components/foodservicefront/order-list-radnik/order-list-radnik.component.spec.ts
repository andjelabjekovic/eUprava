import { ComponentFixture, TestBed } from '@angular/core/testing';

import { OrderListRadnikComponent } from './order-list-radnik.component';

describe('OrderListRadnikComponent', () => {
  let component: OrderListRadnikComponent;
  let fixture: ComponentFixture<OrderListRadnikComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [OrderListRadnikComponent]
    });
    fixture = TestBed.createComponent(OrderListRadnikComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
