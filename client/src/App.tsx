import { Box, List, ThemeIcon, Text } from "@mantine/core";
import { CheckCircleFillIcon } from "@primer/octicons-react";
import useSWR from "swr";
import "./App.css";
import AddTodo from "./components/AddTodo";

export interface Todo {
  id: number;
  title: string;
  body: string;
  done: boolean;
}
// this is the ip and port for server
export const ENDPOINT = "http://localhost:4000";

const fetcher = (url: string) =>
  fetch(`${ENDPOINT}/${url}`).then((r) => r.json());

function App() {
  const { data, mutate } = useSWR<Todo[]>("api/todos", fetcher);

  async function markTodoAsDone(id: number) {
    const updated = await fetch(`${ENDPOINT}/api/todos/${id}/done`, {
      method: "PATCH",
    }).then((r) => r.json());

    mutate(updated);
  }

  async function toggleDone(id: number) {
    const updated = await fetch(`${ENDPOINT}/api/todos/${id}/toggle`, {
      method: "PATCH",
    }).then((r) => r.json());

    mutate(updated);
  }
  return (
    <Box
      sx={(theme) => ({
        padding: "2rem",
        width: "100%",
        maxWidth: "100rem",
        margin: "0 auto",
      })}
    >
      <List size="lg" fz="xl" mb={12} center >
        {data?.map((todo) => {
          return (
            <List.Item
              onClick={() => toggleDone(todo.id)}
              key={`todo_list__${todo.id}`}
              icon={
                todo.done ? (
                  <ThemeIcon color="teal" size={24} radius="xl">
                    <CheckCircleFillIcon size={20} />
                  </ThemeIcon>
                ) : (
                  <ThemeIcon color="gray" size={24} radius="xl">
                    <CheckCircleFillIcon size={20} />
                  </ThemeIcon>
                )
              }
              mb={10}
            >
              {todo.title}
              <List size="xs">
                  {todo.body}
              </List>
            </List.Item>
          );
        })}
      </List>

      <AddTodo mutate={mutate} />
    </Box>
  );
}

export default App;