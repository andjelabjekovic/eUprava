import { Component, OnDestroy, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { Subscription, interval, of } from 'rxjs';
import { catchError, switchMap, startWith } from 'rxjs/operators';

type NotificationState = 'NEW' | 'SENT' | 'ERROR';

interface CateringFoodItem {
  name: string;
  type1: string;
  type2: string;
  quantity: number;
}

interface CateringNotification {
  id?: string;
  requestId: string;
  foods: CateringFoodItem[];
  note?: string;
  universityStatus: 'NEMA' | 'IMA';
  state: NotificationState;
  message?: string;
  lastError?: string;
  createdAt?: string;
  updatedAt?: string;
  sentAt?: string;
}

@Component({
  selector: 'app-home-radnik',
  templateUrl: './home-radnik.component.html',
  styleUrls: ['./home-radnik.component.css']
})
export class HomeRadnikComponent implements OnInit, OnDestroy {

  constructor(
    private router: Router,
    private http: HttpClient
  ) {}

  // ===== existing navigation =====
  view() { this.router.navigate(['/therapy-list']); }
  viewMenu() { this.router.navigate(['/food-list']); }
  createFood() { this.router.navigate(['/create-food-real']); }
  viewOrders() { this.router.navigate(['/order-list-radnik']); }
  viewAcceptedOrders() { this.router.navigate(['/accepted-orders']); }

  // ===== notifications =====
  showNotifications = false;

  notifications: CateringNotification[] = [];
  loadingNotifications = false;
  notificationsError: string | null = null;

  // requestId -> loading flag (da ne može double-click)
  confirming: Record<string, boolean> = {};

  // polling subscription
  private pollSub?: Subscription;

  /**
   * Ako ti Angular ide preko proxy/gateway, ostavi '' (relative).
   * Ako ide direktno na food-service port, stavi npr: 'http://localhost:8003'
   * ili ako koristiš gateway: 'http://localhost:8000'
   */
  private apiBase = 'http://localhost:8003';

  ngOnInit(): void {
    // Poll svake 4s (odmah + interval)
    this.pollSub = interval(4000).pipe(
      startWith(0),
      switchMap(() => this.fetchNotifications$())
    ).subscribe((list) => {
      if (list) {
        // Sort: NEW first, then ERROR, then SENT; inside by createdAt desc
        this.notifications = [...list].sort((a, b) => {
          const rank = (s: NotificationState) => s === 'NEW' ? 0 : (s === 'ERROR' ? 1 : 2);
          const r1 = rank(a.state);
          const r2 = rank(b.state);
          if (r1 !== r2) return r1 - r2;

          const t1 = a.createdAt ? new Date(a.createdAt).getTime() : 0;
          const t2 = b.createdAt ? new Date(b.createdAt).getTime() : 0;
          return t2 - t1;
        });
      }
    });
  }

  ngOnDestroy(): void {
    this.pollSub?.unsubscribe();
  }

  get unreadCount(): number {
    // NEW + ERROR tretiramo kao “zahteva pažnju”
    return this.notifications.filter(n => n.state === 'NEW' || n.state === 'ERROR').length;
  }

  toggleNotifications(): void {
    this.showNotifications = !this.showNotifications;
    // kada otvori panel, uradi refresh odmah
    if (this.showNotifications) {
      this.refreshNotifications();
    }
  }

  refreshNotifications(): void {
    this.fetchNotificationsOnce();
  }

  private fetchNotificationsOnce(): void {
    this.loadingNotifications = true;
    this.notificationsError = null;

    this.http.get<CateringNotification[]>(`${this.apiBase}/catering/notifications`).pipe(
      catchError(err => {
        this.notificationsError = 'Ne mogu da učitam obaveštenja (backend nedostupan).';
        this.loadingNotifications = false;
        return of([] as CateringNotification[]);
      })
    ).subscribe(list => {
      this.loadingNotifications = false;
      this.notifications = list ?? [];
    });
  }

  private fetchNotifications$() {
    return this.http.get<CateringNotification[]>(`${this.apiBase}/catering/notifications`).pipe(
      catchError(_ => of(null))
    );
  }

  confirmIma(requestId: string): void {
    if (!requestId) return;
    if (this.confirming[requestId]) return;

    this.confirming[requestId] = true;
    this.notificationsError = null;

    this.http.post<any>(`${this.apiBase}/catering/notifications/${requestId}/confirm`, {}).pipe(
      catchError(err => {
        // ne rušimo UI — samo pustimo polling da povuče ERROR state iz baze
        this.notificationsError = 'Slanje nije uspelo. Pokušaj ponovo.';
        return of(null);
      })
    ).subscribe(_ => {
      this.confirming[requestId] = false;
      // Brzi refresh (da odmah vidi "Poslato je")
      this.fetchNotificationsOnce();
    });
  }

  trackByRequestId(_: number, item: CateringNotification): string {
    return item.requestId;
  }
}