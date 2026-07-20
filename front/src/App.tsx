import { useState } from "react";
import axios from "axios";

function App() {
  const [user, setUser] = useState("");
  const [fileName, setFileName] = useState("");
  const [action, setAction] = useState("");
  const [destinationType, setDestinationType] = useState("");

  const [events, setEvents] = useState<any[]>([]);
  const [message, setMessage] = useState("");

  async function search() {
    console.log({
      user,
      fileName,
      action,
      destinationType
    });

    if (
      user === "" &&
      fileName === "" &&
      action === "" &&
      destinationType === ""
    ) {
      setEvents([]);
      setMessage("Введите данные для поиска");
      return;
    }

    setMessage("");
    setEvents([]);

    try {
      const res = await axios.post(
        "http://localhost:8080/api/search",
        {
          dataset_id: "control",
          hints: {
            user_id: user,
            file_name: fileName,
            action: action,
            destination_type: destinationType,
          },
        }
      );

      const candidates = res.data.candidates ?? [];

      if (candidates.length === 0) {
        setMessage("Ничего не найдено");
      } else {
        setEvents(candidates);
      }

    } catch (error: any) {
      console.log(error);
      if (error.response) {
        setMessage(
          "Ошибка сервера: " +
          (
            error.response.data?.error?.message ??
            "неизвестная ошибка"
          )
        );
      }
      else {
        setMessage("бэк недоступен!");
      }

    }
  }
  return (
    <div>
      <input
        value={user}
        onChange={(e) => setUser(e.target.value)}
        placeholder="User ID"
      />

      <input
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
        placeholder="File name"
      />

      <input
        value={action}
        onChange={(e) => setAction(e.target.value)}
        placeholder="Action"
      />

      <input
        value={destinationType}
        onChange={(e) => setDestinationType(e.target.value)}
        placeholder="Destination type"
      />

      <button onClick={search}>
        Поиск
      </button>

      {message && (
        <p>
          {message}
        </p>
      )}

      {events.map((item: any) => (
        <div key={item.event.event_id}>
          <p>
            Score: {item.score}
          </p>
          <p>
            Совпадения:
            {" "}
            {item.matched_hints.join(", ")}
          </p>
          <pre>
            {JSON.stringify(item.event, null, 2)}
          </pre>
        </div>
      ))}
    </div>
  );
}

export default App;