import type {
  BenefitItem,
  CarCategory,
  ContactInfo,
  HowItWorksStep,
  PageContent,
  PageKey,
  PageLink,
  ServiceItem,
  ShowroomHint,
} from '../types/site';

export const pages: PageLink[] = [
  { key: 'home', label: 'Home' },
  { key: 'services', label: 'Services' },
  { key: 'showroom', label: 'Showroom' },
  { key: 'news', label: 'News' },
  { key: 'why-choose-us', label: 'Why Choose Us' },
  { key: 'contact', label: 'Contact' },
];

export const pageContent: Record<PageKey, PageContent> = {
  home: {
    title: 'Drive the Future',
    subtitle: 'Premium luxury vehicles at your fingertips. Experience unparalleled comfort and style.',
  },
  services: {
    title: 'Premium Services',
    subtitle: 'Discover the full range of services designed to make your rental experience seamless.',
  },
  showroom: {
    title: 'Virtual Showroom',
    subtitle: 'Our showroom page is dedicated only to luxury vehicle previews and selection details.',
  },
  news: {
    title: 'AutoRent News',
    subtitle: 'Read the latest AutoRent fleet updates, announcements, and service notes.',
  },
  'why-choose-us': {
    title: 'Why Choose AutoRent',
    subtitle: 'Everything that makes our luxury rental service stand apart.',
  },
  contact: {
    title: 'Contact AutoRent',
    subtitle: 'Get in touch with our team for quick support and premium booking assistance.',
  },
  profile: {
    title: 'Your Profile',
    subtitle: 'Manage your AutoRent account and jump back into your rental experience.',
  },
  admin: {
    title: 'Admin Dashboard',
    subtitle: 'Manage live inventory and monitor fleet status.',
  },
};

export const services: ServiceItem[] = [
  { icon: 'shield', title: 'VIP Concierge', desc: 'Personalized service from booking to return' },
  { icon: 'zap', title: 'Instant Booking', desc: 'Reserve your vehicle in seconds' },
  { icon: 'star', title: 'Luxury Fleet', desc: 'Access to premium and exotic cars' },
  { icon: 'car', title: '24/7 Support', desc: 'Round-the-clock assistance' },
  { icon: 'map-pin', title: 'Global Network', desc: 'Pickup and drop-off worldwide' },
  { icon: 'calendar', title: 'Flexible Terms', desc: 'Custom rental periods and pricing' },
];

export const showroomHints: ShowroomHint[] = [
  { title: 'Database Inventory', desc: 'Vehicle cards are rendered from the backend car API.' },
  { title: 'Live Availability', desc: 'Only cars marked as available are shown publicly.' },
  { title: 'Admin Managed', desc: 'Fleet changes appear after admin create, edit, or delete actions.' },
];

export const howItWorksSteps: HowItWorksStep[] = [
  { step: '01', title: 'Choose & Book', desc: 'Select your preferred vehicle and dates' },
  { step: '02', title: 'Pickup', desc: 'Meet at your chosen location' },
  { step: '03', title: 'Enjoy', desc: 'Drive with confidence and style' },
];

export const benefits: BenefitItem[] = [
  {
    title: 'Unmatched Quality',
    desc: 'Every vehicle in our fleet undergoes rigorous maintenance and inspection.',
  },
  {
    title: 'Competitive Pricing',
    desc: 'Transparent pricing with no hidden fees or surprises.',
  },
  {
    title: 'Customer Satisfaction',
    desc: 'Rated 5-star by thousands of satisfied customers worldwide.',
  },
  {
    title: 'Innovation',
    desc: 'Cutting-edge technology for seamless booking and management.',
  },
];

export const contactInfo: ContactInfo = {
  email: 'info@autorent.com',
  phone: '+1 (555) 123-4567',
};

export const carCategories: Array<CarCategory | 'Any'> = ['Any', 'Luxury', 'SUV', 'Sport', 'Business'];
