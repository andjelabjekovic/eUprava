import { Component } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { HttpClientModule, HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { AuthService } from 'src/app/services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    ReactiveFormsModule,  // Import ReactiveFormsModule for forms
    HttpClientModule,     // Import HttpClientModule for HTTP requests
    CommonModule          // Import CommonModule for common directives
  ],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {
  loginForm: FormGroup;
  errorMessage: string | null = null;

  constructor(private fb: FormBuilder, private http: HttpClient, private router: Router, private authService: AuthService) {
    this.loginForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(8)]]
    });
  }

  onSubmit(): void {
    if (this.loginForm.valid) {
      this.http.post('http://localhost:8080/users/login', this.loginForm.value)
        .subscribe({
          next: (response: any) => {
            localStorage.setItem('token', response.token);
            localStorage.setItem('user_id', response.user.user_id);
             // ✅ OVO DODAJ
            console.log('[LOGIN] Logged user ID:', response.user.user_id);

            this.authService.login(response.user.user_type);
            this.errorMessage = null;
            alert('Login successful!');
            if (response.user.user_type == "student") {
              this.router.navigate(['/homepage']);
            } else if (response.user.user_type == "doctor") {
              this.router.navigate(['/appointment-management']);
            } else if (response.user.user_type == "cook") {
              this.router.navigate(['/home-radnik']);
            }
          },
          error: (err) => {
            this.errorMessage = err.error.error || 'Login failed';
          }
        });
    }
  }
}
