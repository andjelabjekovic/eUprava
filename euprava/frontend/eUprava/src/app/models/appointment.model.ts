export interface Appointment {
    student_id: string;
    doctor_id: string;
    date: Date;
    door_number: number;
    description: string;
    systematic: boolean;
    faculty_name:string;
    field_of_study:string;
    reserved:boolean;
  }

export interface TherapyData {
  id?: string;
  status: string;
  diagnosis: string;
}


