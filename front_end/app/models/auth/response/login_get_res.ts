export interface LoginUser {
  id: number;
  fullname: string;
  email: string;
  phone_number?: string;
  role: string;
  type: string;
  google_id: string;
  has_password: boolean;
}

export interface LoginResponse {
  success: boolean;
  message?: string;
  token: string;
  refresh_token: string;
  user: LoginUser;
}
