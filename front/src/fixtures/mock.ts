export const mockDatasets = [
    {
        id: "test-dataset",
        name: "Test Dataset",
        size: 33,
        period: "2026-06-16T09:50:00Z — 2026-06-20T11:40:00Z",
        description: "Набор событий пользователей"
    }
];


export const mockSearch = {
    candidates: [
        {
            event: {
                event_id: "evt_mock_32",
                timestamp: "2026-06-20T11:20:00Z",
                user_id: "ivan",
                file_name: "client_data.zip",
                action: "file_copy",
                destination_type: "usb"
            },

            score: 100,

            matched_hints: [
                "user_id exact",
                "nearby event found"
            ],

            contributions: [
                {
                    hint: "user_id",
                    type: "exact",
                    value: "ivan",
                    query: "ivan",
                    points: 90,
                    matched: true,
                    reason: "Точное совпадение user_id"
                },
                {
                    hint: "nearby",
                    type: "context",
                    value: "email_send",
                    query: "email_send",
                    points: 10,
                    matched: true,
                    reason: "Найдено связанное событие в заданном интервале"
                }
            ]
        },


        {
            event: {
                event_id: "evt_mock_24",
                timestamp: "2026-06-19T09:40:00Z",
                user_id: "ivanov",
                file_name: "clients.zip",
                action: "file_copy",
                destination_type: "usb"
            },

            score: 55,

            matched_hints: [
                "user_id substring",
                "nearby event found"
            ],

            contributions: [
                {
                    hint: "user_id",
                    type: "substring",
                    value: "ivanov",
                    query: "ivan",
                    points: 45,
                    matched: true,
                    reason: "Частичное совпадение user_id"
                },
                {
                    hint: "nearby",
                    type: "context",
                    value: "email_send",
                    query: "email_send",
                    points: 10,
                    matched: true,
                    reason: "Найдено связанное событие"
                }
            ]
        },


        {
            event: {
                event_id: "evt_mock_4",
                timestamp: "2026-06-16T09:50:00Z",
                user_id: "petrov",
                file_name: "report.xlsx",
                action: "create_archive",
                destination_type: "internal"
            },

            score: 40,

            matched_hints: [
                "user_id fuzzy"
            ],

            contributions: [
                {
                    hint: "user_id",
                    type: "fuzzy",
                    value: "petrov",
                    query: "petrov",
                    points: 40,
                    matched: true,
                    reason: "Нечёткое совпадение user_id"
                },
                {
                    hint: "nearby",
                    type: "none",
                    value: "",
                    query: "",
                    points: 0,
                    matched: false,
                    reason: "Связанное событие не найдено"
                }
            ]
        }
    ]
};


export const mockExplain: any = {
    "evt_mock_32": {
        score: 100,

        contributions: [
            {
                hint: "user_id",
                type: "exact",
                value: "ivan",
                query: "ivan",
                points: 90,
                matched: true,
                reason: "Точное совпадение user_id"
            },
            {
                hint: "nearby",
                type: "context",
                value: "email_send",
                query: "email_send",
                points: 10,
                matched: true,
                reason: "Найдено связанное событие"
            }
        ]
    },


    "evt_mock_24": {
        score: 55,

        contributions: [
            {
                hint: "user_id",
                type: "substring",
                value: "ivanov",
                query: "ivan",
                points: 45,
                matched: true,
                reason: "Частичное совпадение user_id"
            },
            {
                hint: "nearby",
                type: "context",
                value: "email_send",
                query: "email_send",
                points: 10,
                matched: true,
                reason: "Найдено связанное событие"
            }
        ]
    },


    "evt_mock_4": {
        score: 40,

        contributions: [
            {
                hint: "user_id",
                type: "fuzzy",
                value: "petrov",
                query: "petrov",
                points: 40,
                matched: true,
                reason: "Нечёткое совпадение user_id"
            }
        ]
    }
};


export const mockContext: any = {

    "evt_mock_32": {
        event: {
            event_id: "evt_mock_32",
            timestamp: "2026-06-20T11:20:00Z",
            user_id: "ivan",
            action: "file_copy",
            file_name: "client_data.zip",
            destination_type: "usb"
        },

        before: [
            {
                event_id: "evt_mock_31",
                timestamp: "2026-06-20T11:00:00Z",
                user_id: "ivan",
                action: "create_archive",
                file_name: "client_data.zip",
                destination_type: "internal"
            }
        ],

        after: [
            {
                event_id: "evt_mock_33",
                timestamp: "2026-06-20T11:40:00Z",
                user_id: "ivan",
                action: "email_send",
                file_name: "client_data.zip",
                destination_type: "external"
            }
        ]
    },


    "evt_mock_24": {
        event: {
            event_id: "evt_mock_24",
            timestamp: "2026-06-19T09:40:00Z",
            user_id: "ivanov",
            action: "file_copy",
            file_name: "clients.zip",
            destination_type: "usb"
        },

        before: [],

        after: [
            {
                event_id: "evt_mock_email",
                timestamp: "2026-06-19T10:00:00Z",
                user_id: "ivanov",
                action: "email_send",
                file_name: "clients.zip",
                destination_type: "external"
            }
        ]
    },


    "evt_mock_4": {
        event: {
            event_id: "evt_mock_4",
            timestamp: "2026-06-16T09:50:00Z",
            user_id: "petrov",
            action: "create_archive",
            file_name: "report.xlsx",
            destination_type: "internal"
        },

        before: [],

        after: []
    }
};