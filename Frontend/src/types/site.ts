export type PageKey =
  | 'home'
  | 'services'
  | 'showroom'
  | 'how-it-works'
  | 'why-choose-us'
  | 'contact'
  | 'profile'
  | 'admin';

export type CarCategory = 'Luxury' | 'SUV' | 'Sport' | 'Business';

export type ServiceIconKey = 'shield' | 'zap' | 'star' | 'car' | 'map-pin' | 'calendar';

export interface PageLink {
  key: PageKey;
  label: string;
}

export interface PageContent {
  title: string;
  subtitle: string;
}

export interface ServiceItem {
  icon: ServiceIconKey;
  title: string;
  desc: string;
}

export interface ShowroomHint {
  title: string;
  desc: string;
}

export interface HowItWorksStep {
  step: string;
  title: string;
  desc: string;
}

export interface BenefitItem {
  title: string;
  desc: string;
}

export interface ContactInfo {
  email: string;
  phone: string;
}

export interface BookingFormValues {
  location: string;
  pickupDate: string;
  returnDate: string;
  pickupTime: string;
  category: CarCategory | 'Any';
}

export interface BookingFormErrors {
  location?: string;
  pickupDate?: string;
  returnDate?: string;
  pickupTime?: string;
}
