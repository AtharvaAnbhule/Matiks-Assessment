import axios, { AxiosInstance } from 'axios';
import Constants from 'expo-constants';

/**
 * API Service Layer
 * Handles all HTTP communication with backend
 * Features:
 * - Error handling and retry logic
 * - Request/response interceptors
 * - Base URL configuration
 * - Timeout management
 */

const API_BASE_URL = Constants?.expoConfig?.extra?.apiBaseUrl || 'http://localhost:8080';

const axiosInstance: AxiosInstance = axios.create({
    baseURL: API_BASE_URL,
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor: Log requests in development
axiosInstance.interceptors.request.use(
    (config) => {
        if (__DEV__) {
            console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`);
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Response interceptor: Handle errors globally
axiosInstance.interceptors.response.use(
    (response) => response,
    (error) => {
        if (__DEV__) {
            console.log('[API Error]', error.response?.status, error.response?.data);
        }
        return Promise.reject(error);
    }
);

export interface User {
    id: string;
    username: string;
    rating: number;
    rank?: number;
}

export interface LeaderboardEntry {
    rank: number;
    username: string;
    rating: number;
}

export interface LeaderboardResponse {
    entries: LeaderboardEntry[];
    total: number;
    page: number;
    page_size: number;
    has_more: boolean;
}

export interface SearchResult {
    user: User | null;
    rank: number;
    found: boolean;
}

/**
 * User API endpoints
 */
export const userAPI = {
    /**
     * Create a new user
     * @param userId - Unique user identifier
     * @param username - Display username
     * @param initialRating - Starting rating (100-5000)
     */
    createUser: async (userId: string, username: string, initialRating: number): Promise<User> => {
        const response = await axiosInstance.post('/users', {
            user_id: userId,
            username,
            initial_rating: initialRating,
        });
        return response.data.data;
    },

    /**
     * Get user details with current rank
     * @param userId - User ID to fetch
     */
    getUser: async (userId: string): Promise<User> => {
        const response = await axiosInstance.get(`/users/${userId}`);
        return response.data.data;
    },

    /**
     * Update user's rating
     * Triggers rank recalculation server-side
     * @param userId - User ID to update
     * @param rating - New rating value (100-5000)
     */
    updateRating: async (userId: string, rating: number): Promise<User> => {
        const response = await axiosInstance.put(`/users/${userId}/rating`, {
            rating,
        });
        return response.data.data;
    },

    /**
     * Search for user by username
     * Case-insensitive search
     * @param username - Username to search for
     */
    searchUser: async (username: string): Promise<SearchResult> => {
        const response = await axiosInstance.get('/users/search', {
            params: { username },
        });
        return response.data.data;
    },

    /**
     * Get leaderboard with pagination
     * @param page - Page number (1-based)
     * @param pageSize - Items per page (default 100, max 1000)
     */
    getLeaderboard: async (page: number = 1, pageSize: number = 100): Promise<LeaderboardResponse> => {
        const response = await axiosInstance.get('/leaderboard', {
            params: { page, page_size: pageSize },
        });
        return response.data.data;
    },

    /**
     * Get leaderboard around user's position
     * Shows context: users ranked before and after target user
     * @param userId - User ID to center on
     * @param contextSize - How many entries before/after to show
     */
    getLeaderboardAroundUser: async (userId: string, contextSize: number = 10): Promise<LeaderboardResponse> => {
        const response = await axiosInstance.get(`/users/${userId}/leaderboard-context`, {
            params: { context_size: contextSize },
        });
        return response.data.data;
    },

    /**
     * Check service health
     */
    checkHealth: async (): Promise<boolean> => {
        try {
            const response = await axiosInstance.get('/health');
            return response.status === 200;
        } catch {
            return false;
        }
    },
};

export default axiosInstance;
