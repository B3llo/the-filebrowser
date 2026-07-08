import { defineStore } from "pinia";
import { ref } from "vue";

interface ExtractTask {
  path: string;
  name: string;
  done: boolean;
}

export const useExtractStore = defineStore("extract", () => {
  const tasks = ref<ExtractTask[]>([]);

  const start = (path: string, name: string) => {
    tasks.value.push({ path, name, done: false });
  };

  const finish = (path: string) => {
    const task = tasks.value.find((t) => t.path === path && !t.done);
    if (!task) return;
    task.done = true;

    setTimeout(() => {
      tasks.value = tasks.value.filter((t) => t !== task);
    }, 2000);
  };

  const fail = (path: string) => {
    const index = tasks.value.findIndex((t) => t.path === path && !t.done);
    if (index !== -1) tasks.value.splice(index, 1);
  };

  return { tasks, start, finish, fail };
});
