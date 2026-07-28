import axios from "axios";
import { API_URL, USE_MOCK } from "../config";

import {
    mockSearch,
    mockExplain,
    mockContext,
    mockDatasets,
} from "../fixtures/mock";

export async function search(request: any) {
    if (USE_MOCK) {
        return {
            search_id: "mock-search-1",
            candidates: mockSearch.candidates,
        };
    }

    const res = await axios.post(`${API_URL}/api/search`, request);
    return res.data;
}

export async function getDatasets() {
    if (USE_MOCK) {
        return mockDatasets;
    }

    const res = await axios.get(`${API_URL}/api/datasets`);
    return res.data;
}

export async function getExplain(searchId: string, eventId: string) {
    if (USE_MOCK) {
        return mockExplain[eventId];
    }

    const res = await axios.get(
        `${API_URL}/api/search/${searchId}/candidates/${eventId}/explain`
    );

    return res.data;
}

export async function getContext(eventId: string) {
    if (USE_MOCK) {
        return mockContext[eventId];
    }

    const res = await axios.get(
        `${API_URL}/api/events/${eventId}/context`
    );

    return res.data;
}