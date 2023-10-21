package main

import ( 
	"fmt"
    "log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request)  {
    err := t.ExecuteTemplate(w, "home.html", nil);
    if err != nil {
        println("ERROR: errror.html render")
        log.Println(err)
    }
}

func register(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            r.ParseForm();
            if r.PostForm.Has("username") && 
               r.PostForm.Has("password") && 
               r.PostForm.Has("confirm"){
                   user     := r.PostForm["username"][0];
                   password := r.PostForm["password"][0];
                   confirm  := r.PostForm["confirm"][0];

                   if user == ""  || password == "" || confirm == "" {
                       t.ExecuteTemplate(w, "error.html", err_values{400,
                                               "Bad Request, missing input"})
                       return
                   } else if password != confirm {
                       t.ExecuteTemplate(w, "error.html", err_values{400,
                                               "Bad Request, password and confirm do not match"})
                       return;
                   }

                   rows, err := db.Query("SELECT id FROM users WHERE username = ?", user);
                   if err != nil {
                       unexpected_err(w, err, 
                           "ERROR: querrying db for checking user registration")
                       return 
                   }
                   defer rows.Close()

                   if rows.Next() { 
                       t.ExecuteTemplate(w, "error.html", err_values{400,
                       "Bad Request, username already taken"})
                       return
                   }

                   res, err := db.Exec("INSERT INTO users(username, password) values(?, ?)", user, password)
                   if err != nil {
                       unexpected_err(w, err, 
                           "ERROR: inserting user and password to db")
                       return
                   }

                   user_id, err := res.LastInsertId();
                   if err != nil {
                       unexpected_err(w, err, 
                           "ERROR: getting last inseted user id from db")
                       return
                   }

                   session_id, err := create_session(w, int(user_id));
                   if err != nil {return }

                   http.SetCookie(w, &http.Cookie{
                       Name: "auth", 
                       Value: session_id, 
                   });

                   http.Redirect(w, r, "/", 301);
                   return
            }
        }

        err := t.ExecuteTemplate(w, "register.html", nil);
        if err != nil {
            println("ERROR: register.html render");
            log.Println(err)
        }
}

func login(w http.ResponseWriter, r *http.Request)  {
    if r.Method == "POST" {
        r.ParseForm()
        if r.PostForm.Has("username") ||
           r.PostForm.Has("password") {
            user := r.PostForm["username"][0];
            password := r.PostForm["password"][0];
            if user == "" || password == ""  {
                err := t.ExecuteTemplate(w, "error.html", err_values{401,
                            "Bad Request: missing username or password"})
                if err != nil {
                    println("ERROR: render error.html for missing username or password");
                    log.Println(err);
                }
                return

            } 

            var user_id int;
            db.QueryRow("SELECT id FROM users WHERE username = ? AND password = ?", 
            user, password).Scan(&user_id);
            if user_id == 0 { 
                println(user_id)
                err := t.ExecuteTemplate(w, "error.html", err_values{401,
                                    "Bad Request: invalid username or password"})
                if err != nil {
                    println("ERROR: render error.html for invalid username or password");
                    log.Println(err);
                }
                return
            }

            session_id, err := create_session(w, user_id);
            if err != nil {return}

            http.SetCookie(w, &http.Cookie{
                Name: "auth",
                Value: session_id,
            });

            http.Redirect(w, r, "/", 301)
            return
        }
    }

    err := t.ExecuteTemplate(w, "login.html", nil);
    if err != nil {
        println("ERROR: login render");
    }
}

func root(w http.ResponseWriter, r *http.Request)  {
    if r.URL.Path != "/" {
        w.WriteHeader(404)
        err := t.ExecuteTemplate(w, "error.html", err_values{404, "Not Found"})
        if err != nil {
            println("ERROR: render error.html");
        }
        return;
    }

    cookie, err := r.Cookie("auth");
    if err != nil {
        http.Redirect(w, r, "/home", 301);
        return
    }

    rows, err := db.Query("SELECT user_id FROM sessions WHERE token = ?", cookie.Value);
    if err != nil {
        unexpected_err(w, err, 
            "ERROR: querry user_id from token: %s", 
            cookie.Value);
        return
    }

    if !rows.Next(){
        w.WriteHeader(403)
        err = t.ExecuteTemplate(w, "error.html", err_values{403, "Forbidden Access, Not Autherize"})
        if err != nil {
            println("ERROR: render errror.html");
            log.Println(err)
        }
        rows.Close()
        return
    }
    var user_id int;
    rows.Scan(&user_id)
    rows.Close()

    // TODO: remember to send 400 if todo doesn't exist
    // also remember to change so respond with json 
    if r.Method == "POST" {
        r.ParseForm();
        if r.PostForm.Has("todo") {
            todo := r.PostForm["todo"][0]
            if todo == "" {
                w.WriteHeader(400)
                t.ExecuteTemplate(w, "error.html", err_values{400,
                            "Bad Request, missing input"})
                return
            }

            res, err := db.Exec("INSERT INTO todo(user_id, task) values(?, ?)", user_id, todo);
            if err != nil {
                unexpected_err(w, err, 
                    "ERROR: inserting task to todo list for user_id: %d", 
                    user_id);
                return
            }

            id, err := res.LastInsertId()
            if err != nil {
                unexpected_err(w, err, 
                    "ERROR: getting LastInsertId", 
                    user_id);
                return
            }

            // TODO: remember to change to 0 if no task 
            // TODO: remember to change to one querry 
            println(id)
            res, err = db.Exec("INSERT INTO task_priority(task_id, priority) values(?, (SELECT COALESCE(MAX(priority) + 1, 0) FROM task_priority WHERE task_id IN (SELECT id FROM todo WHERE user_id = ?)))",  id, user_id);
            if err != nil {
                unexpected_err(w, err, 
                    "ERROR: inserting task to todo list for user_id: %d", 
                    user_id);
                return
            }

            w.Write([]byte(fmt.Sprintf("%d", id)))
            return
        } 
        
        w.WriteHeader(400)
        w.Write([]byte("Bad Request"));
        return
    }

    rows, err = db.Query(`SELECT todo.task, todo.id
                          FROM todo FULL OUTER JOIN task_priority 
                          ON todo.id = task_priority.task_id WHERE todo.user_id = ?
                          ORDER BY task_priority.priority`, user_id)
    if err != nil {
        unexpected_err(w, err, 
            "ERROR: getting tasks for user_id: %d", 
            user_id);
        return
    }
    defer rows.Close()

    var todos []task_t;
    for rows.Next() {
        var task task_t;
        err = rows.Scan(&task.Task, &task.Id);
        if err != nil {
            unexpected_err(w, err, 
                "ERROR: scanning tasks for user_id: %d", 
                user_id);
            return
        }
        todos = append(todos, task); // change to fix array
    }

    err = t.ExecuteTemplate(w, "main.html", todos)
    if err != nil {
        println("ERROR: content render");
        log.Println(err)
    }
}
