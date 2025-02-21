/**
 * NeoBank API Client
 * Centralized API client for all backend service calls
 */

import axios, { AxiosInstance, AxiosError } from 'axios';

// Types
export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  created_at: string;
}

export interface Account {
  id: string;
  user_id: string;
  account_type: string;
  currency: string;
  balance: string;
  created_at: string;
}

export interface Payment {
  id: string;
  from_account_id: string;
  to_account_id: string;
  amount: string;
  currency: string;
  status: string;
  description: string;
  created_at: string;
}

export interface Card {
  id: string;
  user_id: string;
  account_id: string;
  card_number: string;
  expiration_date: string;
  status: string;
  created_at: string;
}

export interface Product {
  id: string;
  name: string;
  description: string;
  type: string;
  active: boolean;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface TransferRequest {
  from_account_id: string;
  to_account_id: string;
  amount: string;
  currency: string;
  description?: string;
}

export interface ApiError {
  message: string;
  code: string;
}

// API Client Class
class NeoBankAPI {
  private client: AxiosInstance;
  private token: string | null = null;

  constructor() {
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL || '',
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor for auth
    this.client.interceptors.request.use((config) => {
      if (this.token) {
        config.headers.Authorization = `Bearer ${this.token}`;
      }
      return config;
    });

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ApiError>) => {
        if (error.response?.status === 401) {
          this.token = null;
          if (typeof window !== 'undefined') {
            localStorage.removeItem('token');
            window.location.href = '/login';
          }
        }
        return Promise.reject(error);
      }
    );

    // Load token from localStorage
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('token');
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
    }
  }

  // Auth endpoints
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/identity/login', data);
    this.setToken(response.data.token);
    return response.data;
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/identity/register', data);
    this.setToken(response.data.token);
    return response.data;
  }

  async logout(): Promise<void> {
    this.clearToken();
  }

  async getProfile(): Promise<User> {
    const response = await this.client.get<User>('/api/identity/profile');
    return response.data;
  }

  // Account endpoints
  async getAccounts(): Promise<Account[]> {
    const response = await this.client.get<Account[]>('/api/ledger/accounts');
    return response.data;
  }

  async getAccount(id: string): Promise<Account> {
    const response = await this.client.get<Account>(`/api/ledger/accounts/${id}`);
    return response.data;
  }

  async createAccount(data: { account_type: string; currency: string }): Promise<Account> {
    const response = await this.client.post<Account>('/api/ledger/accounts', data);
    return response.data;
  }

  // Payment endpoints
  async makeTransfer(data: TransferRequest): Promise<Payment> {
    const response = await this.client.post<Payment>('/api/payment/transfer', data);
    return response.data;
  }

  async getPayments(): Promise<Payment[]> {
    const response = await this.client.get<Payment[]>('/api/payment/payments');
    return response.data;
  }

  async getPayment(id: string): Promise<Payment> {
    const response = await this.client.get<Payment>(`/api/payment/payments/${id}`);
    return response.data;
  }

  // Card endpoints
  async getCards(): Promise<Card[]> {
    const response = await this.client.get<Card[]>('/api/card/cards');
    return response.data;
  }

  async issueCard(accountId: string): Promise<Card> {
    const response = await this.client.post<Card>('/api/card/cards', { account_id: accountId });
    return response.data;
  }

  async blockCard(cardId: string): Promise<void> {
    await this.client.post(`/api/card/cards/${cardId}/block`);
  }

  // Product endpoints
  async getProducts(): Promise<Product[]> {
    const response = await this.client.get<Product[]>('/api/product/products');
    return response.data;
  }

  async getProduct(id: string): Promise<Product> {
    const response = await this.client.get<Product>(`/api/product/products/${id}`);
    return response.data;
  }
}

// Singleton instance
export const api = new NeoBankAPI();
export default api;
