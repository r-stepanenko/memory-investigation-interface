import { describe, it, expect } from "vitest";
import {
    search,
    getDatasets,
    getExplain,
    getContext
} from "../src/services/searchService";


describe("searchService", () => {
    it("returns mock search result", async () => {

        const result = await search({
            dataset_id: "test",
            hints: {
                user_id: "ivan"
            }
        });

        expect(result.search_id).toBeDefined();
        expect(result.candidates.length).toBeGreaterThan(0);

    });


    it("returns mock datasets", async () => {
        const datasets = await getDatasets();

        expect(Array.isArray(datasets)).toBe(true);
        expect(datasets.length).toBeGreaterThan(0);

    });


    it("returns explain for event", async () => {

        const result = await getExplain(
            "mock-search-1",
            "evt_mock_32"
        );

        expect(result).toBeDefined();
        expect(result.score).toBe(100);

        expect(result.contributions.length)
            .toBeGreaterThan(0);

    });


    it("returns context for event", async () => {
        const result = await getContext(
            "evt_mock_32"
        );

        expect(result).toBeDefined();

        expect(result.event.event_id)
            .toBe("evt_mock_32");

    });

});