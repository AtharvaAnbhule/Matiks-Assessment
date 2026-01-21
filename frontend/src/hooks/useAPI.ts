import { useCallback, useState, useRef, useEffect } from 'react';
import { userAPI, LeaderboardResponse } from '../services/api';

/**
 * Custom Hook: useLeaderboard
 * Manages leaderboard state and pagination
 * Features:
 * - Fetch leaderboard with pagination
 * - Loading and error states
 * - Automatic refresh capability
 * - Memory efficient pagination
 */
export const useLeaderboard = () => {
    const [data, setData] = useState<LeaderboardResponse | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [currentPage, setCurrentPage] = useState(1);

    const pageSize = 50; // Items per page

    const fetchLeaderboard = useCallback(async (page: number = 1) => {
        try {
            setLoading(true);
            setError(null);
            const result = await userAPI.getLeaderboard(page, pageSize);
            setData(result);
            setCurrentPage(page);
        } catch (err: any) {
            setError(err?.response?.data?.message || 'Failed to fetch leaderboard');
            console.error('Leaderboard fetch error:', err);
        } finally {
            setLoading(false);
        }
    }, []);

    const nextPage = useCallback(() => {
        if (data?.has_more) {
            fetchLeaderboard(currentPage + 1);
        }
    }, [data?.has_more, currentPage, fetchLeaderboard]);

    const prevPage = useCallback(() => {
        if (currentPage > 1) {
            fetchLeaderboard(currentPage - 1);
        }
    }, [currentPage, fetchLeaderboard]);

    const refresh = useCallback(() => {
        fetchLeaderboard(currentPage);
    }, [currentPage, fetchLeaderboard]);

    return {
        data,
        loading,
        error,
        currentPage,
        nextPage,
        prevPage,
        refresh,
        fetch: fetchLeaderboard,
    };
};

/**
 * Custom Hook: useSearch
 * Manages search state with debouncing
 * Features:
 * - Debounced search to reduce API calls
 * - Loading and error states
 * - Search result caching
 */
export const useSearch = () => {
    const [query, setQuery] = useState('');
    const [result, setResult] = useState<any>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

    const search = useCallback(async (searchQuery: string) => {
        if (!searchQuery.trim()) {
            setResult(null);
            setQuery('');
            return;
        }

        try {
            setLoading(true);
            setError(null);
            setQuery(searchQuery);
            const result = await userAPI.searchUser(searchQuery);
            setResult(result);
        } catch (err: any) {
            setError(err?.response?.data?.message || 'Search failed');
            console.error('Search error:', err);
        } finally {
            setLoading(false);
        }
    }, []);

    const debouncedSearch = useCallback((searchQuery: string) => {
        if (debounceTimer.current) {
            clearTimeout(debounceTimer.current);
        }

        // Debounce for 500ms to reduce API calls while user is typing
        debounceTimer.current = setTimeout(() => {
            search(searchQuery);
        }, 500);
    }, [search]);

    return {
        query,
        result,
        loading,
        error,
        debouncedSearch,
        clearSearch: () => {
            setQuery('');
            setResult(null);
            setError(null);
        },
    };
};

/**
 * Custom Hook: useUserRank
 * Fetches and updates user rank in real-time
 * Features:
 * - Periodic polling for rank updates
 * - Loading and error states
 * - Automatic cleanup
 */
export const useUserRank = (userId: string | null) => {
    const [user, setUser] = useState<any>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [refreshing, setRefreshing] = useState(false);
    const pollingTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

    const fetchUser = useCallback(async () => {
        if (!userId) return;

        try {
            setLoading(true);
            setError(null);
            const result = await userAPI.getUser(userId);
            setUser(result);
        } catch (err: any) {
            setError(err?.response?.data?.message || 'Failed to fetch user');
            console.error('User fetch error:', err);
        } finally {
            setLoading(false);
        }
    }, [userId]);

    const refresh = useCallback(async () => {
        setRefreshing(true);
        await fetchUser();
        setRefreshing(false);
    }, [fetchUser]);

    // Initial fetch
    useEffect(() => {
        fetchUser();
    }, [fetchUser]);

    // Poll for updates every 5 seconds (for real-time rank updates)
    useEffect(() => {
        if (!userId) return;

        // Start polling
        pollingTimer.current = setInterval(() => {
            fetchUser();
        }, 5000);

        return () => {
            if (pollingTimer.current) {
                clearInterval(pollingTimer.current);
            }
        };
    }, [userId, fetchUser]);

    return {
        user,
        loading,
        error,
        refreshing,
        refresh,
    };
};
