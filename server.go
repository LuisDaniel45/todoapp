package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type err_values struct { 
    Code int 
    Msg string
}

var db *sql.DB;
var t *template.Template;

func main()  {
    db_init();
    template_init();
    assets_init();

    http.HandleFunc("/logout", logout);
    http.HandleFunc("/home", home);
    http.HandleFunc("/register", register);
    http.HandleFunc("/login", login);
    http.HandleFunc("/", root);

    err := http.ListenAndServe(":8080", nil);
    if err != nil {
        println("ERROR: opening port");
        log.Fatal(err)
    }
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

func home(w http.ResponseWriter, r *http.Request)  {
    err := t.ExecuteTemplate(w, "home.html", nil);
    if err != nil {
        println("ERROR: errror.html render")
        log.Println(err)
    }
}

func unexpected_err(w http.ResponseWriter, err error, msg string, args...any) {
    log.Printf(msg, args)
    log.Println(err)
    w.WriteHeader(501)
    err = t.ExecuteTemplate(w, "error.html", err_values{501, 
                                    "Internal Server Error"})
    if err != nil{
        println("ERROR: error.html render")
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

            _, err := db.Exec("INSERT INTO todo(user_id, task) values(?, ?)", user_id, todo);
            if err != nil {
                unexpected_err(w, err, 
                    "ERROR: inserting task to todo list for user_id: %d", 
                    user_id);
                return
            }
        }
    }

    rows, err = db.Query("SELECT task FROM todo WHERE user_id = ?", user_id)
    if err != nil {
        unexpected_err(w, err, 
            "ERROR: getting tasks for user_id: %d", 
            user_id);
        return
    }
    defer rows.Close()

    var todos []string;
    for rows.Next() {
        var task string;
        err = rows.Scan(&task);
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

func create_session(w http.ResponseWriter,  user_id int) (string, error) { 
    uuid, err := os.ReadFile("/proc/sys/kernel/random/uuid")
    if err != nil {
        unexpected_err(w, err, 
            "ERROR: generating uuid for session_id error")
        return "nil", err
    }
    session_id := string(uuid[:len(uuid)-1]);
    db.Exec("INSERT INTO sessions(user_id, token) values(?, ?)", 
            user_id, session_id);

    return session_id, nil;
}

func template_init() {
    var err error;
    t, err = template.ParseFiles(
        "views/index.html",
        "views/main.html",
        "views/error.html",
        "views/register.html",
        "views/login.html",
        "views/home.html",
    );
    if err != nil {
        println("ERROR: parsing html files")
        log.Fatal(err) 
    }
}

func db_init() {
    var err error;
    db, err = sql.Open("sqlite3", "./database.db")
    if err != nil {
        println("ERROR: opening database")
        log.Fatal(err) 
    }

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
    if err != nil {
        println("ERROR: creaing tables")
        db.Close()
        log.Fatal(err)
    }
}

func assets_init()  {
    ret, err := os.ReadDir("assets")
    if err != nil {
        log.Fatal(err)
    }

    for _, v := range(ret) { 
        if !v.IsDir() {
            name :=  v.Name();
            var type_header string;
            for i := len(name) - 1; i > 0; i-- {
                if name[i] == '.' {
                    switch name[i:] {
                        case ".js":
                            type_header = "application/javascript" 
                            break

                        case ".css":
                            type_header = "text/css" 
                            break
                    }
                    break
                }
            }

            name = "assets/" + name;
            content, err := os.ReadFile(name);
            if err != nil {
                log.Fatal(err)
            }

            http.HandleFunc("/" + name, func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", type_header);
                w.Write(content)
            })
        }
    }
}

