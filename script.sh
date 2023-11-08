cmd="sqlite3 database.db --box"
user_id="$1"
if [ "$user_id" = ""  ]; then
    user_id="1";
fi

$cmd "SELECT todo.id, task, priority FROM todo
      JOIN task_priority ON todo.id = task_priority.task_id
      WHERE user_id = $user_id ORDER BY priority;"

