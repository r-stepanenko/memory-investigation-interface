export const mockDatasets = [
    {
        id: "test-dataset",
        name: "Test Dataset",
        size: 3,
        period: "2026-06-19 — 2026-06-20",
        description: "Тестовый набор событий для демонстрации поиска"
    }
];

export const mockSearch = {
    candidates: [
        {
            event: {
                event_id: "evt_mock_1",
                timestamp: "2026-06-20T11:40:00Z",
                user_id: "ivan",
                file_name: "client_data.zip",
                action: "email_send",
                destination_type: "external"
            },
            score: 90,
            matched_hints: [
                "user_id exact",
                "file_name substring"
            ]
        },

        {
            event: {
                event_id: "evt_mock_2",
                timestamp: "2026-06-20T10:20:00Z",
                user_id: "alex",
                file_name: "backup.zip",
                action: "file_copy",
                destination_type: "usb"
            },
            score: 60,
            matched_hints: [
                "file_name substring"
            ]
        },

        {
            event: {
                event_id: "evt_mock_3",
                timestamp: "2026-06-19T09:30:00Z",
                user_id: "petrov",
                file_name: "report.xlsx",
                action: "create_archive",
                destination_type: "internal"
            },
            score: 40,
            matched_hints: [
                "user_id fuzzy"
            ]
        }
    ]
};

export const mockExplain: any = {
    "evt_mock_1": {
        score: 90,
        contributions: [
            {
                hint: "user_id",
                type: "exact",
                value: "ivan",
                points: 50
            },

            {
                hint: "file_name",
                type: "substring",
                value: "client_data.zip",
                points: 25
            },

            {
                hint: "destination_type",
                type: "exact",
                value: "external",
                points: 15
            }
        ]
    },

    "evt_mock_2": {
        score: 60,
        contributions: [
            {
                hint: "file_name",
                type: "substring",
                value: "backup.zip",
                points: 40
            },
            {
                hint: "action",
                type: "exact",
                value: "file_copy",
                points: 20
            }
        ]
    },

    "evt_mock_3": {
        score: 40,
        contributions: [
            {
                hint: "user_id",
                type: "fuzzy",
                value: "petrov",
                points: 40
            }
        ]
    }

};

export const mockContext: any = {
    evt_mock_1: {
        event: {
            event_id: "evt_mock_1",
            timestamp: "2026-06-20T11:40:00Z",
            user_id: "ivan",
            action: "email_send",
            file_name: "client_data.zip",
            destination_type: "external"
        },
        before: [
            {
                event_id: "evt_mock_0",
                timestamp: "2026-06-20T11:20:00Z",
                user_id: "ivan",
                action: "file_copy",
                file_name: "client_data.zip",
                destination_type: "usb"
            }
        ],
        after: []
    }

};