export const getToken = () => localStorage.getItem('token');
export const setToken = (token: string) => localStorage.setItem('token', token);
export const removeToken = () => localStorage.removeItem('token');

export const parseJwt = (token: string) => {
  try {
    let base64 = token.split('.')[1];
    // Convert Base64URL to standard Base64
    base64 = base64.replace(/-/g, '+').replace(/_/g, '/');
    // Add padding
    while (base64.length % 4) {
      base64 += '=';
    }
    return JSON.parse(atob(base64));
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
