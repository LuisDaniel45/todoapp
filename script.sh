#!/bin/sh

# cp database.db back.db

#change task 124 priority to 1 and increase priority of task where priority > 1 and user_id = 2 and priority < (task 124 priority)
cmd="sqlite3 back.db"

task_id=7;
task_priority=1;
user_id=1

# $cmd "UPDATE task_priority SET priority = priority + 1 WHERE task_id IN (
#      SELECT id FROM todo WHERE user_id = $user_id) AND priority >= $task_priority AND priority < (
#      SELECT priority FROM task_priority WHERE task_id = $task_id);
#      UPDATE task_priority SET priority = $task_priority WHERE task_id = $task_id;"
#
# $cmd "SELECT * FROM task_priority WHERE task_id IN (
#         SELECT id FROM todo WHERE user_id = $user_id);"
#
# $cmd "SELECT * FROM  task_priority WHERE task_id IN (
#      SELECT id FROM todo WHERE user_id = 2) AND priority > 4; ";
 
