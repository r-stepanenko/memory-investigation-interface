import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import App from "../App";
import { search } from "../services/searchService";

vi.mock("../config", () => ({
    API_URL: "http://localhost:8080",
    USE_MOCK: false,
}));
vi.mock("../services/searchService", () => ({
    search: vi.fn(async () => ({
        search_id: "test-search",
        candidates: [
            {
                score: 100,
                event: {
                    event_id: "evt-1",
                    action: "file_copy",
                    user_id: "ivan",
                    file_name: "test.txt",
                    destination_type: "usb",
                    timestamp: "2026-06-20T10:00:00Z"
                },
                matched_hints: [],
                contributions: []
            },
            {
                score: 40,
                event: {
                    event_id: "evt-2",
                    action: "email_send",
                    user_id: "petrov",
                    file_name: "mail.txt",
                    destination_type: "external",
                    timestamp: "2026-06-20T11:00:00Z"
                },
                matched_hints: [],
                contributions: []
            }
        ]
    })),

    getDatasets: vi.fn(async () => [
        {
            id: "test",
            name: "Test dataset",
            size: 1,
            period: "2026",
            description: "test"
        }
    ]),

    getExplain: vi.fn(),
}));

beforeEach(() => {
    vi.resetAllMocks();
    localStorage.clear();

    vi.mocked(search).mockResolvedValue({
        search_id: "test-search",
        candidates: [
            {
                score: 100,
                event: {
                    event_id: "evt-1",
                    action: "file_copy",
                    user_id: "ivan",
                    file_name: "test.txt",
                    destination_type: "usb",
                    timestamp: "2026-06-20T10:00:00Z"
                },
                matched_hints: [],
                contributions: []
            },
            {
                score: 40,
                event: {
                    event_id: "evt-2",
                    action: "email_send",
                    user_id: "petrov",
                    file_name: "mail.txt",
                    destination_type: "external",
                    timestamp: "2026-06-20T11:00:00Z"
                },
                matched_hints: [],
                contributions: []
            }
        ]
    });
});

describe("App", () => {

    it("renders Dataset section", () => {

        render(<App />);

        expect(
            screen.getByText("Dataset")
        ).toBeInTheDocument();

    });


    it("loads mock datasets", async () => {

        render(<App />);

        await waitFor(() => {

            expect(
                screen.getByText("Название:")
            ).toBeInTheDocument();

        });

    });

    it("renders search button", async () => {

        render(<App />);

        expect(
            await screen.findByRole(
                "button",
                { name: "Поиск" }
            )
        ).toBeInTheDocument();

    });

    it("performs search and shows results", async () => {

        const user = userEvent.setup();

        render(<App />);

        const input = await screen.findByPlaceholderText(
            "User ID"
        );

        await user.type(input, "ivan");


        await user.click(
            screen.getByRole(
                "button",
                { name: "Поиск" }
            )
        );


        await waitFor(() => {
            expect(
                screen.getByText("file_copy")
            ).toBeInTheDocument();
        });

    });

    it("filters results by minimum score", async () => {

        const user = userEvent.setup();

        render(<App />);


        await user.click(
            screen.getByRole(
                "button",
                { name: "Поиск" }
            )
        );


        await waitFor(() => {
            expect(
                screen.getByText("file_copy")
            ).toBeInTheDocument();

            expect(
                screen.getByText("email_send")
            ).toBeInTheDocument();
        });


        const scoreInput = screen.getByPlaceholderText(
            "Минимальный score"
        );


        await user.type(scoreInput, "80");


        await waitFor(
            () => {
                expect(
                    screen.queryByText("email_send")
                ).not.toBeInTheDocument();
            },
            { timeout: 3000 }
        );

    });

    it("sorts results by score ascending", async () => {

        const user = userEvent.setup();

        const { container } = render(<App />);


        await user.click(
            screen.getByRole(
                "button",
                { name: "Поиск" }
            )
        );


        await waitFor(() => {
            expect(
                screen.getByText("file_copy")
            ).toBeInTheDocument();

            expect(
                screen.getByText("email_send")
            ).toBeInTheDocument();
        });


        const sortOrder = screen.getByDisplayValue(
            "По убыванию"
        );


        await user.selectOptions(
            sortOrder,
            "asc"
        );


        await waitFor(() => {

            const cards = container.querySelectorAll(
                ".candidate-card"
            );


            expect(cards.length).toBe(2);


            expect(cards[0])
                .toHaveTextContent("email_send");


            expect(cards[1])
                .toHaveTextContent("file_copy");

        });

    });

    it("shows validation error when time tolerance has no time", async () => {

        const user = userEvent.setup();

        render(<App />);


        const toleranceInput = screen.getByPlaceholderText(
            "Time tolerance"
        );


        await user.type(
            toleranceInput,
            "30m"
        );


        await user.click(
            screen.getByRole(
                "button",
                { name: "Поиск" }
            )
        );


        await waitFor(() => {

            expect(
                screen.getByText(
                    "Укажите примерное время"
                )
            ).toBeInTheDocument();

        });

    });

    it("shows validation error for invalid minimum score", async () => {

        const user = userEvent.setup();

        render(<App />);


        const scoreInput = screen.getByPlaceholderText(
            "Min score"
        );


        await user.type(
            scoreInput,
            "150"
        );


        await user.click(
            screen.getByRole(
                "button",
                { name: "Поиск" }
            )
        );


        await waitFor(() => {

            expect(
                screen.getByText(
                    "Min score должен быть числом от 0 до 100"
                )
            ).toBeInTheDocument();

        });

    });

    it("sends search request with user filter", async () => {

        const user = userEvent.setup();

        const searchMock = vi.fn(async (request: any) => ({
            search_id: "test",
            candidates: []
        }));

        vi.mocked(search).mockImplementation(searchMock);

        render(<App />);

        const input = await screen.findByPlaceholderText(
            "User ID"
        );

        await user.type(input, "ivan");

        await user.click(
            screen.getByRole(
                "button",
                { name: "Поиск" }
            )
        );

        await waitFor(() => {
            expect(searchMock)
                .toHaveBeenCalled();
        });

        expect(searchMock.mock.calls[0][0])
            .toMatchObject({
                hints: {
                    user_id: "ivan"
                }
            });

    });

    it("formats score correctly", async () => {

        const user = userEvent.setup();

        render(<App />);


        await screen.findByText("Название:");


        await user.click(
            screen.getByRole(
                "button",
                { name: "Поиск" }
            )
        );


        await waitFor(() => {

            expect(
                screen.getByText("100.00")
            ).toBeInTheDocument();

        });

    });


    it("shows empty result message", async () => {

        vi.mocked(search).mockResolvedValue({
            search_id: "empty",
            candidates: []
        });

        render(<App />);

        await userEvent.setup()
            .click(
                await screen.findByRole(
                    "button",
                    { name: "Поиск" }
                )
            );


        await waitFor(() => {

            expect(
                screen.getByText(
                    "Ничего не найдено"
                )
            ).toBeInTheDocument();

        });

    });

    it("shows backend error", async () => {

        vi.mocked(search).mockRejectedValue({
            code: "ERR_NETWORK"
        });


        const user = userEvent.setup();

        render(<App />);


        await user.click(
            screen.getByRole(
                "button",
                { name: "Поиск" }
            )
        );


        await waitFor(() => {

            expect(
                screen.getByText(
                    "Backend недоступен."
                )
            ).toBeInTheDocument();

        });

    });

    it("restores search from history", async () => {

        localStorage.setItem(
            "search-history",
            JSON.stringify([
                {
                    id: "1",
                    name: "test",
                    created: new Date().toISOString(),
                    request: {
                        dataset_id: "test",
                        hints: {
                            user_id: "ivan"
                        },
                        scoring: {}
                    }
                }
            ])
        );


        render(<App />);


        await waitFor(() => {

            expect(
                screen.getByText("test")
            ).toBeInTheDocument();

        });

    });

    it("adds candidate to comparison", async () => {

        const user = userEvent.setup();

        render(<App />);


        await user.click(
            screen.getByRole(
                "button",
                {
                    name: "Поиск"
                }
            )
        );


        await waitFor(() => {

            expect(
                screen.getByText("file_copy")
            )
                .toBeInTheDocument();

        });


        await user.click(
            screen.getAllByRole(
                "button",
                {
                    name: "Добавить в сравнение"
                }
            )[0]
        );


        await waitFor(() => {

            expect(
                screen.getByText(
                    "Выбранный кандидат:"
                )
            )
                .toBeInTheDocument();


        });

    });

});