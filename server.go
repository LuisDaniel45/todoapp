package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type err_values struct { 
    Code int 
    Msg string
}

func main()  {
    db, err := sql.Open("sqlite3", "./database.db")
    if err != nil {panic(err) }
    defer db.Close()

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users( 
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            username TEXT NOT NULL UNIQUE, 
            password TEXT NOT NULL 
        );

        CREATE TABLE IF NOT EXISTS sessions(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            user_id INTEGER NOT NULL,
            token TEXT NOT NULL UNIQUE,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );

        CREATE TABLE IF NOT EXISTS todo(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            user_id INTEGER NOT NULL,
            task TEXT,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );
    `);
    if err != nil {panic(err) }

    insert_users, err := db.Prepare("INSERT INTO users(username, password) values(?, ?)")
    if err != nil {panic(err) }
    defer insert_users.Close()

    t, err := template.ParseFiles(
        "index.html",
        "main.html",
        "error.html",
        "register.html",
        "login.html",
    );
    if err != nil {panic(err) }

    http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            r.ParseForm();
            if r.PostForm.Has("username") && 
               r.PostForm.Has("password") {
                   user := r.PostForm["username"][0];
                   password := r.PostForm["password"][0];
                   if user == ""  || password == "" {
                       t.ExecuteTemplate(w, "error.html", err_values{400,
                       "Bad Request, missing input"})
                       return
                   }

                   rows, err := db.Query("SELECT id FROM users WHERE username = ?", user);
                   if err != nil {
                       println("ERROR: querrying db for checking user registration")
                       log.Println(err)
                       t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                       return 
                   }
                   defer rows.Close()

                   if rows.Next() { 
                       t.ExecuteTemplate(w, "error.html", err_values{400,
                       "Bad Request, username already taken"})
                       return
                   }

                   res, err := insert_users.Exec(user, password);
                   if err != nil {
                       println("ERROR: inserting user and password to db")
                       log.Println(err)
                       t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                       return
                   }

                   user_id, err := res.LastInsertId();
                   if err != nil {
                       println("ERROR: getting last inseted user id from db")
                       log.Println(err)
                       t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                       return
                   }

                   uuid, err := os.ReadFile("/proc/sys/kernel/random/uuid")
                   if err != nil {
                       println("ERROR: generating uuid for session_id error")
                       log.Println(err)
                       t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                       return
                   }

                   session_id := string(uuid[:len(uuid)-1])
                   db.Exec(`INSERT INTO sessions(user_id, token) 
                            values(?,?)`, user_id, session_id)  
                   http.SetCookie(w, &http.Cookie{
                       Name: "auth", 
                       Value: session_id, 
                   });

                   http.Redirect(w, r, "/", 301);
                   return
            }
        }

        err = t.ExecuteTemplate(w, "register.html", nil);
        if err != nil {
            println("ERROR: register render");
            log.Println(err)
        }
    });

    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            r.ParseForm()
            if r.PostForm.Has("username") ||
               r.PostForm.Has("password") {
                   user := r.PostForm["username"][0];
                   password := r.PostForm["password"][0];
                   if user == "" || password == ""  {
                       err = t.ExecuteTemplate(w, "error.html", err_values{401,
                                        "Bad Request: missing username or password"})
                       if err != nil {
                           println("ERROR: render error.html for missing username or password");
                           println(err);
                       }
                       return

                   } 
                   
                   var user_id int;
                   db.QueryRow("SELECT id FROM users WHERE username = ? AND password = ?", 
                                user, password).Scan(&user_id);
                   if user_id == 0 { 
                       println(user_id)
                       err = t.ExecuteTemplate(w, "error.html", err_values{401,
                                        "Bad Request: invalid username or password"})
                       if err != nil {
                           println("ERROR: render error.html for invalid username or password");
                           println(err);
                       }
                       return
                   }

                   uuid, err := os.ReadFile("/proc/sys/kernel/random/uuid")
                   if err != nil {
                       println("ERROR: generating uuid for session_id error")
                       t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                       return
                   }

                   session_id := string(uuid[:len(uuid)-1])
                   db.Exec("INSERT INTO sessions(user_id, token) values(?, ?)", user_id, session_id);
                   http.SetCookie(w, &http.Cookie{
                       Name: "auth",
                       Value: session_id,
                   });

                   http.Redirect(w, r, "/", 301)
                   return
            }
        }

        err = t.ExecuteTemplate(w, "login.html", nil);
        if err != nil {
            println("ERROR: login render");
        }
    });

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            w.WriteHeader(404)
            err = t.ExecuteTemplate(w, "error.html", err_values{404, "Not Found"})
            if err != nil {
                println("ERROR: render error.html");
            }
            return;
        }

        cookie, err := r.Cookie("auth");
        if err != nil {
            w.WriteHeader(403)
            err = t.ExecuteTemplate(w, "error.html", err_values{403, "Forbidden Acces, Not Authorize"})
            if err != nil {
                println("ERROR: render errror.html");
                log.Println(err)
            }
            return
        }

        rows, err := db.Query("SELECT user_id FROM sessions WHERE token = ?", cookie.Value);
        if err != nil {
            println("ERROR: querry user_id from token: ", cookie.Value);
            err = t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
            log.Println(err)
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

        if r.Method == "POST" {
            r.ParseForm();
            if r.PostForm.Has("todo") {
                todo := r.PostForm["todo"][0]
                if todo == "" {
                    t.ExecuteTemplate(w, "error.html", err_values{400,
                                "Bad Request, missing input"})
                    return
                }

                _, err := db.Exec("INSERT INTO todo(user_id, task) values(?, ?)", user_id, todo); 
                if err != nil {
                    println("ERROR: inserting task to todo list for user_id: ", user_id);
                    log.Println(err)
                    err = t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                    if err != nil {
                        log.Println(err)
                    }
                }
            }
        }

        rows, err = db.Query("SELECT task FROM todo WHERE user_id = ?", user_id)
        if err != nil {
            println("ERROR: getting tasks for user_id:", user_id);
            log.Println(err)
            err = t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
            if err != nil {
                log.Println(err)
            }
        }
        defer rows.Close()

        var todos []string;
        for rows.Next() {
            var task string;
            err = rows.Scan(&task);
            if err != nil {
                println("ERROR: scanning tasks for user_id: ", user_id);
                log.Println(err)
                err = t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                if err != nil {
                    log.Println(err)
                }
            }
            todos = append(todos, task); // change to fix array 
        }

        err = t.ExecuteTemplate(w, "main.html", todos)
        if err != nil {
            println("ERROR: content render");
            log.Println(err)
        }
    });

    err = http.ListenAndServe(":8080", nil);
    if err != nil {
        println("ERROR: opening port");
        log.Println(err)
    }
}

