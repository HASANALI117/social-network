import { useState, useCallback } from 'react';

interface ApiResponse<T> {
  data: T | null;
  error: Error | null;
  isLoading: boolean;
}

interface UseRequestReturn<T> extends ApiResponse<T> {
  get: (url: string, onSuccess?: (data: T) => void) => Promise<T | null>;
  post: <B>(url: string, body: B, onSuccess?: (data: T) => void) => Promise<T | null>;
  put: <B>(url: string, body: B, onSuccess?: (data: T) => void) => Promise<T | null>;
  del: (url: string, onSuccess?: (data: T) => void) => Promise<T | null>;
}

export function useRequest<T = any>(): UseRequestReturn<T> {
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<T | null>(null);

  const handleRequest = useCallback(async (
    url: string, 
    options: RequestInit,
    onSuccess?: (data: T) => void
  ): Promise<T | null> => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`API error: ${response.status} ${response.statusText}`);
      }

      const result = await response.json();
      setData(result);
      
      // Call onSuccess callback if provided
      if (onSuccess) {
        onSuccess(result);
      }
      
      return result;
    } catch (err: any) {
      setError(err instanceof Error ? err : new Error(String(err)));
      return null;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const get = useCallback((url: string, onSuccess?: (data: T) => void) => {
    return handleRequest(url, { method: 'GET' }, onSuccess);
  }, [handleRequest]);

  const post = useCallback(<B>(url: string, body: B, onSuccess?: (data: T) => void) => {
    return handleRequest(url, {
      method: 'POST',
      body: JSON.stringify(body),
    }, onSuccess);
  }, [handleRequest]);

  const put = useCallback(<B>(url: string, body: B, onSuccess?: (data: T) => void) => {
    return handleRequest(url, {
      method: 'PUT',
      body: JSON.stringify(body),
    }, onSuccess);
  }, [handleRequest]);

  const del = useCallback((url: string, onSuccess?: (data: T) => void) => {
    return handleRequest(url, { method: 'DELETE' }, onSuccess);
  }, [handleRequest]);

  return { data, error, isLoading, get, post, put, del };
}