cp database.db database.bak.db
cmd="sqlite3 database.bak.db --box"
user_id="$1"
if [ "$user_id" = ""  ]; then
    user_id="1";
fi

id="15"
priority="1"

# $cmd "UPDATE task_priority SET
#       priority = CASE
#                     WHEN task_id = $id THEN  $priority
#                     ELSE priority - 1
#                 END
#       WHERE task_id IN (
#       SELECT id FROM todo WHERE user_id = $user_id)
#       AND priority <= $priority AND priority >= (
#       SELECT priority FROM task_priority WHERE task_id = $id)"
#
# $cmd "UPDATE task_priority SET
#       priority = CASE
#                       WHEN task_id = $id THEN $priority
#                       ELSE priority + 1
#                   END
#       WHERE task_id IN (
#       SELECT id FROM todo WHERE user_id = $user_id
#       AND priority >= $priority AND priority <= (
#       SELECT priority FROM task_priority WHERE task_id = $id))"

# $cmd "UPDATE task_priority SET priority = priority - 1 WHERE priority > 1 AND priority < 6"
# $cmd "UPDATE task_priority SET priority = $priority WHERE task_id = $id"

$cmd "SELECT todo.id, task, priority FROM todo
      JOIN task_priority ON todo.id = task_priority.task_id
      WHERE user_id = $user_id ORDER BY priority;"

