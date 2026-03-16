import axios from 'axios';
import { getToken } from '../auth';

const api = axios.create({
  baseURL: '/api',
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);

// Auth
export const login = (username: string, password: string) =>
  api.post('/auth/login', { username, password });
export const register = (username: string, password: string, invite_code?: string) =>
  api.post('/auth/register', { username, password, invite_code });

// Config
export const getConfig = () => api.get('/config');

// Domains
export const getDomains = () => api.get('/domains');
export const createDomain = (subdomain: string) => api.post('/domains', { subdomain });
export const deleteDomain = (id: number) => api.delete(`/domains/${id}`);

// Admin Users
export const getUsers = (status?: string) =>
  api.get('/admin/users', { params: status ? { status } : {} });
export const createUser = (data: any) => api.post('/admin/users', data);
export const updateUser = (id: number, data: any) => api.put(`/admin/users/${id}`, data);
export const activateUser = (id: number) => api.put(`/admin/users/${id}/activate`);
export const deleteUser = (id: number) => api.delete(`/admin/users/${id}`);

// Admin Domains
export const getAllDomains = () => api.get('/admin/domains');
export const adminCreateDomain = (data: any) => api.post('/admin/domains', data);
export const adminUpdateDomain = (id: number, data: any) => api.put(`/admin/domains/${id}`, data);
export const adminDeleteDomain = (id: number) => api.delete(`/admin/domains/${id}`);

// Invite Codes
export const getInviteCodes = () => api.get('/admin/invite-codes');
export const createInviteCode = (data: any) => api.post('/admin/invite-codes', data);
export const deleteInviteCode = (id: number) => api.delete(`/admin/invite-codes/${id}`);

export default api;
