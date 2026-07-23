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
  const [datasets, setDatasets] = useState<string[]>([]);
  const [dataset, setDataset] = useState("");
  const [searchId, setSearchId] = useState("");
  const [selectedEventId, setSelectedEventId] = useState("");
  const [context, setContext] = useState<any>(null);
  const [loadingContext, setLoadingContext] = useState(false);
  const [explain, setExplain] = useState<any>(null);
  const [selectedExplainId, setSelectedExplainId] = useState("");
  const [isMockMode, setIsMockMode] = useState(false);

  useEffect(() => {
    async function loadDatasets() {
      try {
        const res = await axios.get(
          "http://localhost:8080/api/datasets"
        );

        setDatasets(res.data);
        if (res.data.length > 0) {
          setDataset(res.data[0]);
        }
      } catch (error) {

        console.log("Backend недоступен, включен mock");

        setIsMockMode(true);
        setDatasets(mockDatasets);
        setDataset(mockDatasets[0]);
        setMessage("Backend недоступен. Используется mock-режим");
      }

    }
    loadDatasets();

  }, []);

  async function getContext(eventId: string) {
    setLoadingContext(true);
    try {
      if (isMockMode) {
        setContext(
          mockContext[eventId]
        );
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
    try {
      const res = await axios.post(
        "http://localhost:8080/api/search",
        {
          dataset_id: dataset,
          hints: {
            user_id: user,
            file_name: fileName,
            action: action,
            destination_type: destinationType
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

    hints: {
      user_id: user,
      file_name: fileName,
      action: action,
      destination_type: destinationType
    }
  };


  return (
    <div>
      <h3>Dataset</h3>
      <select
        value={dataset}
        onChange={(e) => setDataset(e.target.value)}>
        {
          datasets.map(d => (
            <option key={d}>
              {d}
            </option>
          ))
        }
      </select>

      <h3>Поиск</h3>

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