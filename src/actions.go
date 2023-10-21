package main

import ( 
    "log"
	"net/http"
    "time"
)

func change_priority(w http.ResponseWriter, r *http.Request)  {
    cookie, err := r.Cookie("auth");
    if err != nil {
        w.WriteHeader(401);
        return
    }

    data := r.URL.Query()
    if !data.Has("priority") || !data.Has("task") {
        w.WriteHeader(400);
        w.Write([]byte("Bad Request: not task key"));
        return
    }

    id := data.Get("task");
    priority := data.Get("priority");
    direction := data.Get("direction")
    if id == "" || priority == ""  || direction == "" {
        w.WriteHeader(400);
        w.Write([]byte("Bad Request: not task value"));
        return
    }

    if direction == "down" {
        res, err := db.Exec(`UPDATE task_priority SET priority = priority - 1 WHERE task_id IN (
                             SELECT id FROM todo WHERE user_id = 
                             (SELECT user_id FROM sessions WHERE token = ?)) 
                             AND priority <= ? AND priority >= (
                             SELECT priority FROM task_priority WHERE task_id = ?)`, cookie.Value, priority, id)
        if err != nil {
            unexpected_err(w, err, "ERROR: changing priority in db\n");
            return
        }

        tmp, err := res.RowsAffected();
        if err != nil {
            w.WriteHeader(500)
            return
        } else if tmp < 1 {
            w.WriteHeader(400)
            w.Write([]byte("Bad Request: first"));
            return
        }

        res, err = db.Exec(`UPDATE task_priority SET priority = ? WHERE task_id = ?;`, priority, id);
        if err != nil {
            unexpected_err(w, err, "ERROR: changing priority in db\n");
            return
        }

        tmp, err = res.RowsAffected();
        if err != nil {
            w.WriteHeader(500)
            return
        } else if tmp < 1 {
            w.WriteHeader(400)
            w.Write([]byte("Bad Request: second"));
            return
        }

        w.WriteHeader(200);
        return
    }


    res, err := db.Exec(`UPDATE task_priority SET priority = priority + 1 WHERE task_id IN (
                         SELECT id FROM todo WHERE user_id = (
                         SELECT user_id FROM sessions WHERE token = ?) AND priority >= ? AND priority < (
                         SELECT priority FROM task_priority WHERE task_id = ?))`, cookie.Value, priority, id)
    if err != nil {
        unexpected_err(w, err, "ERROR: changing priority in db\n");
        return
    }

    tmp, err := res.RowsAffected();
    if err != nil {
        w.WriteHeader(500)
        return
    } else if tmp < 1 {
        w.WriteHeader(400)
        w.Write([]byte("Bad Request: first"));
        return
    }

    res, err = db.Exec(`UPDATE task_priority SET priority = ? WHERE task_id = ?;`, priority, id);
    if err != nil {
        unexpected_err(w, err, "ERROR: changing priority in db\n");
        return
    }

    tmp, err = res.RowsAffected();
    if err != nil {
        w.WriteHeader(500)
        return
    } else if tmp < 1 {
        w.WriteHeader(400)
        w.Write([]byte("Bad Request: second"));
        return
    }

    w.WriteHeader(200);
    return
}

func delete_task(w http.ResponseWriter, r *http.Request)  {
    cookie, err := r.Cookie("auth");
    if err != nil {
        w.WriteHeader(401);
        return
    }

    data := r.URL.Query()
    if !data.Has("task") {
        w.WriteHeader(400);
        w.Write([]byte("Bad Request: not task key"));
        return
    }

    id := data.Get("task");
    if id == "" {
        w.WriteHeader(400);
        w.Write([]byte("Bad Request: not task value"));
        return
    }

    res, err := db.Exec(`UPDATE task_priority SET priority = priority - 1 
                        WHERE task_id IN ( 
                        SELECT id from todo WHERE user_id = (
                        SELECT user_id FROM sessions WHERE token = ?)) AND priority > (
                        SELECT priority FROM task_priority WHERE task_id = ?);`, cookie.Value, id); 
    if err != nil {
        unexpected_err(w, err,
        "ERROR: deleting task from user: %s and task: %s", cookie.Value, id)
        return
    } 

    ret, err := res.RowsAffected();
    if err != nil {
        unexpected_err(w, err,
        "ERROR: checking affeted rows from user: %s and task: %s", cookie.Value, id)
        return
    } else if ret < 1 {
        w.WriteHeader(400)
        w.Write([]byte("Task Not Found"))
        return
    }

    res, err = db.Exec(`DELETE FROM todo WHERE id = ? AND user_id = 
                        (SELECT user_id FROM sessions WHERE token = ?)`, 
                        id, cookie.Value); 
    if err != nil {
        unexpected_err(w, err,
        "ERROR: deleting task from user: %s and task: %s", cookie.Value, id)
        return
    } 

    ret, err = res.RowsAffected();
    if err != nil {
        unexpected_err(w, err,
        "ERROR: checking affeted rows from user: %s and task: %s", cookie.Value, id)
        return
    } else if ret < 1 {
        w.WriteHeader(400)
        w.Write([]byte("Task Not Found"))
        return
    }


    w.WriteHeader(200)
    w.Write([]byte("OK"))
}

func logout(w http.ResponseWriter, r *http.Request)  {
    log.Println("logout");
    cookie, err := r.Cookie("auth");
    if err != nil {
        http.Redirect(w, r, "/home", 301);
        if err != nil {
            println("ERROR: render errror.html");
            log.Println(err)
        }
        return
    }

    _, err = db.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value);
    if err != nil {
        unexpected_err(w, err, 
            "ERROR: querry user_id from token: %s", 
            cookie.Value);
        return
    }
    
    http.SetCookie(w, &http.Cookie{
        Name: "auth",
        Value: "deleted",
        Expires: time.Time{}.AddDate(1970, 01, 01),
    })
    http.Redirect(w, r, "/home", 301);
    return
}
