import axios, { AxiosInstance } from 'axios';
import Constants from 'expo-constants';


const API_BASE_URL = Constants?.expoConfig?.extra?.apiBaseUrl || 'https://matiks-assessment.onrender.com';

const axiosInstance: AxiosInstance = axios.create({
    baseURL: API_BASE_URL,
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
});


axiosInstance.interceptors.request.use(
    (config) => {
        if (__DEV__) {
            console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`);
        }
        return config;
    },
    (error) => Promise.reject(error)
);


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


export const userAPI = {

    createUser: async (userId: string, username: string, initialRating: number): Promise<User> => {
        const response = await axiosInstance.post('/users', {
            user_id: userId,
            username,
            initial_rating: initialRating,
        });
        return response.data.data;
    },


    getUser: async (userId: string): Promise<User> => {
        const response = await axiosInstance.get(`/users/${userId}`);
        return response.data.data;
    },


    updateRating: async (userId: string, rating: number): Promise<User> => {
        const response = await axiosInstance.put(`/users/${userId}/rating`, {
            rating,
        });
        return response.data.data;
    },


    searchUser: async (username: string): Promise<SearchResult> => {
        const response = await axiosInstance.get('/users/search', {
            params: { username },
        });
        return response.data.data;
    },


    getLeaderboard: async (page: number = 1, pageSize: number = 100): Promise<LeaderboardResponse> => {
        const response = await axiosInstance.get('/leaderboard', {
            params: { page, page_size: pageSize },
        });
        return response.data.data;
    },


    getLeaderboardAroundUser: async (userId: string, contextSize: number = 10): Promise<LeaderboardResponse> => {
        const response = await axiosInstance.get(`/users/${userId}/leaderboard-context`, {
            params: { context_size: contextSize },
        });
        return response.data.data;
    },


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
