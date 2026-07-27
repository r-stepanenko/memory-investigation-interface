import { useEffect, useState } from "react";
import axios from "axios";
import {
  mockSearch,
  mockExplain,
  mockDatasets,
  mockContext
} from "./fixtures/mock";
import "./App.css"

function App() {
  const [user, setUser] = useState("");
  const [fileName, setFileName] = useState("");
  const [action, setAction] = useState("");
  const [destinationType, setDestinationType] = useState("");
  const [events, setEvents] = useState<any[]>([]);
  const [compareList, setCompareList] = useState<any[]>([]);
  const [message, setMessage] = useState("");
  const [datasets, setDatasets] = useState<any[]>([]);
  const [dataset, setDataset] = useState("");
  const [filters, setFilters] = useState<any>({
    user_id: [],
    file_name: [],
    action: [],
    destination_type: [],
    channel: [],
    severity: [],
  });
  const [searchId, setSearchId] = useState("");
  const [selectedEventId, setSelectedEventId] = useState("");
  const [context, setContext] = useState<any>(null);
  const [loadingContext, setLoadingContext] = useState(false);
  const [explain, setExplain] = useState<any>(null);
  const [selectedExplainId, setSelectedExplainId] = useState("");
  const [isMockMode, setIsMockMode] = useState(false);
  const [nearbyAction, setNearbyAction] = useState("");
  const [timeAround, setTimeAround] = useState("");
  const [timeTolerance, setTimeTolerance] = useState("");
  const [limit, setLimit] = useState("");
  const [minScore, setMinScore] = useState("");
  const [channel, setChannel] = useState("");
  const [severity, setSeverity] = useState("");
  const [before, setBefore] = useState("");
  const [after, setAfter] = useState("");
  const [resultMinScore, setResultMinScore] = useState("");
  const [resultAction, setResultAction] = useState("");
  const [resultUser, setResultUser] = useState("");
  const [resultDestination, setResultDestination] = useState("");
  const [sortBy, setSortBy] = useState("score");
  const [sortOrder, setSortOrder] = useState("desc");
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 10;

  useEffect(() => {
    async function loadDatasets() {
      try {
        const res = await axios.get(
          "http://localhost:8080/api/datasets"
        );
        console.log(res.data);
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

  useEffect(() => {
    async function loadDatasetEvents() {
      try {
        const res = await axios.get(
          `http://localhost:8080/api/datasets/${dataset}/filters`
        );
        setFilters(res.data);
      } catch (error) {
        console.log("Не удалось загрузить события Dataset");
      }
    }
    if (dataset) {
      loadDatasetEvents();
    }
  }, [dataset]);

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
    setCurrentPage(1);

    setMessage(
      "Backend недоступен. Используется mock-режим"
    );

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
                      action: nearbyAction
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
        console.log(candidates);
      }

    } catch (error: any) {
      console.log("SEARCH ERROR:", error.response?.data);
      console.log("STATUS:", error.response?.status);
      console.log("ERROR:", error.message);

      setIsMockMode(true);
      setSearchId("mock-search-1");
      setEvents(mockSearch.candidates);
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
      console.log("EXPLAIN RESPONSE:", res.data);
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
                action: nearbyAction
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

  function scoreClass(score: number) {
    if (score >= 80) return "high";
    if (score >= 50) return "medium";
    return "low";
  }

  const filteredEvents = events.filter((item: any) => {

    if (
      resultMinScore &&
      item.score < Number(resultMinScore)
    ) {
      return false;
    }

    if (
      resultAction &&
      item.event.action !== resultAction
    ) {
      return false;
    }

    if (
      resultUser &&
      item.event.user_id !== resultUser
    ) {
      return false;
    }

    if (
      resultDestination &&
      item.event.destination_type !== resultDestination
    ) {
      return false;
    }

    return true;
  });

  const sortedEvents = [...filteredEvents].sort((a, b) => {

    if (sortBy === "score") {
      return sortOrder === "desc"
        ? b.score - a.score
        : a.score - b.score;
    }

    const timeA = new Date(a.event.timestamp).getTime();
    const timeB = new Date(b.event.timestamp).getTime();

    return sortOrder === "desc"
      ? timeB - timeA
      : timeA - timeB;
  });

  const totalPages = Math.ceil(
    sortedEvents.length / pageSize
  );

  const paginatedEvents = sortedEvents.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  );

  const matchedCount =
    explain?.contributions?.filter(
      (c: any) => c.points > 0
    ).length ?? 0;

  const timeline = context
    ? [
      ...(context.before || []).map((e: any) => ({
        ...e,
        type: "before",
      })),

      {
        ...context.event,
        type: "current",
      },

      ...(context.after || []).map((e: any) => ({
        ...e,
        type: "after",
      })),
    ]
    : [];

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
      <h4>Nearby (поиск связанных событий)</h4>
      <input
        placeholder="Nearby action"
        value={nearbyAction}
        onChange={(e) => setNearbyAction(e.target.value)}
      /><br /><br />
      <h4>Критерии</h4>
      <input
        list="users"
        placeholder="User ID"
        value={user}
        onChange={(e) => setUser(e.target.value)}
      />

      <datalist id="users">
        {filters.user_id.map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>
      <input
        list="files"
        placeholder="File name"
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
      />

      <datalist id="files">
        {filters.file_name.map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>
      <input
        list="actions"
        placeholder="Action"
        value={action}
        onChange={(e) => setAction(e.target.value)}
      />

      <datalist id="actions">
        {filters.action.map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>
      <input
        list="destinations"
        placeholder="Destination type"
        value={destinationType}
        onChange={(e) => setDestinationType(e.target.value)}
      />

      <datalist id="destinations">
        {filters.destination_type.map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>
      <br /><br />
      <p><b>channel</b> - через какой канал произошло действие<br />
        <b>severity</b> - насколько событие подозрительное/опасное</p><br />
      <input
        list="channels"
        placeholder="Channel"
        value={channel}
        onChange={(e) => setChannel(e.target.value)}
      />

      <datalist id="channels">
        {filters.channel.map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>
      <input
        list="severities"
        placeholder="Severity"
        value={severity}
        onChange={(e) => setSeverity(e.target.value)}
      />

      <datalist id="severities">
        {filters.severity.map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>
      <br /><br />
      <button onClick={search}>
        Поиск
      </button>

      <h3>Быстрые фильтры результатов</h3>

      <input
        type="number"
        placeholder="Минимальный score"
        value={resultMinScore}
        onChange={(e) => setResultMinScore(e.target.value)}
      />

      <input
        list="result-actions"
        placeholder="Action"
        value={resultAction}
        onChange={(e) => setResultAction(e.target.value)}
      />

      <datalist id="result-actions">
        {Array.from(
          new Set(events.map(e => e.event.action))
        ).map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>

      <input
        list="result-users"
        placeholder="User"
        value={resultUser}
        onChange={(e) => setResultUser(e.target.value)}
      />

      <datalist id="result-users">
        {Array.from(
          new Set(events.map(e => e.event.user_id))
        ).map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>

      <input
        list="result-destinations"
        placeholder="Destination type"
        value={resultDestination}
        onChange={(e) => setResultDestination(e.target.value)}
      />

      <datalist id="result-destinations">
        {Array.from(
          new Set(events.map(e => e.event.destination_type))
        ).map((value: string) => (
          <option key={value} value={value} />
        ))}
      </datalist>

      <h3>Сортировка</h3>

      <select
        value={sortBy}
        onChange={(e) => setSortBy(e.target.value)}
      >
        <option value="score">По score</option>
        <option value="time">По времени</option>
      </select>

      <select
        value={sortOrder}
        onChange={(e) => setSortOrder(e.target.value)}
      >
        <option value="desc">По убыванию</option>
        <option value="asc">По возрастанию</option>
      </select>

      <h3>JSON Preview</h3>

      <pre>{JSON.stringify(requestPreview, null, 2)}</pre>

      <p>{message}</p>

      <h3>Сравнение</h3>

      {
        compareList.length === 0
          ? (
            <p>Нет выбранных кандидатов</p>
          )
          : (
            <table>
              <thead>
                <tr>
                  <th>Event ID</th>
                  <th>User</th>
                  <th>Action</th>
                  <th>Score</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {
                  compareList.map((c: any) => (
                    <tr key={c.event.event_id}>
                      <td>{c.event.event_id}</td>
                      <td>{c.event.user_id}</td>
                      <td>{c.event.action}</td>
                      <td>{c.score.toFixed(2)}</td>
                      <td>
                        <button
                          onClick={() =>
                            setCompareList(
                              compareList.filter(
                                item => item.event.event_id !== c.event.event_id
                              )
                            )
                          }
                        >
                          Удалить
                        </button>
                      </td>
                    </tr>
                  ))
                }
              </tbody>
            </table>
          )
      }
      <br />
      {
        totalPages > 1 && (
          <div>
            <button
              disabled={currentPage === 1}
              onClick={() =>
                setCurrentPage(currentPage - 1)
              }>
              Назад
            </button>
            <span>
              Страница {currentPage} из {totalPages}
            </span>

            <button
              disabled={currentPage === totalPages}
              onClick={() =>
                setCurrentPage(currentPage + 1)
              }>
              Вперёд
            </button>
          </div>
        )
      }

      {
        paginatedEvents.map((item: any) => (
          <div key={item.event.event_id}>
            <div className="candidate-card">
              <div className="card-header">
                <div>
                  <h3>{item.event.action}</h3>
                  <small>
                    {item.event.timestamp}
                  </small>
                </div>
                <div>
                  <div className={`score ${scoreClass(item.score)}`}>
                    {item.score.toFixed(2)}
                  </div>
                  <div className="score-bar">
                    <div
                      className="score-fill"
                      style={{ width: `${Number(item.score).toFixed(2)}%` }}
                    />
                  </div>
                </div>
              </div>
              <div className="info-grid">
                <div className="info-item">
                  <b>User</b><br />
                  {item.event.user_id}
                </div>

                <div className="info-item">
                  <b>File</b><br />
                  {item.event.file_name}
                </div>

                <div className="info-item">
                  <b>Destination</b><br />
                  {item.event.destination_type}
                </div>

                <div className="info-item">
                  <b>Event ID</b><br />
                  {item.event.event_id}
                </div>
              </div>
              <h4>Совпадения</h4>

              <div className="match-list">
                {
                  item.matched_hints.map((hint: string) => (
                    <div
                      key={hint}
                      className="match-tag">
                      {hint}
                    </div>
                  ))
                }
              </div>
              <p>
                <b>Summary:</b> Совпадение найдено по{" "}
                {item.matched_hints.length}{" "}
                {item.matched_hints.length === 1 ? "критерию" : "критериям"}.
              </p>

              <h4>Основные причины</h4>
              {
                item.contributions
                  .filter((c: any) => c.matched && c.points > 0)
                  .slice(0, 3)
                  .map(
                    (c: any, index: number) => (
                      <div key={index}>
                        <b>{index + 1}. {c.hint}</b>

                        <p>
                          {c.type}: {c.value}
                        </p>

                        <p>
                          +{c.points.toFixed(2)} баллов
                        </p>
                      </div>
                    )
                  )
              }

              <div className="card-buttons">
                <button
                  onClick={() => {
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
                <button
                  onClick={() =>
                    getExplain(item.event.event_id)
                  }>
                  Explain score
                </button>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(item.event.event_id);
                    alert("Event ID скопирован");
                  }}>
                  Копировать event_id
                </button>
                <button
                  onClick={() => {

                    if (
                      compareList.some(
                        e =>
                          e.event.event_id === item.event.event_id
                      )
                    ) return;

                    setCompareList([
                      ...compareList,
                      item,
                    ]);

                  }}>
                  Добавить в сравнение
                </button>
              </div>
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
                          <h4>Timeline</h4>

                          <div className="timeline">
                            {timeline.map((event: any) => (
                              <div
                                key={event.event_id}
                                className={`timeline-item ${event.type}`}>
                                <div>
                                  <b>{event.timestamp}</b>
                                </div>
                                <div>
                                  {event.action}
                                </div>
                                {
                                  event.type === "current" &&
                                  (
                                    <strong> ← кандидат</strong>
                                  )
                                }

                                {
                                  nearbyAction &&
                                  event.action === nearbyAction &&
                                  (
                                    <span className="nearby-badge">
                                      nearby
                                    </span>
                                  )
                                }
                              </div>
                            ))}
                          </div>
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
                    <button
                      onClick={() => {
                        const blob = new Blob(
                          [JSON.stringify(context, null, 2)],
                          { type: "application/json" }
                        );

                        const url = URL.createObjectURL(blob);

                        const a = document.createElement("a");
                        a.href = url;
                        a.download = `${item.event.event_id}.json`;
                        a.click();

                        URL.revokeObjectURL(url);
                      }}
                    >
                      Export JSON Context
                    </button>
                  </div>
                )
              }
              {
                explain && selectedExplainId === item.event.event_id && (
                  <div>
                    <hr />
                    <h3>Explain score</h3>
                    <h4>
                      Итоговый score:
                      {" "}
                      {explain.score.toFixed(2)}
                    </h4>
                    <p>Score рассчитан по вкладу каждого совпавшего признака.</p>
                    <h4>Вклад признаков</h4>
                    {
                      explain.contributions.map(
                        (c: any, index: number) => (

                          <div
                            key={index}
                            className={
                              c.type === "none"
                                ? "explain-missed"
                                : "explain-match"
                            }>
                            <b>{index + 1}. {c.hint} </b>
                            <p>
                              Тип совпадения:
                              {" "}
                              <strong>
                                {c.type}
                              </strong>
                            </p>
                            <p>
                              Значение события:
                              {" "}
                              {c.value}
                            </p>
                            <p>
                              Запрос:
                              {" "}
                              {c.query}
                            </p>
                            <p>
                              Вклад:
                              {" "}
                              <strong>
                                +{c.points.toFixed(2)}
                              </strong>
                            </p>
                          </div>
                        )
                      )
                    }
                    <h4>Объяснение</h4>
                    <p>Итоговый score получен сложением вкладов всех совпавших признаков.</p>
                    <p>
                      Совпало признаков: <b>{matchedCount}</b>.
                      Итоговый score: <b>{explain.score.toFixed(2)}</b>.
                    </p>
                    <button
                      onClick={() => {
                        setExplain(null);
                        setSelectedExplainId("");
                      }}>
                      Закрыть Explain
                    </button>
                  </div>
                )
              }
            </div>
          </div>
        ))
      }
    </div>
  );
}
export default App;