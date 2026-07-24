import { useEffect, useState } from "react";
import axios from "axios";
import {
  mockSearch,
  mockExplain,
  mockDatasets,
  mockContext
} from "./fixtures/mock";

function App() {
  const [user, setUser] = useState("");
  const [fileName, setFileName] = useState("");
  const [action, setAction] = useState("");
  const [destinationType, setDestinationType] = useState("");
  const [events, setEvents] = useState<any[]>([]);
  const [message, setMessage] = useState("");
  const [datasets, setDatasets] = useState<any[]>([]);
  const [dataset, setDataset] = useState("");
  const [searchId, setSearchId] = useState("");
  const [selectedEventId, setSelectedEventId] = useState("");
  const [context, setContext] = useState<any>(null);
  const [loadingContext, setLoadingContext] = useState(false);
  const [explain, setExplain] = useState<any>(null);
  const [selectedExplainId, setSelectedExplainId] = useState("");
  const [isMockMode, setIsMockMode] = useState(false);
  const [nearbyAction, setNearbyAction] = useState("");
  const [nearbyTolerance, setNearbyTolerance] = useState("");
  const [timeAround, setTimeAround] = useState("");
  const [timeTolerance, setTimeTolerance] = useState("");
  const [limit, setLimit] = useState("");
  const [minScore, setMinScore] = useState("");
  const [channel, setChannel] = useState("");
  const [severity, setSeverity] = useState("");
  const [before, setBefore] = useState("");
  const [after, setAfter] = useState("");

  useEffect(() => {
    async function loadDatasets() {
      try {
        const res = await axios.get(
          "http://localhost:8080/api/datasets"
        );

        const formattedDatasets = res.data;

        setDatasets(formattedDatasets);

        if (formattedDatasets.length > 0) {
          setDataset(formattedDatasets[0].id);
        }

      } catch (error) {

        console.log("Backend недоступен, включен mock");

        setIsMockMode(true);
        setDatasets(mockDatasets);
        setDataset(mockDatasets[0].id);
        setMessage("");
      }

    }
    loadDatasets();

  }, []);

  async function getContext(eventId: string) {
    setLoadingContext(true);
    try {
      if (isMockMode) {
        setContext(mockContext[eventId]);
        setLoadingContext(false);
        return;
      }

      const res = await axios.get(
        `http://localhost:8080/api/events/${eventId}/context`
      );

      setContext(res.data);
    }
    catch (error) {
      console.log(error);

      setContext(null);
    }
    finally {
      setLoadingContext(false);
    }
  }

  async function search() {
    setEvents([]);
    setExplain(null);
    setSelectedExplainId("");

    setMessage(
      "Backend недоступен. Используется mock-режим"
    );

    if (!nearbyAction && nearbyTolerance) {
      setMessage("Укажите Nearby action");
      return;
    }

    if (nearbyAction && !nearbyTolerance) {
      setMessage("Укажите Nearby tolerance");
      return;
    }

    if (!timeAround && timeTolerance) {
      setMessage("Укажите примерное время");
      return;
    }

    if (timeAround && !timeTolerance) {
      setMessage("Укажите tolerance времени");
      return;
    }

    const durationRegex = /^\d+[smhd]$/;

    if (timeAround && isNaN(new Date(timeAround).getTime())) {
      setMessage("Некорректная дата");
      return;
    }

    if (limit && (isNaN(Number(limit)) || Number(limit) <= 0)) {
      setMessage("Limit должен быть положительным числом");
      return;
    }

    if ((before && !after) || (!before && after)) {
      setMessage("Before и After должны быть указаны вместе");
      return;
    }

    if (
      minScore &&
      (
        isNaN(Number(minScore)) ||
        Number(minScore) < 0 ||
        Number(minScore) > 100
      )
    ) {
      setMessage("Min score должен быть числом от 0 до 100");
      return;
    }

    if (timeTolerance && !durationRegex.test(timeTolerance)) {
      setMessage("Time tolerance должен быть в формате 30m, 1h, 2d");
      return;
    }

    if (nearbyTolerance && !durationRegex.test(nearbyTolerance)) {
      setMessage("Nearby tolerance должен быть в формате 30m, 1h, 2d");
      return;
    }

    if (before && !durationRegex.test(before)) {
      setMessage("Before должен быть в формате 30m, 1h, 2d");
      return;
    }

    if (after && !durationRegex.test(after)) {
      setMessage("After должен быть в формате 30m, 1h, 2d");
      return;
    }

    try {
      const res = await axios.post(
        "http://localhost:8080/api/search",
        {
          dataset_id: dataset,

          time: timeAround
            ? {
              around: new Date(timeAround).toISOString(),
              tolerance: timeTolerance
            }
            : undefined,

          hints: {
            user_id: user,
            file_name: fileName,
            action: action,
            destination_type: destinationType,
            channel: channel,
            severity: severity
          },

          context:
            before || after || nearbyAction
              ? {
                before: before || undefined,
                after: after || undefined,
                require_nearby: nearbyAction
                  ? [
                    {
                      action: nearbyAction,
                      within: nearbyTolerance
                    }
                  ]
                  : undefined,
              }
              : undefined,

          scoring: {
            limit: limit ? Number(limit) : undefined,
            min_score: minScore ? Number(minScore) : undefined
          }
        }
      );
      setSearchId(res.data.search_id);


      const candidates = res.data.candidates ?? [];

      if (candidates.length === 0) {
        setMessage("Ничего не найдено");

      }
      else {
        setMessage("");
        setEvents(candidates);
      }

    } catch (error: any) {
      console.log("SEARCH ERROR:", error.response?.data);
      console.log("STATUS:", error.response?.status);
      console.log("ERROR:", error.message);

      setIsMockMode(true);
      setSearchId("mock-search-1");
      setEvents(mockSearch.candidates);
      setMessage(
        "Backend недоступен. Используется mock-режим"
      );
    }
  }

  async function getExplain(eventId: string) {
    if (selectedExplainId === eventId) {
      setExplain(null);
      setSelectedExplainId("");

      return;
    }

    if (isMockMode) {
      setExplain(
        mockExplain[eventId]
      );
      setSelectedExplainId(eventId);

      return;
    }

    try {
      const res = await axios.get(
        `http://localhost:8080/api/search/${searchId}/candidates/${eventId}/explain`

      );

      setExplain(res.data);
      setSelectedExplainId(eventId);

    }
    catch (error) {
      console.log(error);

    }
  }

  const requestPreview = {
    dataset_id: dataset,
    time: timeAround
      ? {
        around: new Date(timeAround).toISOString(),
        tolerance: timeTolerance
      }
      : undefined,
    hints: {
      user_id: user,
      file_name: fileName,
      action: action,
      destination_type: destinationType,
      channel: channel,
      severity: severity
    },
    context:
      before || after || nearbyAction
        ? {
          before: before || undefined,
          after: after || undefined,
          require_nearby: nearbyAction
            ? [
              {
                action: nearbyAction,
                within: nearbyTolerance
              }
            ]
            : undefined,
        }
        : undefined,
    scoring: {
      limit: limit ? Number(limit) : undefined,
      min_score: minScore ? Number(minScore) : undefined
    },
  };


  return (
    <div>
      <h3>Dataset</h3>
      <select
        value={dataset}
        onChange={(e) => setDataset(e.target.value)}>
        {
          datasets.map(d => (
            <option
              key={d.id}
              value={d.id}
            >
              {d.name}
            </option>
          ))
        }
      </select>
      {
        datasets
          .filter(d => d.id === dataset)
          .map(d => (
            <div key={d.id}>
              <p>
                <b>Название:</b> {d.name}
              </p>
              <p>
                <b>Размер:</b> {d.size} событий
              </p>
              <p>
                <b>Период:</b> {d.period}
              </p>
              <p>
                <b>Описание:</b> {d.description}
              </p>
            </div>
          ))
      }
      <h3>Поиск</h3>
      <h4>Примерное время</h4>
      <input
        type="datetime-local"
        value={timeAround}
        onChange={(e) => setTimeAround(e.target.value)}
      />
      <input
        placeholder="Time tolerance"
        value={timeTolerance}
        onChange={(e) => setTimeTolerance(e.target.value)}
      /><br /><br />
      <h4>Nearby</h4>
      <input
        placeholder="Nearby action"
        value={nearbyAction}
        onChange={(e) => setNearbyAction(e.target.value)}
      />

      <input
        placeholder="Nearby tolerance"
        value={nearbyTolerance}
        onChange={(e) => setNearbyTolerance(e.target.value)}
      /><br /><br />
      <h4>Ограничение</h4>
      <input
        type="number"
        placeholder="Limit"
        value={limit}
        onChange={(e) => setLimit(e.target.value)}
      />
      <input
        type="number"
        placeholder="Min score"
        value={minScore}
        onChange={(e) => setMinScore(e.target.value)}
      /><br /><br />
      <h4>Контекст</h4>
      <input
        placeholder="Before"
        value={before}
        onChange={(e) => setBefore(e.target.value)}
      />
      <input
        placeholder="After"
        value={after}
        onChange={(e) => setAfter(e.target.value)}
      />

      <br /><br />
      <h4>Критерии</h4>
      <input
        placeholder="User ID"
        value={user}
        onChange={(e) => setUser(e.target.value)}
      />
      <input
        placeholder="File name"
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
      />
      <input
        placeholder="Action"
        value={action}
        onChange={(e) => setAction(e.target.value)}
      />

      <input
        placeholder="Destination type"
        value={destinationType}
        onChange={(e) => setDestinationType(e.target.value)}
      />
      <br /><br />
      <p><b>channel</b> - через какой канал произошло действие<br />
        <b>severity</b> - насколько событие подозрительное/опасное</p><br />
      <input
        placeholder="Channel"
        value={channel}
        onChange={(e) => setChannel(e.target.value)}
      />
      <input
        placeholder="Severity"
        value={severity}
        onChange={(e) => setSeverity(e.target.value)}
      />
      <br /><br />
      <button onClick={search}>
        Поиск
      </button>

      <h3>JSON Preview</h3>

      <pre>{JSON.stringify(requestPreview, null, 2)}</pre>

      <p>{message}</p>

      {
        events.map((item: any) => (
          <div key={item.event.event_id}>
            <hr />
            <p>Event ID: {item.event.event_id}</p>
            <p>Score: {item.score}</p>
            <p>Action: {item.event.action}</p>
            <p> Destination: {item.event.destination_type}</p>
            <p>Совпадения: {item.matched_hints.join(", ")}</p>

            <button onClick={() => {
              if (selectedEventId === item.event.event_id) {
                setSelectedEventId("");
                setContext(null);
              } else {
                setSelectedEventId(item.event.event_id);
                getContext(item.event.event_id);
              }
            }}>
              Подробнее
            </button>

            <button onClick={() => getExplain(item.event.event_id)}>
              Explain score
            </button>

            {
              selectedEventId === item.event.event_id && (
                <div>
                  <hr />
                  <h3>Контекст события</h3>
                  {
                    context && (
                      <div>
                        <h4>Текущее событие</h4>
                        <p>Event ID: {context.event?.event_id}</p>
                        <p>Timestamp: {context.event?.timestamp}</p>
                        <p>User: {context.event?.user_id}</p>
                        <p>Action: {context.event?.action}</p>
                        <p>File: {context.event?.file_name}</p>
                        <p>Destination type: {context.event?.destination_type}</p>

                        <hr />

                        <h4>До события</h4>

                        {
                          context.before && context.before.length > 0 ? (
                            context.before.map((event: any) => (
                              <div key={event.event_id}>
                                <p>Event ID: {event.event_id}</p>
                                <p>Timestamp: {event.timestamp}</p>
                                <p>Action: {event.action}</p>
                                <p>File: {event.file_name}</p>
                                <p>Destination: {event.destination_type}</p>

                                <hr />
                              </div>
                            ))
                          ) : (
                            <p>Нет событий до</p>
                          )
                        }

                        <h4>После события</h4>

                        {
                          context.after && context.after.length > 0 ? (
                            context.after.map((event: any) => (
                              <div key={event.event_id}>
                                <p>Event ID: {event.event_id}</p>
                                <p>Timestamp: {event.timestamp}</p>
                                <p>Action: {event.action}</p>
                                <p>File: {event.file_name}</p>
                                <p>Destination: {event.destination_type}</p>

                                <hr />
                              </div>
                            ))
                          ) : (
                            <p>Нет событий после</p>
                          )
                        }

                        <h4>JSON Context</h4>

                        <pre> {
                          JSON.stringify(
                            context,
                            null,
                            2
                          )
                        }
                        </pre>
                      </div>
                    )
                  }

                  {
                    !context && !loadingContext && (
                      <p>Контекст отсутствует</p>
                    )
                  }

                  <button onClick={() => {
                    setSelectedEventId("");
                    setContext(null);
                  }}>
                    Закрыть
                  </button>
                </div>
              )
            }

            {
              explain && selectedExplainId === item.event.event_id && (

                <div>
                  <hr />

                  <h3>Explain score</h3>

                  <p>Score: {explain.score}</p>

                  {
                    explain.contributions.map(
                      (c: any, index: number) => (
                        <div key={index}>
                          <hr />

                          <b>{c.hint}</b>
                          <p>Тип: {c.type}</p>
                          <p>Значение: {c.value}</p>
                          <p>Балл: +{c.points}</p>
                        </div>
                      )
                    )
                  }

                  <button onClick={() => {
                    setExplain(null);
                    setSelectedExplainId("");
                  }}>
                    Закрыть Explain
                  </button>
                </div>
              )
            }
          </div>
        ))
      }
    </div>
  );
}
export default App;