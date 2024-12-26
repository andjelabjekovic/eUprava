import { Component,OnInit } from '@angular/core';
import { TherapyData } from 'src/app/models/appointment.model';
import { TherapyService } from 'src/app/services/therapy.service';


@Component({
  selector: 'app-create-food',
  templateUrl: './create-food.component.html',
  styleUrls: ['./create-food.component.css']
})
export class CreateFoodComponent implements OnInit {

  therapies: TherapyData[] = [];

  constructor(private therapyService: TherapyService) {}

  ngOnInit(): void {
    this.loadTherapies();
  }

  loadTherapies(): void {
    this.therapyService.getAllTherapies().subscribe(
      (data: TherapyData[]) => {
        this.therapies = data;
      },
      error => {
        console.error('Greška prilikom preuzimanja terapija:', error);
      }
    );
  }

  approveTherapy(therapyId: string | undefined): void {
    if (!therapyId) {
      return;
    }
    this.therapyService.approveTherapy(therapyId).subscribe(
      response => {
        console.log('Terapija uspešno odobrena:', response);
        // Nakon uspešnog odobravanja, možete ponovo učitati listu terapija
        this.loadTherapies();
      },
      error => {
        console.error('Greška prilikom odobravanja terapije:', error);
      }
    );
  }

}
