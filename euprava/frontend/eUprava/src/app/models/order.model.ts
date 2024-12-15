export interface OrderData {
  id?: string; // ID porudžbine
  food: {
      id: string;
      foodName?: string;
  };
  statusO?: string; // Status porudžbine ('Prihvacena' ili 'Neprihvacena')
  // Dodaj ostala polja ako je potrebno
}
