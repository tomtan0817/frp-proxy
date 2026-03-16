export const getToken = () => localStorage.getItem('token');
export const setToken = (token: string) => localStorage.setItem('token', token);
export const removeToken = () => localStorage.removeItem('token');

export const parseJwt = (token: string) => {
  try {
    return JSON.parse(atob(token.split('.')[1]));
  } catch {
    return null;
  }
};

export const getUser = () => {
  const token = getToken();
  if (!token) return null;
  const payload = parseJwt(token);
  if (!payload || payload.exp * 1000 < Date.now()) {
    removeToken();
    return null;
  }
  return payload;
};

export const isAdmin = () => getUser()?.role === 'admin';
