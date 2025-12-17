export interface OrderData {
  id?: string; // ID porudžbine
  food: {
      id: string;
      foodName?: string;
      type1?: string;
     type2?: string;
  };
  userId?: string; // ID korisnika koji je kreirao porudžbinu
  statusO?: string; // Status porudžbine ('Prihvacena' ili 'Neprihvacena')
  statusO2?: string; // Status porudžbine ('Otkazana' ili 'Neotkazana')
}
