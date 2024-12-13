import { ComponentFixture, TestBed } from '@angular/core/testing';

import { StudentHomepageComponent } from './student-homepage.component';

describe('StudentHomepageComponent', () => {
  let component: StudentHomepageComponent;
  let fixture: ComponentFixture<StudentHomepageComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [StudentHomepageComponent]
    });
    fixture = TestBed.createComponent(StudentHomepageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
