import { Component } from '@angular/core';

import { Router } from '@angular/router';
@Component({
  selector: 'app-student-homepage',
  templateUrl: './student-homepage.component.html',
  styleUrls: ['./student-homepage.component.css']
})
export class StudentHomepageComponent {
  constructor(private router: Router) {}

viewOrder() {
  this.router.navigate(['/food-list-student']);
}

}
