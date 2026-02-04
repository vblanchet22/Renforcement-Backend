// User types
export interface User {
  id: string;
  email: string;
  nom: string;
  prenom: string;
  telephone?: string;
  avatar_url?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserInfo {
  id: string;
  email: string;
  nom: string;
  prenom: string;
  telephone?: string;
  avatar_url?: string;
}

// Auth types
export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: UserInfo;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  nom: string;
  prenom: string;
  telephone?: string;
}

// Colocation types
export interface Colocation {
  id: string;
  name: string;
  description?: string;
  address?: string;
  invite_code: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface ColocationMember {
  id: string;
  user_id: string;
  colocation_id: string;
  role: 'admin' | 'member';
  joined_at: string;
  user?: User;
}

export interface ColocationWithMembers extends Colocation {
  members: ColocationMember[];
}

// Category types
export interface Category {
  id: string;
  name: string;
  icon: string;
  color: string;
  colocation_id?: string;
  is_global: boolean;
}

export interface CategoryStat {
  category: Category;
  total_amount: number;
  expense_count: number;
  percentage: number;
}

// Expense types
export type SplitType = 'equal' | 'percentage' | 'custom';
export type Recurrence = 'daily' | 'weekly' | 'monthly' | 'yearly';

export interface ExpenseSplit {
  id: string;
  expense_id: string;
  user_id: string;
  amount: number;
  percentage?: number;
  is_settled: boolean;
  user?: User;
}

export interface Expense {
  id: string;
  colocation_id: string;
  paid_by: string;
  category_id: string;
  title: string;
  description?: string;
  amount: number;
  split_type: SplitType;
  expense_date: string;
  recurring_id?: string;
  created_at: string;
  payer?: User;
  category?: Category;
  splits: ExpenseSplit[];
}

export interface RecurringExpense {
  id: string;
  colocation_id: string;
  paid_by: string;
  category_id: string;
  title: string;
  description?: string;
  amount: number;
  split_type: SplitType;
  recurrence: Recurrence;
  next_due_date: string;
  end_date?: string;
  is_active: boolean;
  created_at: string;
  payer?: User;
  category?: Category;
}

// Balance types
export interface UserBalance {
  user_id: string;
  user?: User;
  total_paid: number;
  total_owed: number;
  net_balance: number;
}

export interface SimplifiedDebt {
  from_user_id: string;
  from_user?: User;
  to_user_id: string;
  to_user?: User;
  amount: number;
}

// Payment types
export type PaymentStatus = 'pending' | 'confirmed' | 'rejected';

export interface Payment {
  id: string;
  colocation_id: string;
  from_user_id: string;
  to_user_id: string;
  amount: number;
  status: PaymentStatus;
  note?: string;
  confirmed_at?: string;
  created_at: string;
  from_user?: User;
  to_user?: User;
}

// Decision types
export type DecisionStatus = 'open' | 'closed';

export interface DecisionOption {
  index: number;
  text: string;
}

export interface Decision {
  id: string;
  colocation_id: string;
  created_by: string;
  title: string;
  description?: string;
  options: DecisionOption[];
  status: DecisionStatus;
  deadline?: string;
  allow_multiple: boolean;
  is_anonymous: boolean;
  created_at: string;
  creator?: User;
  user_votes?: number[];
  total_votes: number;
}

export interface DecisionResult {
  option_index: number;
  option_text: string;
  vote_count: number;
  percentage: number;
  voters?: User[];
}

// Fund types
export interface FundContribution {
  id: string;
  fund_id: string;
  user_id: string;
  amount: number;
  note?: string;
  created_at: string;
  user?: User;
}

export interface ContributorSummary {
  user_id: string;
  user?: User;
  total_contributed: number;
}

export interface CommonFund {
  id: string;
  colocation_id: string;
  name: string;
  description?: string;
  target_amount: number;
  current_amount: number;
  is_active: boolean;
  created_by: string;
  created_at: string;
  creator?: User;
  contributors?: ContributorSummary[];
  progress_percentage: number;
}

// Notification types
export type NotificationType =
  | 'expense_created'
  | 'expense_updated'
  | 'expense_deleted'
  | 'payment_received'
  | 'payment_confirmed'
  | 'payment_rejected'
  | 'member_joined'
  | 'member_left'
  | 'invitation_received'
  | 'decision_created'
  | 'decision_closed'
  | 'fund_contribution'
  | 'fund_goal_reached';

export interface Notification {
  id: string;
  user_id: string;
  colocation_id: string;
  type: NotificationType;
  title: string;
  body: string;
  data?: Record<string, string>;
  is_read: boolean;
  created_at: string;
}

// Forecast types
export interface CategoryForecast {
  category_id: string;
  category?: Category;
  amount: number;
}

export interface MonthlyForecast {
  month: string;
  year: number;
  total_amount: number;
  categories: CategoryForecast[];
}

// API Response types
export interface ApiError {
  code: number;
  message: string;
  details?: string[];
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}
