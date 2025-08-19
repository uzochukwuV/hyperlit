import { useState, useEffect } from "react";
import axios, { AxiosRequestConfig } from "axios";

/**
 * Simple hook for fetching data from an API endpoint.
 * @param url The endpoint to fetch.
 * @param config Optional Axios config.
 */
export function useApi<T>(url: string, config?: AxiosRequestConfig) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<any>(null);

  useEffect(() => {
    let isMounted = true;
    setLoading(true);
    axios(url, config)
      .then((res) => {
        if (isMounted) setData(res.data);
      })
      .catch((err) => {
        if (isMounted) setError(err);
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });
    return () => {
      isMounted = false;
    };
  }, [url]);

  return { data, loading, error };
}